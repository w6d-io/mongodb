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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/w6d-io/mongodb/internal/mongodb"
	"github.com/w6d-io/mongodb/internal/util"
	"github.com/w6d-io/mongodb/pkg/k8s/secret"
	"go.mongodb.org/mongo-driver/bson"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	db "github.com/w6d-io/mongodb/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Create(ctx context.Context, r client.Client, user *db.MongoDBUser) error {
	log := util.GetLog(ctx, user).WithName("User").WithName("Create")
	log.V(1).Info("create MongoDB user")
	ok, err := IsUserExist(ctx, r, user)
	if err != nil {
		log.Error(err, "check user exist failed")
		return err
	}
	if ok {
		return Update(ctx, r, user)
	}
	mdb, err := GetMongoDB(ctx, r, user)
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
	d := c.Database("admin")
	res := d.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: user.Spec.Username},
		{Key: "customData", Value: bson.D{
			{Key: "parentID", Value: user.UID},
		}},
		{Key: "pwd", Value: passwd},
		{Key: "roles", Value: GetPrivileges(ctx, user)},
	})
	if res.Err() != nil {
		log.Error(res.Err(), "create user failed")
		return res.Err()
	}
	return nil
}

func Update(ctx context.Context, r client.Client, user *db.MongoDBUser) error {
	log := util.GetLog(ctx, user).WithName("User").WithName("Update")
	log.V(1).Info("update MongoDB user")
	users, err := GetUser(ctx, r, user)
	if err != nil {
		log.Error(err, "get user failed")
		return err
	}
	if users[0].CustomData.ParentID != string(user.GetUID()) {
		log.Error(nil, "this user is already handle by an other resource", "user", user.Spec.Username)
		return errors.New("a user can be handle only by one resource")
	}
	mdb, err := GetMongoDB(ctx, r, user)
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
	d := c.Database("admin")
	res := d.RunCommand(ctx, bson.D{
		{Key: "updateUser", Value: user.Spec.Username},
		{Key: "pwd", Value: passwd},
		{Key: "roles", Value: GetPrivileges(ctx, user)},
	})
	if res.Err() != nil {
		log.Error(res.Err(), "update user failed")
		return res.Err()
	}

	return nil
}

func Delete(ctx context.Context, r client.Client, user *db.MongoDBUser) error {
	log := util.GetLog(ctx, user).WithName("User").WithName("Delete").WithValues("username", user.Spec.Username)
	log.V(1).Info("delete MongoDB user")
	mdb, err := GetMongoDB(ctx, r, user)
	if err != nil {
		log.Error(err, "get MongoDB failed")
		return err
	}
	users, err := GetUser(ctx, r, user)
	if err != nil {
		log.Error(err, "get user failed")
		return err
	}
	log.V(1).Info("users", "content", users)
	if len(users) == 0 {
		return nil
	}
	if users[0].CustomData.ParentID != string(user.GetUID()) {
		log.V(1).Info("skipped deletion", "user", user.Spec.Username)
		return nil
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
		log.V(1).Info("delete user", "database", priv.DatabaseName)
		d := c.Database("admin")
		res := d.RunCommand(ctx, bson.D{
			{Key: "dropUser", Value: user.Spec.Username}})
		if res.Err() != nil && !IsNotFound(res.Err()) {
			log.Error(res.Err(), "delete user failed")
			return res.Err()
		}
		log.V(1).Info("user deleted", "database", priv.DatabaseName)
	}
	return nil
}

// GetMongoDB return the mongoDB resource referenced by the name
func GetMongoDB(ctx context.Context, r client.Client, user *db.MongoDBUser) (*db.MongoDB, error) {
	correlationID := ctx.Value("correlation_id")
	log := ctrl.Log.WithValues("correlation_id", correlationID).WithName("User").WithName("GetMongoDB")
	log.V(1).Info("get")

	if user.Spec.ExternalRef != nil {
		return &db.MongoDB{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: user.Namespace,
			},
			Spec: db.MongoDBSpec{
				AuthSecret: user.Spec.ExternalRef.Auth,
				Service:    &corev1.LocalObjectReference{Name: user.Spec.ExternalRef.Service},
				Port:       user.Spec.ExternalRef.Port,
			},
		}, nil
	}
	mongoDB := &db.MongoDB{}
	if err := r.Get(ctx, types.NamespacedName{Name: user.Spec.DBRef.Name, Namespace: user.Namespace}, mongoDB); err != nil {
		log.Error(err, "get MongoDB failed")
		return nil, err
	}
	return mongoDB, nil
}

// GetUserPassword get the password from either in value or in valueFrom
func GetUserPassword(ctx context.Context, r client.Client, user *db.MongoDBUser) string {
	correlationID := ctx.Value("correlation_id")
	log := ctrl.Log.WithValues("correlation_id", correlationID).WithName("User").WithName("GetUserPassword")
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

// IsNotFound check whether or not the error message container "not found"
func IsNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

// IsUserExist check if user exist in database
func IsUserExist(ctx context.Context, r client.Client, user *db.MongoDBUser) (bool, error) {
	log := util.GetLog(ctx, user).WithName("User").WithName("IsUserExist")
	log.V(1).Info("Is MongoDB user exist")

	rsp, err := GetUser(ctx, r, user)
	if err != nil {
		return false, err
	}
	if len(rsp) == 1 {
		return true, nil
	}
	return false, nil
}

// GetPrivileges return a bson.M slice with all user's privilege
func GetPrivileges(ctx context.Context, user *db.MongoDBUser) []bson.M {
	log := util.GetLog(ctx, user).WithName("User").WithName("GetPrivileges")
	log.V(1).Info("get privileges")
	var p []bson.M
	for _, priv := range user.Spec.Privileges {
		p = append(p, bson.M{
			"role": priv.Permission,
			"db":   priv.DatabaseName,
		})
	}
	return p
}

func GetUsers(ctx context.Context, r client.Client, user *db.MongoDBUser) (*Response, error) {
	log := util.GetLog(ctx, user).WithName("User").WithName("GetUsers")
	log.V(1).Info("get MongoDB users")
	mdb, err := GetMongoDB(ctx, r, user)
	if err != nil {
		log.Error(err, "get MongoDB failed")
		return nil, err
	}
	c, err := mongodb.GetClient(ctx, r, mdb)
	if err != nil {
		log.Error(err, "get MongoDB client")
		return nil, err
	}
	if err = c.Ping(ctx, nil); err != nil {
		log.Error(err, "ping db failed")
		return nil, err
	}

	d := c.Database("admin")
	res := d.RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: 1},
	})
	var response = new(Response)
	err = res.Decode(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func GetUser(ctx context.Context, r client.Client, user *db.MongoDBUser) ([]User, error) {
	log := util.GetLog(ctx, user).WithName("User").WithName("GetUsers")
	log.V(1).Info("get MongoDB user")

	rsp, err := GetUsers(ctx, r, user)
	if err != nil {
		log.Error(err, "get users failed")
		return nil, err
	}
	var users []User
	for _, usr := range rsp.Users {
		if usr.User == user.Spec.Username {
			users = append(users, usr)
		}
	}
	return users, nil
}
