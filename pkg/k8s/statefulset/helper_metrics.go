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
	"context"
	"fmt"
	db "github.com/w6d-io/mongodb/api/v1alpha1"
	"github.com/w6d-io/mongodb/internal/config"
	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/k8s/secret"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMetricsContainers(ctx context.Context, r client.Client, mongoDB *db.MongoDB) corev1.Container {
	cmd := `|
/bin/mongodb_exporter --web.listen-address ":9216" --mongodb.uri "%s"`
	cmd = fmt.Sprintf(cmd, getMetricsURL(ctx, r, mongoDB))
	return corev1.Container{
		Name:            "metrics",
		Image:           config.GetImage("metrics"),
		ImagePullPolicy: "Always",
		Command: []string{
			"/bin/base",
			"-ec",
		},
		Args: []string{
			cmd,
		},
		Env: []corev1.EnvVar{
			{
				Name: "MONGODB_ROOT_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: mongoDB.Name,
						},
						Key: MongoRootPasswordKey,
					},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			AddVolumeMountTLS(mongoDB.Spec.TLS),
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: MongoContainerMetricsPort,
			},
		},
	}
}

func getMetricsURL(ctx context.Context, r client.Client, mongoDB *db.MongoDB) string {
	password := secret.GetContentFromKey(ctx, r, mongoDB.Name, MongoRootPasswordKey)
	return fmt.Sprintf("mongodb://root:%slocalhost:%d/admin?%s",
		util.EscapePassword(password), MongoContainerPort, getTLSMetricsArgs(mongoDB))
}

func getTLSMetricsArgs(mongoDB *db.MongoDB) string {
	if mongoDB.Spec.TLS == nil {
		return ""
	}
	return "tls=true&tlsCertificateKeyFile=/certs/mongodb.pem&tlsCAFile=/certs/mongodb-ca-cert"
}
