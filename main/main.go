package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cataclyst/archivist-server/graphql"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const sqliteDateFormatIso8601 = "2006-01-02 15:04:05Z07:00"

func main() {

	fmt.Println("Connecting to SQLite database")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Starting GraphQL server...")
	go graphql.StartGraphQlServer(0, db)

	fmt.Println("Starting API server")

	insertTestData(db)

	archivistHandler := &ArchivistHandler{}

	server := &http.Server{
		Addr:           ":8080",
		Handler:        archivistHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}

func insertTestData(db *sql.DB) {

	sqlStmt := `create table if not exists Document (id text not null primary key, title text not null, description text, date text not null, created_at text not null, modified_at text not null);`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into Document(id, title, description, date, created_at, modified_at) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 5; i++ {
		randomTime := asDatabaseTime(time.Now().Add(-time.Duration(rand.Intn(100)) * time.Hour))
		_, err = stmt.Exec(
			uuid.New().String(),
			"Some title " + strconv.Itoa(i),
			"This is a description " + strconv.Itoa(i),
			randomTime, randomTime, randomTime)
		
		if err != nil {
			log.Fatal(err)
		}
		log.Println("One row added")
	}
	tx.Commit()
}

func asDatabaseTime(input time.Time) string {
	return input.UTC().Format(sqliteDateFormatIso8601)
}

type ArchivistHandler struct{}

func (*ArchivistHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	time.Sleep(1 * time.Second)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var arrayResult []ArchivistDocument

	for i := 0; i <= 10; i++ {
		arrayResult = append(arrayResult, ArchivistDocument{
			ID:          "abcdef" + strconv.Itoa(i),
			Name:        "Some name " + strconv.Itoa(i),
			Date:        time.Now(),
			Description: "Some desc",
			Labels:      []string{"label1", "label2"},
		})
	}

	result, err := json.Marshal(arrayResult)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(result))
}

type ArchivistDocument struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Labels      []string  `json:"labels"`
}
