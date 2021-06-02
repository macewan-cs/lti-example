// Copyright (c) 2021 MacEwan University. All rights reserved.
//
// This source code is licensed under the MIT-style license found in
// the LICENSE file in the root directory of this source tree.

package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/macewan-cs/lti"
	"github.com/macewan-cs/lti/connector"
	"github.com/macewan-cs/lti/datastore"
	"github.com/macewan-cs/lti/datastore/nonpersistent"

	_ "github.com/mattn/go-sqlite3"
)

const sqlite3Database = "test.db"

// registrationFromEnvironment loads the registration details from environment variables.
//
// Variables:
// ('reg_' + ) issuer, clientID, authTokenURI, authLoginURI, keysetURI, launchURI
func registrationFromEnvironment() datastore.Registration {
	var registration datastore.Registration
	err := envconfig.Process("reg", &registration)
	if err != nil {
		log.Fatalf("registration environment parse error: %v", err)
	}

	return registration
}

// deploymentFromEnvironment loads the deployment details from environment variables.
//
// Variables:
// ('dep_' + ) deploymentID
func deploymentFromEnvironment() datastore.Deployment {
	var deployment datastore.Deployment
	err := envconfig.Process("dep", &deployment)
	if err != nil {
		log.Fatalf("deployment environment parse error: %v", err)
	}

	return deployment
}

type Key struct {
	Private string
}

// keyFromEnvironment loads the key details from environment variables.
//
// Variables:
// ('key_' + ) private
func keyFromEnvironment() Key {
	var key Key
	err := envconfig.Process("key", &key)
	if err != nil {
		log.Fatalf("key environment parse error: %v", err)
	}

	return key
}

func mustNotExist(filename string) {
	if f, err := os.Open(filename); err == nil {
		log.Fatalf("database file already exists (%s)", filename)
		f.Close()
	}
}

func mustPopulateDatabase(db *sql.DB) {
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

	registration := registrationFromEnvironment()

	q := `INSERT INTO registration ( issuer, client_id, auth_token_uri, auth_login_uri, keyset_uri,
                                            target_link_uri )
                   VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(q, registration.Issuer, registration.ClientID, registration.AuthTokenURI.String(),
		registration.AuthLoginURI.String(), registration.KeysetURI.String(), registration.TargetLinkURI.String())
	if err != nil {
		log.Fatalf("cannot populate registration: %v", err)
	}

	deployment := deploymentFromEnvironment()

	q = `INSERT INTO deployment (issuer, deployment_id)
                  VALUES ($1, $2)`
	_, err = db.Exec(q, registration.Issuer, deployment.DeploymentID)
	if err != nil {
		log.Fatalf("cannot populate deployment: %v", err)
	}
}

func sqlDatastoreConfig() datastore.Config {
	// Create and populate an sqlite3 database for testing.
	mustNotExist(sqlite3Database)
	db, err := sql.Open("sqlite3", sqlite3Database)
	if err != nil {
		log.Fatalf("registration database error: %v", err)
	}
	mustPopulateDatabase(db)

	datastoreConfig := lti.NewDatastoreConfig()

	sqlDatastore := lti.NewSQLDatastore(db, lti.NewSQLDatastoreConfig())
	datastoreConfig.Registrations = sqlDatastore

	return datastoreConfig
}

func nonpersistentDatastoreConfig() datastore.Config {
	registration := registrationFromEnvironment()
	err := nonpersistent.DefaultStore.StoreRegistration(registration)
	if err != nil {
		log.Fatalf("registration store error: %v", err)
	}

	deployment := deploymentFromEnvironment()
	err = nonpersistent.DefaultStore.StoreDeployment(registration.Issuer, deployment.DeploymentID)
	if err != nil {
		log.Fatalf("deployment store error: %v", err)
	}

	// The default datastore configuration uses nonpersistent.DefaultStore.
	return lti.NewDatastoreConfig()
}

func postLaunchHandler(datastoreConfig datastore.Config) func(http.ResponseWriter, *http.Request) {
	key := keyFromEnvironment()

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<p>Launch successful!</p>")

		fmt.Fprintf(w, "<p>Launch ID from request context: %s</p>", lti.LaunchIDFromContext(r.Context()))
		fmt.Fprintf(w, "<p>Launch ID from request: %s</p>", lti.LaunchIDFromRequest(r))

		conn, err := connector.New(datastoreConfig, lti.LaunchIDFromRequest(r))
		if err != nil {
			fmt.Println(err)
		}
		conn.SetSigningKey(key.Private)

		nrps, err := conn.UpgradeNRPS()
		if err != nil {
			fmt.Println(err)
		}

		membership, err := nrps.GetMembership()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Fprintf(w, "<p>Members:</p><ul>")
		for _, member := range membership.Members {
			fmt.Fprintf(w, "<li>%s</li>", member.Name)
		}
		fmt.Fprintf(w, "</ul>")
	}
}

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
	var datastoreType = flag.String("datastore", "nonpersistent", "datastore to use")
	flag.Parse()

	var datastoreConfig datastore.Config
	switch *datastoreType {
	case "sqlite3":
		datastoreConfig = sqlDatastoreConfig()
	case "nonpersistent":
		datastoreConfig = nonpersistentDatastoreConfig()
	default:
		log.Fatalf("unsupported datastore (%s)", *datastoreType)
	}

	http.Handle("/login", lti.NewLogin(datastoreConfig))
	http.Handle("/launch", lti.NewLaunch(datastoreConfig,
		postLaunchHandler(datastoreConfig)))

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
