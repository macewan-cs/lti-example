// Copyright (c) 2021 MacEwan University. All rights reserved.
//
// This source code is licensed under the MIT-style license found in
// the LICENSE file in the root directory of this source tree.

package env

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/macewan-cs/lti/datastore"
)

// RegistrationFromEnvironment loads the registration details from
// environment variables. The expected variables include REG_ISSUER,
// REG_CLIENTID, REG_KEYSETURI, REG_AUTHTOKENURI, REG_AUTHLOGINURI,
// REG_TARGETLINKURI.
func RegistrationFromEnvironment() datastore.Registration {
	var registration datastore.Registration
	err := envconfig.Process("reg", &registration)
	if err != nil {
		log.Fatalf("registration environment parse error: %v", err)
	}

	return registration
}

// DeploymentFromEnvironment loads the deployment details from
// environment variables. The expected variable is DEP_DEPLOYMENTID.
func DeploymentFromEnvironment() datastore.Deployment {
	var deployment datastore.Deployment
	err := envconfig.Process("dep", &deployment)
	if err != nil {
		log.Fatalf("deployment environment parse error: %v", err)
	}

	return deployment
}

// Key holds the private key read from an environment variable.
type Key struct {
	Private string
}

// KeyFromEnvironment loads the key details from environment
// variables. The expected variable is KEY_PRIVATE.
func KeyFromEnvironment() Key {
	var key Key
	err := envconfig.Process("key", &key)
	if err != nil {
		log.Fatalf("key environment parse error: %v", err)
	}

	return key
}
