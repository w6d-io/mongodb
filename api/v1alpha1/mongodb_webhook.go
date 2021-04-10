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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var mongodblog = logf.Log.WithName("mongodb-resource")

func (in *MongoDB) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-db-w6d-io-v1alpha1-mongodb,mutating=true,failurePolicy=fail,admissionReviewVersions=v1;v1beta1,sideEffects=None,groups=db.w6d.io,resources=mongodbs,verbs=create;update,versions=v1alpha1,name=mutate.mongodb.db.w6d.io

var _ webhook.Defaulter = &MongoDB{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MongoDB) Default() {
	mongodblog.Info("default", "name", in.Name)
	if in.Spec.Replicas == nil {
		in.Spec.Replicas = new(int32)
		*in.Spec.Replicas = 1
	}
	if in.Spec.Version == "" {
		in.Spec.Version = "4.4"
	}
}

// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-db-w6d-io-v1alpha1-mongodb,mutating=false,failurePolicy=fail,admissionReviewVersions=v1;v1beta1,sideEffects=None,groups=db.w6d.io,resources=mongodbs,versions=v1alpha1,name=validate.mongodb.db.w6d.io

var _ webhook.Validator = &MongoDB{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDB) ValidateCreate() error {
	mongodblog.Info("validate create", "name", in.Name)
	return DBCreate(in)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDB) ValidateUpdate(old runtime.Object) error {
	mongodblog.Info("validate update", "name", in.Name)

	return DBUpdate(old.(*MongoDB), in)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDB) ValidateDelete() error {
	mongodblog.Info("validate delete", "name", in.Name)

	return nil
}
