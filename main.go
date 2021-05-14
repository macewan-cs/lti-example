// Copyright (c) 2021 MacEwan University. All rights reserved.
//
// This source code is licensed under the MIT-style license found in
// the LICENSE file in the root directory of this source tree.

package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/macewan-cs/lti/datastore"
	"github.com/macewan-cs/lti/datastore/nonpersistent"
	"github.com/macewan-cs/lti/launch"
	"github.com/macewan-cs/lti/login"
	"github.com/urfave/negroni"
)

func logger(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	source, _, _ := net.SplitHostPort(r.RemoteAddr)
	log.Println("request URI:", r.RequestURI, "method:", r.Method, "ip addr:", source)
	next(w, r)
}

func main() {
	var httpAddr = flag.String("addr", ":8080", "example app listen address")
	flag.Parse()

	var registration datastore.Registration

	// Environment variables:
	// issuer, clientID, authTokenURI, authLoginURI, keysetURI, launchURI
	err := envconfig.Process("reg", &registration)
	if err != nil {
		log.Fatalf("environment parse error: %v", err)
	}

	err = nonpersistent.DefaultStore.StoreRegistration(registration)
	if err != nil {
		log.Fatalf("store error: %v", err)
	}

	mux := http.NewServeMux()
	// Use a blank configuration to get default nonpersistent datastore.
	login := login.New(login.Config{})
	mux.Handle("/login", login)
	launch := launch.New(launch.Config{})
	mux.Handle("/launch", launch)

	n := negroni.New()
	n.Use(negroni.HandlerFunc(logger))
	n.UseHandler(mux)

	err = http.ListenAndServe(*httpAddr, n)
	if err != nil {
		log.Fatalf("http server error: %v", err)
	}
}
