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
Created on 01/04/2021
*/
package statefulset

import (
	"context"
	"fmt"

	"github.com/w6d-io/mongodb/internal/config"
	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/k8s/configmap"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	k8sdbv1alpha1 "github.com/w6d-io/mongodb/apis/k8sdb/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getStatefulSetMongoDB(ctx context.Context, r client.Client, mongoDB *db.MongoDB) *appsv1.StatefulSet {
	log := util.GetLog(ctx, mongoDB)
	ls := util.LabelsForMongoDB(mongoDB.Name)
	log.V(1).Info("build statefulSet")
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongoDB.Name,
			Namespace: mongoDB.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: mongoDB.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
					Annotations: map[string]string{
						"checksum/configuration": util.AsSha256(mongoDB),
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						getInitContainers(mongoDB),
					},
					Containers: []corev1.Container{
						getContainers(mongoDB),
						getMetricsContainers(ctx, r, mongoDB),
					},
					NodeSelector:       util.GetNodeSelector(mongoDB.Spec.PodTemplate.NodeSelector),
					ServiceAccountName: util.GetServiceAccount(mongoDB.Spec.PodTemplate.ServiceAccountName),
					SecurityContext:    util.GetSecurityContext(mongoDB.Spec.PodTemplate.SecurityContext),
					Affinity:           util.GetAffinity(mongoDB.Spec.PodTemplate.Affinity),
					Tolerations:        util.GetTolerations(mongoDB.Spec.PodTemplate.Tolerations),
					Volumes:            append(AddVolumes(mongoDB), AddVolumeTLS(mongoDB.Spec.TLS)...),
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "datadir",
					},
					Spec: mongoDB.Spec.Storage,
				},
			},
			ServiceName:         mongoDB.Name,
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type:          appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: nil,
			},
		},
	}
	return sts
}

func getContainers(mongoDB *db.MongoDB) corev1.Container {
	container := corev1.Container{
		Name:  "mongodb",
		Image: config.GetImage(MongoName),
		Command: []string{
			"/scripts/setup.sh",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          MongoName,
				ContainerPort: MongoContainerPort,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "BITNAMI_DEBUG",
				Value: "false",
			},
			{
				Name:  "MY_POD_NAME",
				Value: mongoDB.Namespace,
			},
			{
				Name:  "K8S_SERVICE_NAME",
				Value: mongoDB.Name,
			},
			{
				Name:  "MONGODB_INITIAL_PRIMARY_HOST",
				Value: fmt.Sprintf("%s-0.%s.%s.svc.cluster.local", mongoDB.Name, mongoDB.Name, mongoDB.Namespace),
			},
			{
				Name:  "MONGODB_REPLICA_SET_NAME",
				Value: "rs0",
			},
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
			{
				Name:  "ALLOW_EMPTY_PASSWORD",
				Value: "no",
			},
			{
				Name:  "MONGODB_SYSTEM_LOG_VERBOSITY",
				Value: "0",
			},
			{
				Name:  "MONGODB_DISABLE_SYSTEM_LOG",
				Value: "no",
			},
			{
				Name:  "MONGODB_DISABLE_JAVASCRIPT",
				Value: "no",
			},
			{
				Name:  "MONGODB_ENABLE_IPV6",
				Value: "no",
			},
			{
				Name:  "MONGODB_ENABLE_DIRECTORY_PER_DB",
				Value: "no",
			},
			AddEnvTLS(mongoDB.Spec.TLS),
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "datadir",
				MountPath: "/bitnami/mongodb",
			},
			{
				Name:      "scripts",
				MountPath: "/scripts/setup.sh",
				SubPath:   "setup.sh",
			},
		},
	}
	return container
}

func getInitContainers(mongoDB *db.MongoDB) corev1.Container {
	if mongoDB.Spec.TLS == nil {
		return corev1.Container{}
	}
	return corev1.Container{

		Name:            "generate-tls-certs",
		Image:           config.GetImage("tls"),
		ImagePullPolicy: "Always",
		Env: []corev1.EnvVar{
			{
				Name:  "MY_POD_NAMESPACE",
				Value: mongoDB.Namespace,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "certs-volume",
				MountPath: "/certs/CAs",
			},
			{
				Name:      "certs",
				MountPath: "/certs",
			},
			AddVolumeMountTLS(mongoDB.Spec.TLS),
		},
		Command: []string{
			"sh",
			"-c",
			`|
/bin/bash <<'EOF'
my_hostname=$(hostname)
svc=$(echo -n "$my_hostname" | sed s/-[0-9]*$//)-headless
cp /certs/CAs/* /certs/
cat >/certs/openssl.cnf <<EOL
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = $svc
DNS.2 = $my_hostname
DNS.3 = $my_hostname.$svc.$MY_POD_NAMESPACE.svc.cluster.local
DNS.4 = localhost
DNS.5 = 127.0.0.1
{{- if .Values.externalAccess.service.loadBalancerIPs }}
{{- range $key, $val := .Values.externalAccess.service.loadBalancerIPs }}
IP.{{ $key }} = {{ $val | quote }}
{{- end }}
{{- end }}
EOL

export RANDFILE=/certs/.rnd && openssl genrsa -out /certs/mongo.key 2048
#CreateUpdate the client/server certificate
openssl req -new -key /certs/mongo.key -out /certs/mongo.csr -subj "/C=US/O=My Organisations/OU=IT/CN=$my_hostname" -config /certs/openssl.cnf
#Signing the server certificate with the CA cert and key
openssl x509 -req -in /certs/mongo.csr -CA /certs/mongodb-ca-cert -CAkey /certs/mongodb-ca-key -CAcreateserial -out /certs/mongo.crt -days 3650 -extensions v3_req -extfile /certs/openssl.cnf
rm /certs/mongo.csr
#Concatenate to a pem file for use as the client PEM file which can be used for both member and client authentication.
cat /certs/mongo.crt /certs/mongo.key > /certs/mongodb.pem
cd /certs/
shopt -s extglob
rm -rf !(mongodb-ca-cert|mongodb.pem|CAs|openssl.cnf)
chmod 0600 mongodb-ca-cert mongodb.pem
EOF
`,
		},
	}
}

func AddVolumes(mongoDB *db.MongoDB) []corev1.Volume {
	var v []corev1.Volume
	v = append(v, configmap.GetVolume("script", util.GetLocalObjectReference(mongoDB.Name+"-scripts")))
	return v
}

func AddEnvTLS(tlsConfig *k8sdbv1alpha1.TLSConfig) corev1.EnvVar {
	env := corev1.EnvVar{}
	if tlsConfig != nil {
		env = corev1.EnvVar{
			Name:  "MONGODB_EXTRA_FLAGS",
			Value: "--tlsMode=requireTLS --tlsCertificateKeyFile=/certs/mongodb.pem --tlsCAFile=/certs/mongodb-ca-cert",
		}
	}
	return env
}

func AddVolumeMountTLS(tlsConfig *k8sdbv1alpha1.TLSConfig) corev1.VolumeMount {
	vm := corev1.VolumeMount{}
	if tlsConfig != nil {
		vm = corev1.VolumeMount{
			Name:      "certs",
			MountPath: "/certs",
		}
	}
	return vm
}

func AddVolumeTLS(tlsConfig *k8sdbv1alpha1.TLSConfig) []corev1.Volume {
	var v []corev1.Volume
	var mode int32 = 0600
	if tlsConfig != nil {
		v = append(v, corev1.Volume{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
		v = append(v, corev1.Volume{
			Name: "certs-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					Items: []corev1.KeyToPath{
						{
							Key:  "mongodb-ca-cert",
							Path: "mongodb-ca-cert",
							Mode: &mode,
						},
						{
							Key:  "mongodb-ca-key",
							Path: "mongodb-ca-key",
							Mode: &mode,
						},
					},
				},
			},
		})
	}
	return v
}
