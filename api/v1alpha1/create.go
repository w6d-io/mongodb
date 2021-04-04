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
Created on 04/04/2021
*/
package v1alpha1

import (
	"github.com/w6d-io/mongodb/internal/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	AccessModes = []string{
		"ReadWriteOnce",
		"ReadOnlyMany",
		"ReadWriteMany",
	}
)

func Create(mongoDB *MongoDB) error {
	var allErrs field.ErrorList
	for _, accessMode := range mongoDB.Spec.Storage.AccessModes {
		if !util.StringInArray(string(accessMode), AccessModes) {
			allErrs = append(allErrs,
				field.Invalid(field.NewPath("spec").Child("storage").Child("accessModes"),
					accessMode,
					"it should be either ReadWriteOnce, ReadOnlyMny or ReadWriteMany"))
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "db.w6d.io", Kind: "MongoDB"},
		mongoDB.Name, allErrs)
}

func Update(old, new *MongoDB) error {
	var allErrs field.ErrorList
	if len(old.Spec.Storage.AccessModes) != len(new.Spec.Storage.AccessModes) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec").Child("storage").Child("accessModes"),
				new.Spec.Storage.AccessModes,
				"it can be changed",
			))
	} else {
		for _, accessMode := range new.Spec.Storage.AccessModes {
			if !util.AccessModeIn(accessMode, old.Spec.Storage.AccessModes) {
				allErrs = append(allErrs,
					field.Invalid(field.NewPath("spec").Child("storage").Child("accessModes"),
						accessMode,
						"could not be updated"))
			}
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "db.w6d.io", Kind: "MongoDB"},
		old.Name, allErrs)
}
