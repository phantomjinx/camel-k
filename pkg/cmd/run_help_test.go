/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfigOption(t *testing.T) {
	validConfigMap := "configmap:my-config_map"
	validSecret := "secret:my-secret"
	validFile := "file:/tmp/my-file.txt"
	notValid := "someprotocol:wrong"

	configmap, err := ParseConfigOption(validConfigMap)
	assert.Nil(t, err)
	assert.Equal(t, ConfigOptionTypeConfigmap, configmap.ConfigType)
	assert.Equal(t, "my-config_map", configmap.Value)
	secret, err := ParseConfigOption(validSecret)
	assert.Nil(t, err)
	assert.Equal(t, ConfigOptionTypeSecret, secret.ConfigType)
	assert.Equal(t, "my-secret", secret.Value)
	file, err := ParseConfigOption(validFile)
	assert.Nil(t, err)
	assert.Equal(t, ConfigOptionTypeFile, file.ConfigType)
	assert.Equal(t, "/tmp/my-file.txt", file.Value)
	_, err = ParseConfigOption(notValid)
	assert.NotNil(t, err)
}
