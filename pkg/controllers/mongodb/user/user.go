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
package user

import (
	"context"
	"errors"
	"github.com/w6d-io/mongodb/internal/mongodb"
	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/k8s/secret"
	"go.mongodb.org/mongo-driver/bson"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
)

func Create(ctx context.Context, r client.Client, user *db.MongoDBUser) error {
	log := util.GetLog(ctx, user).WithName("Create")
	log.V(1).Info("create MongoDB user")
	mdb, err := GetMongoDB(ctx, r, types.NamespacedName{Namespace: user.Namespace, Name: user.Spec.DBRef.Name})
	if err != nil {
		log.Error(err, "get MongoDB failed")
		return err
	}
	c, err := mongodb.GetClient(ctx, r, mdb)
	if err != nil {
		log.Error(err, "get MongoDB client")
		return err
	}
	if err = c.Ping(ctx, nil); err != nil {
		log.Error(err, "ping db failed")
		return err
	}
	passwd := GetUserPassword(ctx, r, user)
	if passwd == "" {
		log.Error(nil, "password cannot be empty")
		return errors.New("password cannot be empty")
	}
	for _, priv := range user.Spec.Privileges {
		d := c.Database("admin")
		res := d.RunCommand(ctx, bson.D{
			{Key: "createUser", Value: user.Spec.Username},
			{Key: "pwd", Value: passwd},
			{Key: "roles", Value: []bson.M{{"role": priv.Permission, "db": priv.DatabaseName}}},
		})
		if res.Err() != nil {
			log.Error(err, "create user failed")
			return err
		}
	}
	return nil
}

func Delete(ctx context.Context, r client.Client, user *db.MongoDBUser) error {
	log := util.GetLog(ctx, user).WithName("Delete")
	log.V(1).Info("delete MongoDB user")
	mdb, err := GetMongoDB(ctx, r, types.NamespacedName{Namespace: user.Namespace, Name: user.Spec.DBRef.Name})
	if err != nil {
		log.Error(err, "get MongoDB failed")
		return err
	}
	c, err := mongodb.GetClient(ctx, r, mdb)
	if err != nil {
		log.Error(err, "get MongoDB client")
		return err
	}
	if err = c.Ping(ctx, nil); err != nil {
		log.Error(err, "ping db failed")
		return err
	}
	passwd := GetUserPassword(ctx, r, user)
	if passwd == "" {
		log.Error(nil, "password cannot be empty")
		return errors.New("password cannot be empty")
	}
	for _, priv := range user.Spec.Privileges {
		d := c.Database(priv.DatabaseName)
		res := d.RunCommand(ctx, bson.D{
			{Key: "dropUser", Value: user.Spec.Username}})
		if res.Err() != nil {
			log.Error(err, "delete user failed")
			return err
		}
	}
	return nil
}

func GetMongoDB(ctx context.Context, r client.Client, name types.NamespacedName) (*db.MongoDB, error) {
	correlationID := ctx.Value("correlation_id")
	log := ctrl.Log.WithValues("correlation_id", correlationID, "object", name.String()).WithName("GetMongoDB")
	log.V(1).Info("get")
	mongoDB := &db.MongoDB{}
	if err := r.Get(ctx, name, mongoDB); err != nil {
		log.Error(err, "get MongoDB failed")
		return nil, err
	}
	return mongoDB, nil
}

func GetUserPassword(ctx context.Context, r client.Client, user *db.MongoDBUser) string {
	correlationID := ctx.Value("correlation_id")
	log := ctrl.Log.WithValues("correlation_id", correlationID).WithName("GetUserPassword")
	log.V(1).Info("get")
	if user.Spec.Password.Value != nil {
		return *user.Spec.Password.Value
	}
	if user.Spec.Password.ValueFrom == nil {
		return ""
	}
	return secret.GetContentFromKey(
		ctx,
		r,
		user.Namespace+"/"+user.Spec.Password.ValueFrom.SecretKeyRef.Name,
		user.Spec.Password.ValueFrom.SecretKeyRef.Key)
}
