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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
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
	o := client.ObjectKey{Name: c.Name, Namespace: config.GetNamespace()}
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
		log.Error(err, "decode secret failed")
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
	log := ctrl.Log.WithValues("correlation_id", correlationID, "name", c.Name, "key", c.Key)
	if r == nil || c == nil {
		log.V(1).Info("k8s client or configmap key is nil")
		return false
	}
	secret := &corev1.Secret{}
	o := client.ObjectKey{Name: c.Name, Namespace: config.GetNamespace()}
	err := r.Get(ctx, o, secret)
	if err != nil {
		log.Error(err, "get secret")
		return false
	}
	_, ok := secret.Data[c.Key]
	return ok
}

func getRootSecret(mongoDB *db.MongoDB) *corev1.Secret {
	passwd := util.GeneratePassword(30, 3, 3, 2)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongoDB.Name,
			Namespace: mongoDB.Namespace,
			Labels:    util.LabelsForMongoDB(mongoDB.Name),
		},
		StringData: map[string]string{
			MongoRootPasswordKey: passwd,
		},
	}
}
