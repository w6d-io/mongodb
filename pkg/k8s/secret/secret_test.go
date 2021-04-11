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
package secret_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	db "github.com/w6d-io/mongodb/api/v1alpha1"
	"github.com/w6d-io/mongodb/pkg/k8s/secret"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Secret", func() {
	Context("Helper", func() {
		BeforeEach(func() {
		})
		AfterEach(func() {
		})
		It("gets empty due to key not present", func() {
			var err error
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-wrong-key",
					Namespace: "default",
				},
				StringData: map[string]string{
					"secret": "test-password",
				},
			}
			err = k8sClient.Create(ctx, s)
			Expect(err).To(Succeed())

			c := &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "default/test-sec-wrong-key",
				},
				Key: "key",
			}
			content := secret.GetContentFromKeySelector(ctx, k8sClient, c)
			Expect(content).To(Equal(""))
			err = k8sClient.Delete(ctx, s)
			Expect(err).To(Succeed())
		})
		It("return empty due to selector nil", func() {
			content := secret.GetContentFromKeySelector(ctx, k8sClient, nil)
			Expect(content).To(Equal(""))
		})
		It("failed by secret does not exist", func() {
			c := &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test-sec",
				},
				Key: "secret",
			}
			content := secret.GetContentFromKeySelector(ctx, k8sClient, c)
			Expect(content).To(Equal(""))
		})
		It("gets secret content with selector", func() {
			var err error
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-w-selector",
					Namespace: "default",
				},
				StringData: map[string]string{
					"secret": "test-password",
				},
			}
			err = k8sClient.Create(ctx, s)
			Expect(err).To(Succeed())

			c := &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "default/test-sec-w-selector",
				},
				Key: "secret",
			}
			content := secret.GetContentFromKeySelector(ctx, k8sClient, c)
			Expect(content).To(Equal("test-password"))
			err = k8sClient.Delete(ctx, s)
			Expect(err).To(Succeed())
		})
		It("gets secret content", func() {
			var err error
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-key",
					Namespace: "default",
				},
				StringData: map[string]string{
					"secret": "test-from-key",
				},
			}
			err = k8sClient.Create(ctx, s)
			Expect(err).To(Succeed())

			content := secret.GetContentFromKey(ctx, k8sClient, "default/test-sec-key", "secret")
			Expect(content).To(Equal("test-from-key"))
			err = k8sClient.Delete(ctx, s)
			Expect(err).To(Succeed())
		})
		It("return true because key exists in secret", func() {
			var err error
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-true",
					Namespace: "default",
				},
				StringData: map[string]string{
					"secret": "test-password",
				},
			}
			err = k8sClient.Create(ctx, s)
			Expect(err).To(Succeed())

			c := &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "default/test-sec-true",
				},
				Key: "secret",
			}
			ok := secret.IsKeyExist(ctx, k8sClient, c)
			Expect(ok).To(Equal(true))
			err = k8sClient.Delete(ctx, s)
			Expect(err).To(Succeed())
		})
		It("return false because the secret does not exist", func() {
			c := &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "default/test-sec-not-found",
				},
				Key: "secret",
			}
			ok := secret.IsKeyExist(ctx, k8sClient, c)
			Expect(ok).To(Equal(false))
		})
		It("return false because the secret has a bad name", func() {
			c := &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "default/",
				},
				Key: "secret",
			}
			ok := secret.IsKeyExist(ctx, k8sClient, c)
			Expect(ok).To(Equal(false))
		})
		It("return false because selector is nil", func() {
			ok := secret.IsKeyExist(ctx, k8sClient, nil)
			Expect(ok).To(Equal(false))
		})
	})
	Context("Create", func() {
		It("success with secret already exist", func() {
			var err error
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-root-password",
					Namespace: "default",
				},
				StringData: map[string]string{
					secret.MongoRootPasswordKey: "test-root-password",
				},
			}
			err = k8sClient.Create(ctx, s)
			Expect(err).To(Succeed())
			mongodb := &db.MongoDB{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-root-password",
					Namespace: "default",
				},
			}
			err = secret.Create(ctx, k8sClient, scheme, mongodb)
			Expect(err).To(Succeed())
			err = k8sClient.Delete(ctx, s)
			Expect(err).To(Succeed())
		})
		It("failed due to wrong scheme", func() {
			var err error
			mongodb := &db.MongoDB{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-uui-absent",
					Namespace: "default",
				},
			}
			err = secret.Create(ctx, k8sClient, runtime.NewScheme(), mongodb)
			Expect(err).ToNot(Succeed())
			Expect(err.Error()).To(Equal("get secret return nil"))
		})
		It("fails to create secret due to bad name", func() {
			var err error
			mongodb := &db.MongoDB{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-/",
					Namespace: "default",
					UID:       "099ca89f-1da8-4430-b46f-29d02d8fa9a5",
				},
			}
			err = secret.Create(ctx, k8sClient, scheme, mongodb)
			Expect(err).ToNot(Succeed())
			Expect(err.Error()).To(ContainSubstring("invalid resource name"))
		})
		It("fails to create secret due to", func() {
			var err error
			mongodb := &db.MongoDB{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-ns-not-exists",
					Namespace: "test",
					UID:       "099ca89f-1da8-4430-b46f-29d02d8fa9a5",
				},
			}
			err = secret.Create(ctx, k8sClient, scheme, mongodb)
			Expect(err).ToNot(Succeed())
			Expect(err.Error()).To(ContainSubstring("fail to  create secret : namespaces \"test\" not found"))
		})
		It("success", func() {
			var err error
			mongodb := &db.MongoDB{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sec-root-password-not-exist",
					Namespace: "default",
					UID:       "099ca89f-1da8-4430-b46f-29d02d8fa9a5",
				},
			}
			err = secret.Create(ctx, k8sClient, scheme, mongodb)
			Expect(err).To(Succeed())
		})
	})
})
