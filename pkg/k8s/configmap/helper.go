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
	"github.com/w6d-io/mongodb/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func GetVolume(name string, reference corev1.LocalObjectReference) corev1.Volume {
	var mode int32 = 0755
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: reference.Name,
				},
				DefaultMode: &mode,
			},
		},
	}
}

func getScriptConfigMap(mongoDB *db.MongoDB) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongoDB.Name + "-scripts",
			Namespace: mongoDB.Namespace,
			Labels:    util.LabelsForMongoDB(mongoDB.Name),
		},
		Data: map[string]string{
			"auto-discovery.sh": AutoDiscovery,
			"setup.sh":          Setup,
			"setup-hidden.sh":   SetupHidden,
		},
	}
}
