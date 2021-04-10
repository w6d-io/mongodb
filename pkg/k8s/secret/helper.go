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
Created on 31/03/2021
*/
package secret

import (
	"context"
	"encoding/base64"
	db "github.com/w6d-io/mongodb/api/v1alpha1"
	"github.com/w6d-io/mongodb/internal/config"
	"github.com/w6d-io/mongodb/internal/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetContentFromKeySelector Get Secret content and decode
func GetContentFromKeySelector(ctx context.Context, r client.Client, c *corev1.SecretKeySelector) string {
	correlationID := ctx.Value("correlation_id")
	log := ctrl.Log.WithValues("correlation_id", correlationID, "name", c.Name, "key", c.Key)

	if r == nil || c == nil {
		log.V(1).Info("k8s client or configmap key is nil")
		return ""
	}
	secret := &corev1.Secret{}
	o := util.GetTypesNamespacedNameFromString(c.Name,config.GetNamespace())
	log.V(1).Info("get types namespaced Name", "object", o)
	err := r.Get(ctx, o, secret)
	if err != nil {
		log.Error(err, "get secret")
		return ""
	}
	encryptedContent, ok := secret.Data[c.Key]
	if !ok {
		log.Error(nil, "no such element in configmap")
		return ""
	}
	content, err := base64.StdEncoding.DecodeString(string(encryptedContent))
	if err != nil {
		log.V(1).Error(err, "decode secret failed")
		return ""
	}
	return string(content)
}

// GetContentFromKey get secret content from key
func GetContentFromKey(ctx context.Context, r client.Client, name, key string) string {

	c := &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: name,
		},
		Key: key,
	}
	return GetContentFromKeySelector(ctx, r, c)
}

func IsKeyExist(ctx context.Context, r client.Client, c *corev1.SecretKeySelector) bool {
	correlationID := ctx.Value("correlation_id")
	if r == nil || c == nil {
		return false
	}
	log := ctrl.Log.WithValues("correlation_id", correlationID, "name", c.Name, "key", c.Key)
	secret := &corev1.Secret{}

	o := util.GetTypesNamespacedNameFromString(c.Name,config.GetNamespace())
	log.V(1).Info("get types namespaced Name", "object", o)
	err := r.Get(ctx, o, secret)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "get secret")
		return false
	} else if err != nil {
		log.Error(err, "get secret")
		return false
	}
	_, ok := secret.Data[c.Key]
	return ok
}

func getRootSecret(ctx context.Context, scheme *runtime.Scheme, mongoDB *db.MongoDB) *corev1.Secret {
	log := util.GetLog(ctx, mongoDB).WithName("GetRootSecret")

	passwd := util.GeneratePassword(30, 3, 3, 2)
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongoDB.Name,
			Namespace: mongoDB.Namespace,
			Labels:    util.LabelsForMongoDB(mongoDB.Name),
		},
		StringData: map[string]string{
			MongoRootPasswordKey: passwd,
		},
	}
	if err := ctrl.SetControllerReference(mongoDB, sec, scheme); err != nil {
		log.Error(err, "set owner failed")
		return nil
	}
	return sec
}
