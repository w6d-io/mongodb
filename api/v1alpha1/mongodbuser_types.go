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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MongoDBUserSpec defines the desired state of MongoDBUser
type MongoDBUserSpec struct {
	// Username is the user name to be create on the MongoDB Instance
	Username string `json:"username,omitempty"`

	// Password is the password associated to the user
	Password Password `json:"password,omitempty"`

	// Privileges
	Privileges []Privilege `json:"privileges,omitempty"`

	// MongoDBRef represents the reference to the mongoDB instance for the user
	DBRef *corev1.LocalObjectReference `json:"dbref"`
}

// Password defines the password of the MongoDB
// One of Value or ValueFrom
type Password struct {
	// Value represents a raw value
	// +optional
	Value *string `json:"value,omitempty"`

	// ValueFrom represent a value from a secret
	ValueFrom *PasswordFrom `json:"valueFrom,omitempty"`
}

type PasswordFrom struct {
	// SecretKeyRef selects a key of secret in the same namespace where password's user is set
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// Privilege defines a link to MongoDB
type Privilege struct {
	// DatabaseName is the name to a MongoDB Database for this privilege
	DatabaseName string `json:"databaseName,omitempty"`
	// Permission is the given permission for this privilege
	Permission Permission `json:"permission"`
}

// Permission define the permission for a privilege
// +kubebuilder:validation:Enum=read;readWrite;dbAdmin;dbOwner;userAdmin;root
type Permission string

// MongoDBUserStatus defines the observed state of MongoDBUser
type MongoDBUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=mongodbusers,singular=mongodbuser,shortName=mgu
//+kubebuilder:printcolumn:name="Username",type="string",JSONPath=".spec.username"
//+kubebuilder:printcolumn:name="Instance",priority=1,type="string",JSONPath=".spec.dbref"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// MongoDBUser is the Schema for the mongodbusers API
type MongoDBUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBUserSpec   `json:"spec,omitempty"`
	Status MongoDBUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongoDBUserList contains a list of MongoDBUser
type MongoDBUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDBUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBUser{}, &MongoDBUserList{})
}
