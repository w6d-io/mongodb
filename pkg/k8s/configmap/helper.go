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
	"fmt"
	"github.com/w6d-io/mongodb/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

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

func getScriptConfigMap(ctx context.Context, scheme *runtime.Scheme, mongoDB *db.MongoDB) *corev1.ConfigMap {
	log := util.GetLog(ctx, mongoDB)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongoDB.Name + "-scripts",
			Namespace: mongoDB.Namespace,
			Labels:    util.LabelsForMongoDB(mongoDB.Name),
		},
		Data: map[string]string{
			"auto-discovery.sh": fmt.Sprintf(AutoDiscovery, mongoDB.Namespace, mongoDB.Namespace),
			"setup.sh":          fmt.Sprintf(Setup, getFullname(mongoDB)),
			"setup-hidden.sh":   SetupHidden,
			"replicaset-entrypoint.sh": ReplicasetEntrypoint,
		},
	}
	if err := ctrl.SetControllerReference(mongoDB, cm, scheme); err != nil {
		log.Error(err, "set owner failed")
		return nil
	}
	return cm
}
func getReplicaSetConfigMap(ctx context.Context, scheme *runtime.Scheme, mongoDB *db.MongoDB) *corev1.ConfigMap {
	log := util.GetLog(ctx, mongoDB)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongoDB.Name + "-replicaset-entrypoint",
			Namespace: mongoDB.Namespace,
			Labels:    util.LabelsForMongoDB(mongoDB.Name),
		},
		Data: map[string]string{
			"replicaset-entrypoint.sh": ReplicasetEntrypoint,
		},
	}
	if err := ctrl.SetControllerReference(mongoDB, cm, scheme); err != nil {
		log.Error(err, "set owner failed")
		return nil
	}
	return cm
}
func getFullname(mongoDB *db.MongoDB) string {
	return fmt.Sprintf("%s-0", mongoDB.Name)
}

func getTypesNamespacedNameScript(mongoDB *db.MongoDB) types.NamespacedName {
	return types.NamespacedName{Name: mongoDB.Name + "-scripts", Namespace: mongoDB.Namespace}
}

func getTypesNamespacedNameReplicaset(mongoDB *db.MongoDB) types.NamespacedName {
	return types.NamespacedName{Name: mongoDB.Name + "-replicaset-entrypoint", Namespace: mongoDB.Namespace}
}
