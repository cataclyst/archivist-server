package graphql

import (
	"log"
	"net/http"
	"strconv"
	"database/sql"

	"github.com/99designs/gqlgen/handler"
)

const defaultPort = 9090

// StartGraphQlServer starts a GraphQL server on the given port
// and blocks until the server terminates, in which case the
// Go routine panics.
func StartGraphQlServer(port int, db *sql.DB) {
	if port == 0 {
		port = defaultPort
	}

	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	http.Handle("/query", handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{
		db:db,
	}})))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
