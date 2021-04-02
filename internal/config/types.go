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
package config

import corev1 "k8s.io/api/core/v1"

// Config for the controller
type Config struct {

	// Namespace where controller running
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Images map of image to use for build resource (sts, jobs)
	Images map[string]Image `json:"images,omitempty" yaml:"images,omitempty"`

	// ServiceAccount for the Pod
	ServiceAccount corev1.LocalObjectReference `json:"serviceAccount,omitempty" yaml:"serviceAccount,omitempty"`

	// SecurityContext is the default securityContext for the mongodb container
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty" yaml:"securityContext,omitempty"`

	// Affinity to applied to pods
	Affinity *corev1.Affinity `json:"affinity,omitempty" yaml:"affinity,omitempty"`

	// NodeSelector to applied to pods
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`

	// Tolerations to set for pods
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
}

type Image string

var (
	config = &Config{}
)
