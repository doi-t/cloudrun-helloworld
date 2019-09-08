package main

// Sample code reference: https://cloud.google.com/functions/docs/sql#functions_sql_mysql-go

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	// Import the MySQL SQL driver.
	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB

	// MYSQL_INSTANCE_CONNECTION_NAME is PROJECT-ID:REGION:INSTANCE-ID
	connectionName = os.Getenv("MYSQL_INSTANCE_CONNECTION_NAME")
	dbUser         = os.Getenv("MYSQL_USER")
	dbPassword     = os.Getenv("MYSQL_PASSWORD")
	databaseName   = os.Getenv("MYSQL_DATABASE_NAME")

	// Use Cloud SQL Proxy that is automatically activated for a managed Cloud Run
	// Ref. https://cloud.google.com/run/docs/configuring/connect-cloudsql#connect
	dsn = fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s", dbUser, dbPassword, connectionName, databaseName)
)

type entry struct {
	guestName string
	content   string
	entryID   int
}

func init() {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not open db: %v", err)
	}

	// Only allow 1 connection to the database to avoid overloading it.
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
}

// MySQLDemo is an example of making a MySQL database query.
func MySQLDemo(w http.ResponseWriter, r *http.Request) {
	log.Print("MySQLDemo received a request.")
	target := os.Getenv("TARGET")
	if target == "" {
		target = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)

	rows, err := db.Query("SELECT * FROM entries") // Read data from 'entries' table
	if err != nil {
		log.Printf("db.Query: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.guestName, &e.content, &e.entryID); err != nil {
			log.Printf("rows.Scan: %v", err)
			http.Error(w, "Error scanning database", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "guestName: %v, content: %v, entryID: %v\n", e.guestName, e.content, e.entryID)
	}
}

func main() {
	log.Print("MySQLDemo sample started.")

	http.HandleFunc("/", MySQLDemo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
