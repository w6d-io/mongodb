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

const (
	AutoDiscovery string = `|-
    #!/bin/bash
    SVC_NAME="${MY_POD_NAME}-external"
    # Auxiliary functions
    retry_while() {
        local -r cmd="${1:?cmd is missing}"
        local -r retries="${2:-12}"
        local -r sleep_time="${3:-5}"
        local return_value=1
        read -r -a command <<< "$cmd"
        for ((i = 1 ; i <= retries ; i+=1 )); do
            "${command[@]}" && return_value=0 && break
            sleep "$sleep_time"
        done
        return $return_value
    }
    k8s_svc_lb_ip() {
        local namespace=${1:?namespace is missing}
        local service=${2:?service is missing}
        local service_ip=$(kubectl get svc "$service" -n "$namespace" -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
        local service_hostname=$(kubectl get svc "$service" -n "$namespace" -o jsonpath="{.status.loadBalancer.ingress[0].hostname}")
        if [[ -n ${service_ip} ]]; then
            echo "${service_ip}"
        else
            echo "${service_hostname}"
        fi
    }
    k8s_svc_lb_ip_ready() {
        local namespace=${1:?namespace is missing}
        local service=${2:?service is missing}
        [[ -n "$(k8s_svc_lb_ip "$namespace" "$service")" ]]
    }
    # Wait until LoadBalancer IP is ready
    retry_while "k8s_svc_lb_ip_ready %s $SVC_NAME" || exit 1
    # Obtain LoadBalancer external IP
    k8s_svc_lb_ip "%s" "$SVC_NAME" | tee "$SHARED_FILE"
  {{- end }}
`
	Setup string = `|-
    #!/bin/bash
    . /opt/bitnami/scripts/mongodb-env.sh
    echo "Advertised Hostname: $MONGODB_ADVERTISED_HOSTNAME"
    if [[ "$MY_POD_NAME" = "%s" ]]; then
        echo "Pod name matches initial primary pod name, configuring node as a primary"
        export MONGODB_REPLICA_SET_MODE="primary"
    else
        echo "Pod name doesn't match initial primary pod name, configuring node as a secondary"
        export MONGODB_REPLICA_SET_MODE="secondary"
        export MONGODB_INITIAL_PRIMARY_ROOT_PASSWORD="$MONGODB_ROOT_PASSWORD"
        export MONGODB_INITIAL_PRIMARY_PORT_NUMBER="$MONGODB_PORT_NUMBER"
        export MONGODB_ROOT_PASSWORD="" MONGODB_USERNAME="" MONGODB_DATABASE="" MONGODB_PASSWORD=""
    fi
    exec /opt/bitnami/scripts/mongodb/entrypoint.sh /opt/bitnami/scripts/mongodb/run.sh
`
	SetupHidden = `|-
    #!/bin/bash
    . /opt/bitnami/scripts/mongodb-env.sh
    echo "Advertised Hostname: $MONGODB_ADVERTISED_HOSTNAME"
    echo "Configuring node as a hidden node"
    export MONGODB_REPLICA_SET_MODE="hidden"
    export MONGODB_INITIAL_PRIMARY_ROOT_PASSWORD="$MONGODB_ROOT_PASSWORD"
    export MONGODB_INITIAL_PRIMARY_PORT_NUMBER="$MONGODB_PORT_NUMBER"
    export MONGODB_ROOT_PASSWORD="" MONGODB_USERNAME="" MONGODB_DATABASE="" MONGODB_PASSWORD=""
    exec /opt/bitnami/scripts/mongodb/entrypoint.sh /opt/bitnami/scripts/mongodb/run.sh
`

	ReplicasetEntrypoint = `
    #!/bin/bash

    sleep 5

    . /liblog.sh

    # Perform adaptations depending on the host name
    if [[ $HOSTNAME =~ (.*)-0$ ]]; then
      info "Setting node as primary"
      export MONGODB_REPLICA_SET_MODE=primary
    else
      info "Setting node as secondary"
      export MONGODB_REPLICA_SET_MODE=secondary
      export MONGODB_INITIAL_PRIMARY_ROOT_PASSWORD="$MONGODB_ROOT_PASSWORD"
      unset MONGODB_ROOT_PASSWORD
    fi

    exec /entrypoint.sh /run.sh
`
)
