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
package mongodb

import (
	"context"

	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/k8s/secret"
	"github.com/w6d-io/mongodb/pkg/k8s/service"
	"github.com/w6d-io/mongodb/pkg/k8s/statefulset"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
)

func CreateUpdate(ctx context.Context, r client.Client, scheme *runtime.Scheme, mongoDB *db.MongoDB) error {
	var err error
	log := util.GetLog(ctx, mongoDB)
	//err = configmap.CreateUpdate(ctx, r, scheme, mongoDB)
	//if err != nil {
	//	log.Error(err, "configmap processing failed")
	//	return err
	//}
	//record.EventRecorder(mongoDB, corev1.EventTypeNormal, "Creating", "creating secret in progress")
	err = secret.Create(ctx, r, scheme, mongoDB)
	if err != nil {
		log.Error(err, "secret processing failed")
		return err
	}
	err = statefulset.CreateUpdate(ctx, r, scheme, mongoDB)
	if err != nil {
		log.Error(err, "statefulSet processing failed")
		return err
	}
	err = service.Create(ctx, r, scheme, mongoDB)
	if err != nil {
		log.Error(err, "service processing failed")
		return err
	}
	return nil
}
