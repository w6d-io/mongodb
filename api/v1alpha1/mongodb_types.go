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
*/

package v1alpha1

import (
	k8sv1alpha1 "github.com/w6d-io/mongodb/apis/k8s/v1alpha1"
	k8sdbv1alpha1 "github.com/w6d-io/mongodb/apis/k8sdb/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MongoDBSpec defines the desired state of MongoDB
type MongoDBSpec struct {
	// Version of MongoDB
	Version string `json:"version,omitempty"`

	// Replicas number of instance
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Storage spec for persistence
	Storage corev1.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// AuthSecret contains database secret credential
	AuthSecret *corev1.LocalObjectReference `json:"authSecret,omitempty"`

	// PodTemplate is a configuration for pod
	// +optional
	PodTemplate *k8sv1alpha1.PodTemplate `json:"podTemplate,omitempty"`

	// TLS configuration
	// +optional
	TLS *k8sdbv1alpha1.TLSConfig `json:"tls,omitempty"`

	// Service name of mongo to create or if empty default name will be used
	// +optional
	Service *corev1.LocalObjectReference `json:"service,omitempty"`

	// Port of mongo service to create or if empty will be set with
	// +optional
	Port *int32 `json:"port,omitempty"`
}

// MongoDBStatus defines the observed state of MongoDB
type MongoDBStatus struct {
	// Phase of MongoDB instance health
	// +optional
	Phase MongoDBPhase `json:"phase,omitempty"`

	// Conditions of the instances
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mongodbs,singular=mongodb,shortName=mdb
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// MongoDB is the Schema for the mongodbs API
type MongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBSpec   `json:"spec,omitempty"`
	Status MongoDBStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongoDBList contains a list of MongoDB
type MongoDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDB{}, &MongoDBList{})
}
