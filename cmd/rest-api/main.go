package main

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/anandasatriaadi/go-idx-scraper/internal/api"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/api/option"
)

// FirebaseAuthMiddleware verifies Firebase ID tokens for authentication
func FirebaseAuthMiddleware(authClient *auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}
			// Expect "Bearer <token>"
			const bearerPrefix = "Bearer "
			if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			idToken := authHeader[len(bearerPrefix):]
			_, err := authClient.VerifyIDToken(r.Context(), idToken)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	// Initialize Firebase app
	ctx := context.Background()
	opt := option.WithCredentialsFile("path/to/your/firebase-credentials.json") // Update with actual path
	firebaseApp, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to get Firebase auth client: %v", err)
	}

	// Assume database initialization (replace with actual DB setup)
	db := &mongo.Database{} // Placeholder; implement proper DB connection

	// Set up Chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(FirebaseAuthMiddleware(authClient))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/announcements", api.AnnouncementRoutes(db))
		r.Mount("/financial-reports", api.FinancialReportRoutes(db))

		newsRepo := model.NewNewsRepository(db)
		r.Mount("/news", api.NewsRoutes(newsRepo))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
