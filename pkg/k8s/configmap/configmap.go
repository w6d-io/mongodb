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
Created on 02/04/2021
*/
package configmap

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/w6d-io/mongodb/internal/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// CreateUpdate scripts configmap
func CreateUpdate(ctx context.Context, r client.Client, scheme *runtime.Scheme, mongoDB *db.MongoDB) error {
	log := util.GetLog(ctx, mongoDB).WithName("Create").WithName("configmap")
	var err error

	cm := &corev1.ConfigMap{}
	log.V(1).Info("create configmap")
	err = r.Get(ctx, getTypesNamespacedNameScript(mongoDB), cm)
	if err != nil && errors.IsNotFound(err) {
		cm = getScriptConfigMap(ctx, scheme, mongoDB)
		if cm == nil {
			log.Error(nil, "get configmap return nil")
			return &Error{Cause: nil, Detail: "get configmap return nil"}
		}
		err = r.Create(ctx, cm)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "fail to create configmap")
			return &Error{Cause: err, Detail: "fail to create configmap"}
		}
	} else if err != nil {
		log.Error(err, "fail to get configmap")
		return &Error{Cause: err, Detail: "failed to get configmap"}
	}
	err = r.Update(ctx, cm)
	if err != nil {
		log.Error(err, "fail to update configmap")
		return &Error{Cause: err, Detail: "fail to update configmap"}
	}
	return nil
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}
