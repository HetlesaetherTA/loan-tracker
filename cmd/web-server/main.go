package main

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	pb "hetlesaether.com/auth/pkg/auth"
)

var HOST string = "0.0.0.0:8002"
var APPLICATION_NAME string = "loan-tracker"
var ROLES = []string{"admin", "user"}

var templates = template.Must(template.ParseGlob("web/*.html"))

func main() {
	opts := &slog.HandlerOptions{
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Starting webserver for loans.hetlesaether.com")
	auth, err := pb.NewClient("localhost:50051")


	if err != nil {
		slog.Error("Cloud not connect to auth sever (gRPC)")
	}
	
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// -- Serve /
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "home.html", nil)
	})

	r.Group(func(subRouter chi.Router) {
		subRouter.Use(auth.Middleware(APPLICATION_NAME, ROLES))

		subRouter.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
			if err != nil {
				return
			}


			res := map[string]any {
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


	http.ListenAndServe(HOST, r)
}
