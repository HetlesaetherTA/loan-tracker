package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	pb "hetlesaether.com/auth/pkg/auth"
	"hetlesaether.com/loan-tracker/internal/api"
)

var HOST string = "0.0.0.0:8002"
var APPLICATION_NAME string = "loan-tracker"
var ROLES = []string{"admin", "user"}

var tmpl = template.Must(template.ParseGlob("web/*.html"))

func main() {
	setupLogging()

	ctx := context.Background()

	auth := setupAuth(ctx)
	routes := setupRouteHandler(ctx)

	rPublic := chi.NewRouter()
	rPublic.Use(middleware.Logger)

	rPublic.Group(func(r chi.Router) {
		// User is sent to login (defined in pb) if client cookie "session_token" is not valid
		r.Use(auth.Middleware(APPLICATION_NAME, ROLES))

		r.Get("/", routes.HandleRoot)

		r.Get("/{loanID}", routes.HandleLoanPage)

		r.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
			res := map[string]any{
				"userID": r.Context().Value("user_id"),
			}

			w.Header().Set("Content-Type", "appliaction/json")
			w.WriteHeader(http.StatusOK)

			if err := json.NewEncoder(w).Encode(res); err != nil {
				http.Error(w, "Failed to encode json", http.StatusInternalServerError)
				return
			}
		})
	})

	http.ListenAndServe(HOST, rPublic)
}

func setupLogging() {
	opts := &slog.HandlerOptions{
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Starting webserver for loans.hetlesaether.com")

}

func setupRouteHandler(ctx context.Context) *api.RouteHandler {
	for {
		routes := api.NewRouteHandler(tmpl)

		if routes != nil {
			slog.Info("Connected to RouteHandler api")
			return routes
		}

		slog.Error("Could not connect to API, trying again in 10 seconds...")

		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			return nil
		}
	}
}

func setupAuth(ctx context.Context) *pb.Client {
	for {
		client, err := pb.NewClient("localhost:50051")

		if err == nil {
			slog.Info("Connected to auth server")
			return client
		}

		slog.Error("Cloud not connect to auth sever (gRPC), trying again in 10 seconds...")

		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			return nil
		}
	}
}
