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
package secret

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/w6d-io/mongodb/internal/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func Create(ctx context.Context, r client.Client, scheme *runtime.Scheme, mongoDB *db.MongoDB) error {
	log := util.GetLog(ctx, mongoDB).WithName("Create").WithName("Secret")
	var err error

	sec := &corev1.Secret{}
	log.V(1).Info("")
	err = r.Get(ctx, util.GetTypesNamespaceNamed(ctx, mongoDB), sec)
	if err != nil && errors.IsNotFound(err) {
		log.V(1).Info("create secret")
		sec = getRootSecret(ctx, scheme, mongoDB)
		if sec == nil {
			log.Error(nil, "get secret resource return nil")
			return &Error{Cause: nil, Detail: "get secret return nil"}
		}
		err = r.Create(ctx, sec)
		if err != nil {
			log.Error(err, "fail to create secret")
			return &Error{Cause: err, Detail: "fail to  create secret"}
		}
		log.V(1).Info("secret created")
	} else if err != nil {
		log.Error(err, "fail to get secret")
		return &Error{Cause: err, Detail: "fail to get secret"}
	}
	return nil
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}
