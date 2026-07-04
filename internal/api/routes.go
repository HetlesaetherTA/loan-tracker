package api

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"hetlesaether.com/loan-tracker/internal/database"
)

type RouteHandler struct {
	tmpl    *template.Template
	queries *database.Queries
}

func NewRouteHandler(tmpl *template.Template) *RouteHandler {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		slog.Error("Could not connect to database", "error", err)
		return nil
	}

	err = pool.Ping(ctx)

	if err != nil {
		slog.Error("Ping to database failed", "error", err)
		return nil
	}

	return &RouteHandler{
		tmpl,
		database.New(pool),
	}
}

func (h *RouteHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	userUUID := getUserID(w, r)

	res, err := h.queries.GetUserLoans(ctx, pgtype.UUID{
		Bytes: userUUID,
		Valid: true,
	})

	if err != nil {
		slog.Error("Failed to get loans from DB", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	loans := make([]Loan, 0, len(res))

	for _, v := range res {
		loans = append(loans, parseLoan(v))
	}

	w.WriteHeader(http.StatusOK)

	h.tmpl.ExecuteTemplate(w, "home.html", loans)
}

func (h *RouteHandler) HandleLoanPage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	userUUID := getUserID(w, r)

	// Extract loanID from body
	loanID := chi.URLParam(r, "loanID")

	// Convert loanID to UUID
	loanUUID, err := stringToUUID(loanID)

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get selected loan from DB only if user owns the loan
	res, err := h.queries.GetUserLoanByID(ctx, database.GetUserLoanByIDParams{
		ID: pgtype.UUID{
			Bytes: loanUUID,
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		slog.Error("Failed to get loan from DB", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	loan := parseLoan(res)

	w.WriteHeader(http.StatusOK)
	h.tmpl.ExecuteTemplate(w, "loan.html", loan)
}

func getUserID(w http.ResponseWriter, r *http.Request) [16]byte {
	ctx := r.Context()
	value := ctx.Value("user_id")
	userID, ok := value.(string)

	if !ok || userID == "" {
		slog.Warn("user_id missing or invalid type in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return [16]byte{}
	}

	userUUID, err := stringToUUID(userID)

	if err != nil {
		slog.Error("Failed to parse UUID from context string", "userID", userID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return [16]byte{}
	}

	return userUUID
}
