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
	"github.com/google/uuid"
	"github.com/w6d-io/mongodb/pkg/controllers/mongodb"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
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
		err = mongodb.CreateUpdate(ctx, r.Client, mdb)
		if err != nil {
			log.Error(err, "failed to create resources")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "failed to get StatefulSet")
		return ctrl.Result{}, err
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
