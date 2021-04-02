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
    "fmt"

    "github.com/w6d-io/mongodb/internal/config"
    "github.com/w6d-io/mongodb/internal/util"

    db "github.com/w6d-io/mongodb/api/v1alpha1"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getStatefulSetMongoDB(db db.MongoDB) *appsv1.StatefulSet {
    ls := labelsForMongoDB(db.Name)
    sec := db.Spec.SecurityContext
    if sec == nil {
        sec = config.GetSecurityContext()
    }
    sts := &appsv1.StatefulSet{
        ObjectMeta: metav1.ObjectMeta{
            Name: db.Name,
            Namespace: db.Namespace,
        },
        Spec: appsv1.StatefulSetSpec{
            ServiceName: db.Name,
            PodManagementPolicy: "OrderedReady",
            Replicas: db.Spec.Replicas,
            UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
                Type:          "RollingUpdate",
                RollingUpdate: nil,
            },
            Selector: &metav1.LabelSelector{
                MatchLabels: ls,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels:      ls,
                    Annotations: map[string]string{
                        "checksum/configuration": util.AsSha256(db),
                    },
                },
                Spec: corev1.PodSpec{
                    InitContainers:     getInitContainers(db),
                    Containers:         getContainers(db),
                    NodeSelector:       nil,
                    ServiceAccountName: config.GetServiceAccountName(),
                    SecurityContext:    sec,
                    Affinity:           nil,
                    Tolerations:        nil,
                },
            },
        },
    }
    return sts
}

func labelsForMongoDB(name string) map[string]string {
    return map[string]string{
        "app.kubernetes.io/component": "mongodb",
        "release": name,
    }
}

func getContainers(db db.MongoDB) []corev1.Container {
    return []corev1.Container{
        {
            Name: "mongodb",
            Image: config.GetImage("mongodb"),
            Command: []string{
                "/scripts/setup.sh",
            },
            Env: []corev1.EnvVar{
                {
                    Name:  "BITNAMI_DEBUG",
                    Value: "false",
                },
                {
                    Name:  "MY_POD_NAME",
                    Value: db.Namespace,
                },
                {
                    Name: "K8S_SERVICE_NAME",
                    Value: db.Name,
                },
                {
                    Name: "MONGODB_INITIAL_PRIMARY_HOST",
                    Value: fmt.Sprintf("%s-0.%s.%s.svc.cluster.local", db.Name, db.Name, db.Namespace),
                },
                {
                    Name: "MONGODB_REPLICA_SET_NAME",
                    Value: "rs0",
                },
            },
        },
    }
}

func getInitContainers(db db.MongoDB) []corev1.Container {
    if db.Spec.TLS == nil {
        return []corev1.Container{}
    }
    return []corev1.Container{
        {
            Name: "generate-tls-certs",
            Image: config.GetImage("tls"),
            ImagePullPolicy: "Always",
            Env: []corev1.EnvVar{
                {
                    Name:  "MY_POD_NAMESPACE",
                    Value: db.Namespace,
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
#Create the client/server certificate
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
        },
    }
}
