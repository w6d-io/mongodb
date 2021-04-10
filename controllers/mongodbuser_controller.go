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
	"github.com/w6d-io/mongodb/pkg/controllers/mongodb/user"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// MongoDBUserReconciler reconciles a MongoDBUser object
type MongoDBUserReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=db.w6d.io,resources=mongodbusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.w6d.io,resources=mongodbusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.w6d.io,resources=mongodbusers/finalizers,verbs=update

func (r *MongoDBUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	correlationID := uuid.New().String()
	ctx = context.WithValue(context.Background(), "correlation_id", correlationID)
	logger := r.Log.WithValues("user", req.NamespacedName, "correlation_id", correlationID)
	log := logger.WithName("Reconcile")
	var err error

	usr := &db.MongoDBUser{}
	if err = r.Get(ctx, req.NamespacedName, usr); err != nil {
		if errors.IsNotFound(err) {
			log.Info("User ")
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get MongoDB User")
		return ctrl.Result{}, err
	}

	if usr.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(usr, FinalizerName) {
			if err = user.Delete(ctx, r.Client, usr); err != nil && !errors.IsNotFound(err) {
				log.Error(err, "delete MongoDB user failed")
				return ctrl.Result{}, err
			}
		}

		controllerutil.RemoveFinalizer(usr, FinalizerName)
		if err = r.Update(ctx, usr); err != nil {
			log.Error(err, "remove finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(usr, FinalizerName) {
		controllerutil.AddFinalizer(usr, FinalizerName)
		if err = r.Update(ctx, usr); err != nil {
			log.Error(err, "add finalizer")
			return ctrl.Result{}, err
		}
	}

	if err = user.Create(ctx, r.Client, usr); err != nil {
		// TODO: if err returned is a non exist maybe return nil
		log.Error(err, "create MongoDB user")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&db.MongoDBUser{}).
		Complete(r)
}
