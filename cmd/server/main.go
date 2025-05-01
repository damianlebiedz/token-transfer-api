package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/damianlebiedz/token-transfer-api/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/lib/pq"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database: ", err)
	}

	log.Println("Connected to the database successfully!")

	resolver := &graph.Resolver{}

	schema := graph.NewExecutableSchema(graph.Config{Resolvers: resolver})
	srv := handler.New(schema)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
