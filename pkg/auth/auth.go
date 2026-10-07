package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	pb "hetlesaether.com/loan-tracker/pkg/auth/proto"
)

type Client struct {
	conn pb.AuthServiceClient
}

func NewClient(grpcTarget string) (*Client, error) {
	conn, err := grpc.NewClient(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{pb.NewAuthServiceClient(conn)}, nil
}

func (c *Client) Middleware(application string, authorizedRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 150*time.Millisecond)
			defer cancel()

			res, err := c.conn.ValidateSession(ctx, &pb.SessionRequest{
				SessionToken:    cookie.Value,
				Appliaction:     application,
				AuthorizedRoles: authorizedRoles,
			})
			if err != nil {
				status, ok := status.FromError(err)

				if !ok {
					slog.Error("Failed to authenticate session",
						"error", err,
						"reason", "service offline?")

					http.Error(w, "Auth service offline...", http.StatusServiceUnavailable)
					return
				}

				switch status.Code() {
				case codes.NotFound:
					redirectToLogin(w, r)
					return
				default:
					slog.Error("Failed to authenticate session", "error", err)
					http.Error(w, status.Message(), http.StatusInternalServerError)
				}

			}

			newCtx := context.WithValue(r.Context(), "user_id", res.UserId)

			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}

func setRedirectURLCookie(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "redirect_url",
		Value:  r.Host,
		Path:   "/",
		Domain: ".hetlesaether.com",
		// Domain:     "",
		Expires:    time.Time{},
		RawExpires: "",
		MaxAge:     1800, // 30 minutes
		// Secure:     true,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	loginURL := "https://auth.hetlesaether.com/login"

	setRedirectURLCookie(w, r)

	isFetch := r.Header.Get("Sec-Fetch-Mode") == "cors" ||
		r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		r.Header.Get("Accept") == "application/json"

	if isFetch {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized, redirecting to login"}`))
		return
	}

	http.Redirect(w, r, loginURL, http.StatusSeeOther)
}
