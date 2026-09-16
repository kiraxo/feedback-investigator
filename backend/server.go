package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kiraxo/feedback-investigator/backend/graph"
	"github.com/kiraxo/feedback-investigator/backend/internal/agent"
	"github.com/kiraxo/feedback-investigator/backend/internal/storage"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	ctx := context.Background()

	var repository storage.RunRepository = storage.NewMemoryRepository()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL != "" {
		postgresRepository, err :=
			storage.NewPostgresRepository(ctx, databaseURL)
		if err != nil {
			log.Fatalf(
				"initialize PostgreSQL repository: %v",
				err,
			)
		}

		repository = postgresRepository
		log.Print("storage: PostgreSQL")
	} else {
		log.Print("storage: in-memory demo repository")
	}

	agentService := agent.NewService(repository)
	defer agentService.Close()

	resolver := graph.NewResolver(agentService)

	srv := handler.New(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: resolver,
			},
		),
	)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(
		lru.New[*ast.QueryDocument](1000),
	)

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle(
		"/",
		playground.Handler(
			"Feedback Investigator GraphQL",
			"/query",
		),
	)
	http.Handle("/query", srv)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf(
		"GraphQL playground: http://localhost:%s/",
		port,
	)

	log.Fatal(
		http.ListenAndServe(":"+port, nil),
	)
}
