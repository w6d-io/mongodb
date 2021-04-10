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
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
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

	usr := db.MongoDBUser{}
	if err = r.Get(ctx, req.NamespacedName, usr); err != nil {
		if errors.IsNotFound(err) {
			log.Info("User ")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&db.MongoDBUser{}).
		Complete(r)
}
