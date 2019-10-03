package graphql

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/handler"
	"github.com/rs/cors"
)

const defaultPort = 9090

// StartGraphQlServer starts a GraphQL server on the given port
// and blocks until the server terminates, in which case the
// Go routine panics.
func StartGraphQlServer(port int, db *sql.DB) {
	if port == 0 {
		port = defaultPort
	}

	playgroundHandler := handler.Playground("GraphQL playground", "/query")
	queryHandler := handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{
		db: db,
	}}))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	mux := http.NewServeMux()
	mux.Handle("/", playgroundHandler)
	mux.Handle("/query", queryHandler)

	// http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	// http.Handle("/query", handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{
	// 	db:db,
	// }})))

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), corsHandler.Handler(mux)))
}
