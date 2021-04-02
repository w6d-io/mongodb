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
Created on 29/03/2021
*/
package statefulset

import (
	"context"
	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/k8s/secret"

	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
)

func CreateUpdate(ctx context.Context, r client.Client, mongoDB *db.MongoDB) error {
	log := util.GetLog(ctx, mongoDB)
	if !secret.IsKeyExist(ctx, r, util.GetSecretKeySelector(mongoDB.Name, MongoRootPasswordKey)) {
		log.Error(nil, "Secret mongodb-root-password key does not exists")
		return &Error{
			Cause:  nil,
			Detail: "Secret mongodb-root-password key does not exists",
		}
	}
	sts := getStatefulSetMongoDB(ctx, r, mongoDB)
	log.V(1).Info("create statefulSet")
	err := r.Create(ctx, sts)
	if err != nil {
		log.Error(err, "create statefulSet failed")
		return &Error{
			Cause:  err,
			Detail: "create statefulSet failed",
		}
	}
	return nil
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}
