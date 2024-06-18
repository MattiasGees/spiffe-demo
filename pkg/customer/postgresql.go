/*
Copyright © 2024 Mattias Gees mattias.gees@venafi.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package customer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

const (
	dbUser = "spiffe-customer"
	dbName = "testdb"
	dbPort = "5432"
)

var firstNames = []string{
	"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
	"William", "Elizabeth", "David", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Jones", "Brown", "Davis", "Miller", "Wilson",
	"Moore", "Taylor", "Anderson", "Thomas", "Jackson", "White", "Harris", "Martin",
}

func (c *CustomerService) postgreSQLRetrievalHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the PostgreSQL Retrieval handler from %s", r.RemoteAddr)

	db, err := c.setupPostgreSQLConnection()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	queryStmt := `SELECT name, text FROM test_table`
	rows, err := db.Query(queryStmt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying data: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprint(w, "<p>The following data has been retrieved from PostgreSQL</p>")

	for rows.Next() {
		var name string
		var text string
		err := rows.Scan(&name, &text)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "<p>Name: %s, Text: %s</p>", name, text)
	}

	err = rows.Err()
	if err != nil {
		log.Fatalf("Error iterating rows: %v", err)
		return
	}
}

func (c *CustomerService) postgreSQLPutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the PostgreSQL Put handler from %s", r.RemoteAddr)

	db, err := c.setupPostgreSQLConnection()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)
	firstName := firstNames[randGen.Intn(len(firstNames))]
	lastName := lastNames[randGen.Intn(len(lastNames))]
	fullName := fmt.Sprintf("%s %s", firstName, lastName)

	currentTime := time.Now()
	formattedTime := currentTime.Format("02/01/06 15:04:05")
	text := fmt.Sprintf("The time is %s", formattedTime)

	insertStmt := `INSERT INTO test_table (name, text) VALUES ($1, $2)`
	_, err = db.Exec(insertStmt, fullName, text)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<p>Data inserted successfully! Used the following details</p>")
	fmt.Fprintf(w, "<p>Name: %s, Text: %s</p>", fullName, text)

}

func (c *CustomerService) setupPostgreSQLConnection() (*sql.DB, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket path
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create X509Source: %v", err)
	}
	defer source.Close()

	tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny())

	connStr := fmt.Sprintf(
		"postgres://%s@%s:%s/%s?sslmode=require",
		dbUser, c.postgreSQLHost, dbPort, dbName)

	config, err := pgx.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse connection string: %v", err)
	}

	config.TLSConfig = tlsConfig

	db := stdlib.OpenDB(*config)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return db, nil
}
