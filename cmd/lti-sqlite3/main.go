// Copyright (c) 2021 MacEwan University. All rights reserved.
//
// This source code is licensed under the MIT-style license found in
// the LICENSE file in the root directory of this source tree.

// Package main implements an example of some the LTI library features. Unlike the program in ../minimal, this program
// uses an SQL database for registration/deployment storage and a nonpersistent store for the other data it needs to
// store.
//
// On startup, the program loads all configuration data from environment variables.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/macewan-cs/lti"
	"github.com/macewan-cs/lti-example/internal/env"
	"github.com/macewan-cs/lti/connector"
	"github.com/macewan-cs/lti/datastore"

	_ "github.com/mattn/go-sqlite3"
)

const sqlite3Database = "test.db"

// mustNotExist attempts to read the specified filename, and if it can be read, it terminates the program with an error.
func mustNotExist(filename string) {
	if f, err := os.Open(filename); err == nil {
		f.Close()
		log.Fatalf("database file already exists (%s)", filename)
	}
}

// populateRegistrationAndDeployment attempts to populate the database with tables and an initial registration and deployment. If it
// encounters any errors, it terminates the program with an error.
func populateRegistrationAndDeployment(db *sql.DB) {
	createStatements := []string{
		`CREATE TABLE registration ( issuer text, client_id text, auth_token_uri text, auth_login_uri text,
                                             keyset_uri text, target_link_uri text,
                                             UNIQUE (issuer) )`,
		`CREATE TABLE deployment ( issuer text, deployment_id text,
                                           UNIQUE (issuer, deployment_id) )`,
	}

	for _, createStatement := range createStatements {
		_, err := db.Exec(createStatement)
		if err != nil {
			log.Fatalf("cannot create table: %v", err)
		}
	}

	registration := env.RegistrationFromEnvironment()

	q := `INSERT INTO registration ( issuer, client_id, auth_token_uri, auth_login_uri, keyset_uri,
                                            target_link_uri )
                   VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(q, registration.Issuer, registration.ClientID, registration.AuthTokenURI.String(),
		registration.AuthLoginURI.String(), registration.KeysetURI.String(), registration.TargetLinkURI.String())
	if err != nil {
		log.Fatalf("cannot populate registration: %v", err)
	}

	deployment := env.DeploymentFromEnvironment()

	q = `INSERT INTO deployment (issuer, deployment_id)
                  VALUES ($1, $2)`
	_, err = db.Exec(q, registration.Issuer, deployment.DeploymentID)
	if err != nil {
		log.Fatalf("cannot populate deployment: %v", err)
	}
}

// sqlConfig returns a datastore.Config, which is suitable for creating LTI login handlers, LTI launch handlers, and
// after a launch, LTI connectors.
func sqlConfig() datastore.Config {
	// Create an sqlite3 database for testing.
	mustNotExist(sqlite3Database)
	db, err := sql.Open("sqlite3", sqlite3Database)
	if err != nil {
		log.Fatalf("registration database error: %v", err)
	}

	// Populate the database with registration and deployment details from environment variables.
	populateRegistrationAndDeployment(db)

	datastoreConfig := lti.NewDatastoreConfig()
	sqlDatastore := lti.NewSQLDatastore(db, lti.NewSQLDatastoreConfig())
	datastoreConfig.Registrations = sqlDatastore

	return datastoreConfig
}

// postLaunchHandler returns an http.HandlerFunc suitable for the second argument of lti.NewLaunch.
func postLaunchHandler(datastoreConfig datastore.Config) http.HandlerFunc {
	// Retrieve the key from environment variables.
	key := env.KeyFromEnvironment()

	return func(w http.ResponseWriter, r *http.Request) {
		// Create a connector, which is necessary to access LTI services.
		conn, err := connector.New(datastoreConfig, lti.LaunchIDFromRequest(r))
		if err != nil {
			log.Printf("cannot create connector for launch: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		conn.SetSigningKey(key.Private)

		// Upgrade the connector to access Name and Role Provisioning Services.
		nrps, err := conn.UpgradeNRPS()
		if err != nil {
			log.Printf("cannot upgrade connector for NRPS: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Get membership to demonstrate access to NRPS.
		membership, err := nrps.GetMembership()
		if err != nil {
			log.Printf("cannot get membership: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `<p>Launch successful!</p>
<p>Launch ID from request: %s</p>
<p>Course title: %s</p>`, lti.LaunchIDFromRequest(r), membership.Context.Title)
	}
}

// logRequest logs a request made to the HTTP server.
func logRequest(r *http.Request) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(struct {
		RequestURI string `json:"requestUri"`
		Method     string `json:"method"`
		RemoteAddr string `json:"remoteAddr"`
	}{
		RequestURI: r.RequestURI,
		Method:     r.Method,
		RemoteAddr: r.RemoteAddr,
	})
}

func main() {
	var httpAddr = flag.String("addr", ":8080", "example app listen address")
	flag.Parse()

	key := env.KeyFromEnvironment()
	datastoreConfig := sqlConfig()
	http.Handle("/login", lti.NewLogin(datastoreConfig))
	http.Handle("/launch", lti.NewLaunch(datastoreConfig,
		postLaunchHandler(datastoreConfig)))
	http.Handle("/keyset", lti.NewKeyset("MyKeyIdentifier", key.Private))

	log.Printf("Listening for connections on %s...\n", *httpAddr)
	err := http.ListenAndServe(*httpAddr,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logRequest(r)
			http.DefaultServeMux.ServeHTTP(w, r)
		}),
	)
	if err != nil {
		log.Fatalf("http server error: %v", err)
	}
}
