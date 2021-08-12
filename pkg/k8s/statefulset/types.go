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
Created on 29/03/2021
*/
package statefulset

// TODO Move constant to avoid duplicate entries with secret package
const (
	MongoName                  string = "mongodb"
	MongoRootPasswordKey       string = "mongodb-root-password"
	MongoReplicaSetPasswordKey string = "mongodb-replica-set-key"
	MongoContainerPort         int32  = 27017
	MongoContainerMetricsPort  int32  = 9216
)

type Error struct {
	Cause  error
	Detail string
}
