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
package statefulset

import (
    db "github.com/w6d-io/mongodb/api/v1alpha1"
    "github.com/w6d-io/mongodb/internal/config"
    corev1 "k8s.io/api/core/v1"
)

func getMetricsContainers(mongoDB *db.MongoDB) corev1.Container {
    return corev1.Container{
        Name:  "metrics",
        Image: config.GetImage("metrics"),
    }
}
