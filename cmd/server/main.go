package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"log"
	"net/http"
	"os"

	"token-transfer-api/graph"
	"token-transfer-api/internal/db"
)

const defaultPort = "8080"

func main() {
	db.Init()

	// Get the port from environment variable or use the default one
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Set up the GraphQL schema with resolvers
	resolver := &graph.Resolver{}
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: resolver})
	srv := handler.New(schema)

	// Enable standard HTTP transports (OPTIONS, GET, POST)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// Enable introspection for the GraphQL Playground
	srv.Use(extension.Introspection{})

	// Set up a Playground
	http.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Connect to http://localhost:%s/playground for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
