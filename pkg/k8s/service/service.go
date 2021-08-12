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
Created on 11/04/2021
*/

package service

import (
	"context"
	"github.com/w6d-io/mongodb/internal/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Error struct {
	Cause  error
	Detail string
}

func Create(ctx context.Context, r client.Client, scheme *runtime.Scheme, mongoDB *db.MongoDB) error {
	log := util.GetLog(ctx, mongoDB).WithName("Create").WithName("Service")
	var err error

	log.V(1).Info("")
	svc := &corev1.Service{}
	err = r.Get(ctx, util.GetTypesNamespaceNamed(ctx, mongoDB), svc)
	if err != nil && errors.IsNotFound(err) {
		log.V(1).Info("create service")
		svc = getService(ctx, scheme, mongoDB)
		if svc == nil {
			log.Error(nil, "get service resource return nil")
			return &Error{Cause: nil, Detail: "get service return nil"}
		}
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "fail to create service")
			return &Error{Cause: err, Detail: "fail to create service"}
		}
	}
	return nil
}

func getService(ctx context.Context, scheme *runtime.Scheme, mongoDB *db.MongoDB) *corev1.Service {
	log := util.GetLog(ctx, mongoDB).WithName("GetService")
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongoDB.Name,
			Namespace: mongoDB.Namespace,
			Labels:    util.LabelsForMongoDB(mongoDB.Name),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "mongodb",
					Protocol: "TCP",
					Port:     db.MongoDBPort,
					TargetPort: intstr.IntOrString{
						IntVal: db.MongoDBPort,
					},
				},
			},
			Selector: util.LabelsForMongoDB(mongoDB.Name),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	if err := ctrl.SetControllerReference(mongoDB, svc, scheme); err != nil {
		log.Error(err, "set owner failed")
		return nil
	}
	return svc
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}
