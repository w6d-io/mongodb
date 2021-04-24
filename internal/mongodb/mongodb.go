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
Created on 10/04/2021
*/
package mongodb

import (
	"context"
	"fmt"
	db "github.com/w6d-io/mongodb/api/v1alpha1"
	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/k8s/secret"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetClient(ctx context.Context, r client.Client, mongoDB *db.MongoDB) (*mongo.Client, error) {
	log := util.GetLog(ctx, mongoDB).WithName("GetClient")
	log.V(1).Info("create MongoDB client")
	name := fmt.Sprintf("%s/%s", mongoDB.Namespace, mongoDB.Name)
	password := secret.GetContentFromKey(ctx, r, name, secret.MongoRootPasswordKey)
	credential := options.Credential{
		Username: "root",
		Password: password,
	}
	URL := fmt.Sprintf("mongodb://%s", GetService(mongoDB))
	opts := options.Client().ApplyURI(URL).SetAuth(credential)
	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	return c, nil

}

// GetService return the service of mongodb
func GetService(mongoDB *db.MongoDB) string {
	return fmt.Sprintf("%s.%s:%d", mongoDB.Name, mongoDB.Namespace, db.MongoDBPort)
}
