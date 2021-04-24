/*
Copyright 2021 WILDCARD SA.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/controllers/mongodb"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// MongoDBReconciler reconciles a MongoDB object
type MongoDBReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=db.w6d.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.w6d.io,resources=mongodbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.w6d.io,resources=mongodbs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	correlationID := uuid.New().String()
	ctx = context.WithValue(context.Background(), "correlation_id", correlationID)
	logger := r.Log.WithValues("mongodb", req.NamespacedName, "correlation_id", correlationID)
	log := logger.WithName("Reconcile")
	var err error
	// your logic here
	mdb := &db.MongoDB{}
	if err = r.Get(ctx, req.NamespacedName, mdb); err != nil {
		if errors.IsNotFound(err) {
			log.Info("MongoDB resource not found. Ignore since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get MongoDB")
		return ctrl.Result{}, err
	}

	sts := &appsv1.StatefulSet{}
	err = r.Get(ctx, req.NamespacedName, sts)
	if err != nil && errors.IsNotFound(err) {
		err = mongodb.CreateUpdate(ctx, r.Client, r.Scheme, mdb)
		if err != nil {
			log.Error(err, "failed to create resources")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "failed to get StatefulSet")
		return ctrl.Result{}, err
	}
	log.V(1).Info("update sts")
	if err = r.updateSTS(ctx, req, mdb); err != nil {
		log.Error(err, "update sts failed")
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}
	log.V(1).Info("update status")
	if err = r.UpdateStatus(ctx, mdb, sts); err != nil {
		log.Error(err, "update status failed")
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&db.MongoDB{}).
		Owns(&appsv1.StatefulSet{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Complete(r)
}

func (r *MongoDBReconciler) updateSTS(ctx context.Context, req ctrl.Request, mongoDB *db.MongoDB) error {
	log := util.GetLog(ctx, mongoDB)
	var err error
	sts := &appsv1.StatefulSet{}
	err = r.Get(ctx, req.NamespacedName, sts)
	if err != nil {
		return err
	}
	if *sts.Spec.Replicas != *mongoDB.Spec.Replicas {
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			sts := &appsv1.StatefulSet{}
			err = r.Get(ctx, req.NamespacedName, sts)
			if err != nil {
				return client.IgnoreNotFound(err)
			}
			sts.Spec.Replicas = mongoDB.Spec.Replicas
			if err := r.Update(ctx, sts); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Error(err, "update sts")
			return err
		}
	}
	return nil
}

func (r *MongoDBReconciler) UpdateStatus(ctx context.Context, mdb *db.MongoDB, sts *appsv1.StatefulSet) error {
	log := util.GetLog(ctx, mdb)
	var err error
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		mdb.Status.Phase, err = r.GetMongoDBStatus(ctx, mdb, sts)
		if err != nil {
			log.Error(err, "get mongodb status failed")
			return err
		}
		if err := r.Status().Update(ctx, mdb); err != nil {
			log.Error(err, "unable to update MongoDB status")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MongoDBReconciler) GetMongoDBStatus(ctx context.Context, mdb *db.MongoDB, sts *appsv1.StatefulSet) (db.MongoDBPhase, error) {
	var indexes []int
	if mdb.Spec.Replicas == nil || *mdb.Spec.Replicas == 0 {
		return db.MongoDBPhasePaused, nil
	}
	for p := 0; p < int(sts.Status.Replicas); p++ {
		i, err := r.GetPodStatus(ctx, mdb, p)
		if err != nil {
			return "", err
		}
		indexes = append(indexes, i)
	}
	var max int
	for _, index := range indexes {
		if index > max {
			max = index
		}
	}
	return status[max], nil
}

func (r *MongoDBReconciler) GetPodStatus(ctx context.Context, mdb *db.MongoDB, index int) (int, error) {
	log := util.GetLog(ctx, mdb)
	var err error
	po := &corev1.Pod{}
	name := fmt.Sprintf("%s-%d", mdb.Name, index)
	nn := types.NamespacedName{Name: name, Namespace: mdb.Namespace}
	err = r.Get(ctx, nn, po)
	if err != nil {
		log.Error(err, "get pod failed")
		return 0, err
	}
	return GetStatusIndex(po.Status.ContainerStatuses), nil
}

// GetStatusIndex returns a kubernetes State
func GetStatusIndex(cs []corev1.ContainerStatus) int {
	if len(cs) == 0 {
		return 2
	}
	for _, c := range cs {
		if c.Name != "mongodb" {
			continue
		}
		if c.Ready && c.State.Running != nil {
			return 0
		}
		if c.State.Waiting != nil && c.State.Waiting.Reason == "ContainerCreating" {
			return 1
		}
		if c.State.Waiting != nil && c.State.Waiting.Reason == "CrashLoopBackOff" {
			return 3
		}
	}
	return 2
}

var status = map[int]db.MongoDBPhase{
	-1: db.MongoDBPhasePaused,
	0:  db.MongoDBPhaseReady,
	1:  db.MongoDBPhaseNotReady,
	2:  db.MongoDBPhaseProvisioning,
	3:  db.MongoDBPhaseCritical,
}
