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

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
)

// New get the filename and fill Config struct
func New(filename string) error {
	// TODO add dynamic configuration feature
	log := ctrl.Log.WithName("controllers").WithName("Config")
	log.V(1).Info("read config file")
	config = new(Config)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error(err, "error reading the configuration")
		return err
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		log.Error(err, "Error unmarshal the configuration")
		return err
	}
	config.Namespace = os.Getenv("NAMESPACE")
	if config.ServiceAccount.Name == "" {
		config.ServiceAccount.Name = "default"
	}
	return nil
}

// GetNamespace return the namespace
func GetNamespace() string {
	return config.Namespace
}

// GetServiceAccountName return the Service Account Name
func GetServiceAccountName() string {
	return config.ServiceAccount.Name
}

func GetImage(key string) string {
	return string(config.Images[key])
}

func GetSecurityContext() *corev1.PodSecurityContext {
	if config.SecurityContext == nil {
		return &corev1.PodSecurityContext{}
	}
	return config.SecurityContext
}

func GetNodeSelector() map[string]string {
	return config.NodeSelector
}

func GetAffinity() *corev1.Affinity {
	return config.Affinity
}

func GetTolerations() []corev1.Toleration {
	return config.Tolerations
}
