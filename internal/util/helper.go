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
package util

import (
	"context"
	"crypto/sha256"
	"fmt"
	"k8s.io/client-go/tools/cache"
	"math/rand"
	"strings"

	"github.com/go-logr/logr"
	"github.com/w6d-io/mongodb/internal/config"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	k8sv1alpha1 "github.com/w6d-io/mongodb/apis/k8s/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// AsSha256 return the sha256 hash from any
func AsSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// GetSecretKeySelector return  secret info from name and key
func GetSecretKeySelector(name, key string) *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: name,
		},
		Key: key,
	}
}

// GetConfigMapKeySelector return  secret info from name and key
func GetConfigMapKeySelector(name, key string) *corev1.ConfigMapKeySelector {
	return &corev1.ConfigMapKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: name,
		},
		Key: key,
	}
}

// GetLocalObjectReference return LocalObjectReference
func GetLocalObjectReference(name string) corev1.LocalObjectReference {
	return corev1.LocalObjectReference{
		Name: name,
	}
}

// EscapePassword return the password by replacing `@` and `:` with html encoding
func EscapePassword(password string) string {
	password = strings.Replace(password, "@", "%40", -1)
	password = strings.Replace(password, ":", "%3A", -1)
	return password
}

func GetNodeSelector(podTemplate *k8sv1alpha1.PodTemplate) map[string]string {
	ns := config.GetNodeSelector()
	if podTemplate != nil && len(podTemplate.NodeSelector) != 0 {
		ns = podTemplate.NodeSelector
	}
	return ns
}

func GetServiceAccount(podTemplate *k8sv1alpha1.PodTemplate) string {
	sa := config.GetServiceAccountName()
	if podTemplate != nil && podTemplate.ServiceAccountName != "" {
		sa = podTemplate.ServiceAccountName
	}
	return sa
}
func GetSecurityContext(podTemplate *k8sv1alpha1.PodTemplate) *corev1.PodSecurityContext {
	sc := config.GetSecurityContext()
	if podTemplate != nil && podTemplate.SecurityContext != nil {
		sc = podTemplate.SecurityContext
	}
	return sc
}

func GetAffinity(podTemplate *k8sv1alpha1.PodTemplate) *corev1.Affinity {
	af := config.GetAffinity()
	if podTemplate != nil && podTemplate.Affinity != nil {
		af = podTemplate.Affinity
	}
	return af
}

func GetTolerations(podTemplate *k8sv1alpha1.PodTemplate) []corev1.Toleration {
	to := config.GetTolerations()
	if podTemplate != nil && len(podTemplate.Tolerations) != 0 {
		to = podTemplate.Tolerations
	}
	return to
}

func LabelsForMongoDB(name string) map[string]string {
	return map[string]string{
		"db.w6d.io/component": "mongodb",
		"db.w6d.io/release":   name,
	}
}

func GetTypesNamespaceNamed(ctx context.Context, object runtime.Object) types.NamespacedName {
	o, err := meta.Accessor(object)
	if err != nil {
		ctrl.Log.Error(err, "failed to implement accessor", "correlation_id", ctx.Value("correlation_id"))
		return types.NamespacedName{}
	}
	return types.NamespacedName{Name: o.GetName(), Namespace: o.GetNamespace()}
}

func GetTypesNamespacedNameFromString(namespacedName string, defaultNamespace string) types.NamespacedName {
	name, namespace, err := cache.SplitMetaNamespaceKey(namespacedName)
	if err != nil {
		ctrl.Log.Error(err, "split namespacedName failed")
		return types.NamespacedName{Name: namespacedName, Namespace: defaultNamespace}
	}
	if namespace == "" {
		namespace = defaultNamespace
	}
	return types.NamespacedName{Name: name, Namespace: namespace}
}

func GetLog(ctx context.Context, obj runtime.Object) logr.Logger {
	correlationID := ctx.Value("correlation_id")
	nn := GetTypesNamespaceNamed(ctx, obj)
	return ctrl.Log.WithValues("correlation_id", correlationID, "object", nn.String())
}

func GeneratePassword(passwordLength, minSpecialChar, minNum, minUpperCase int) string {
	var password strings.Builder

	//Set special character
	for i := 0; i < minSpecialChar; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	//Set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	//Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

// StringInArray ...
func StringInArray(needle string, haystack []string) bool {
	for _, elem := range haystack {
		if elem == needle {
			return true
		}
	}
	return false
}

// AccessModeIn ...
func AccessModeIn(needle corev1.PersistentVolumeAccessMode, slice []corev1.PersistentVolumeAccessMode) bool {
	for _, elem := range slice {
		if elem == needle {
			return true
		}
	}
	return false
}
