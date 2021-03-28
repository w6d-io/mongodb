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
Created on 28/03/2021
*/
package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/w6d-io/mongodb/internal/util"
	"os"
)

var _ = Describe("Flags", func() {
	Context("lookup env string", func() {
		BeforeEach(func() {
		})
		AfterEach(func() {
		})
		It("get variable value", func() {
			err := os.Setenv("TEST", "test")
			Expect(err).To(Succeed())
			Expect(util.LookupEnvOrString("TEST", "default")).To(Equal("test"))
		})
		It("get default value", func() {
			err := os.Unsetenv("TEST")
			Expect(err).To(Succeed())
			Expect(util.LookupEnvOrString("TEST", "default")).To(Equal("default"))
		})
		It("get variable value", func() {
			err := os.Setenv("TEST", "true")
			Expect(err).To(Succeed())
			Expect(util.LookupEnvOrBool("TEST", false)).To(Equal(true))
		})
		It("get default value", func() {
			err := os.Unsetenv("TEST")
			Expect(err).To(Succeed())
			Expect(util.LookupEnvOrBool("TEST", false)).To(Equal(false))
		})
	})
})
