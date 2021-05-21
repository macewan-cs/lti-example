// Copyright (c) 2021 MacEwan University. All rights reserved.
//
// This source code is licensed under the MIT-style license found in
// the LICENSE file in the root directory of this source tree.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/macewan-cs/lti"
	"github.com/macewan-cs/lti/datastore"
	"github.com/macewan-cs/lti/datastore/nonpersistent"
)

func main() {
	var httpAddr = flag.String("addr", ":8080", "example app listen address")
	flag.Parse()

	// Environment variables. Set to your values to test/demo.
	// Registration: ('reg_' + ) issuer, clientID, authTokenURI, authLoginURI, keysetURI, launchURI
	var registration datastore.Registration
	err := envconfig.Process("reg", &registration)
	if err != nil {
		log.Fatalf("registration environment parse error: %v", err)
	}
	err = nonpersistent.DefaultStore.StoreRegistration(registration)
	if err != nil {
		log.Fatalf("registration store error: %v", err)
	}

	// Deployment: ('dep_' + ) deploymentID
	var deployment datastore.Deployment
	err = envconfig.Process("dep", &deployment)
	if err != nil {
		log.Fatalf("deployment environment parse error: %v", err)
	}
	err = nonpersistent.DefaultStore.StoreDeployment(registration.Issuer, deployment.DeploymentID)
	if err != nil {
		log.Fatalf("deployment store error: %v", err)
	}

	// Use a blank configuration to get default nonpersistent datastore.
	http.Handle("/login", lti.NewLogin(lti.NewLoginConfig()))
	http.Handle("/launch",
		lti.NewLaunch(lti.NewLaunchConfig(),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Launch successful.")
			})))

	err = http.ListenAndServe(*httpAddr,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			source, _, _ := net.SplitHostPort(r.RemoteAddr)
			log.Println("request URI:", r.RequestURI, "method:", r.Method, "ip addr:", source)

			http.DefaultServeMux.ServeHTTP(w, r)
		}),
	)
	if err != nil {
		log.Fatalf("http server error: %v", err)
	}
}
