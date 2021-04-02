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
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type configFlag struct {
	value string
}

func (f configFlag) String() string {
	return f.value
}

func (f configFlag) Set(flagValue string) error {
	if flagValue == "" {
		return errors.New("config cannot be empty")
	}
	isFileExists := func(filename string) bool {
		info, err := os.Stat(filename)
		if os.IsNotExist(err) {
			return false
		}
		return !info.IsDir()
	}
	if !isFileExists(flagValue) {
		return fmt.Errorf("file %s does not exist", flagValue)
	}
	if err := New(flagValue); err != nil {
		return fmt.Errorf("instanciate config returns %s", err)
	}
	f.value = flagValue
	return nil
}

func BindFlag(fs *flag.FlagSet) {

	var c configFlag
	fs.Var(&c, "config", "config file")

}
