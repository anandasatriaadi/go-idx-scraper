package main

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	announcementApp "github.com/anandasatriaadi/go-idx-scraper/internal/application/announcement"
	financialreportApp "github.com/anandasatriaadi/go-idx-scraper/internal/application/financialreport"
	newsApp "github.com/anandasatriaadi/go-idx-scraper/internal/application/news"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db"
	persistence "github.com/anandasatriaadi/go-idx-scraper/internal/infrastructure/persistence/mongo"
	presentation "github.com/anandasatriaadi/go-idx-scraper/internal/presentation/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
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
	// Initialize Logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Initialize Config
	// Note: Config loading logic should be consistent, here assuming standard location or just default
	cfg, err := config.Load("config/config.yml") // Ideally pass via arg
	if err != nil {
		logger.Warn("Failed to load config, some features might be disabled", zap.Error(err))
	}

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

	// Initialize Database
	dbClient, err := db.New(logger)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	database := dbClient.GetDatabase("idx")

	// Infrastructure: Repositories
	newsRepo := persistence.NewNewsRepository(database)
	announcementRepo := persistence.NewAnnouncementRepository(database)
	financialRepo := persistence.NewFinancialReportRepository(database)

	// Application: Services
	newsService := newsApp.NewService(newsRepo, logger, cfg)
	announcementService := announcementApp.NewService(announcementRepo, logger)
	financialService := financialreportApp.NewService(financialRepo, logger)

	// Presentation: Handlers
	newsHandler := presentation.NewNewsHandler(newsService)
	announcementHandler := presentation.NewAnnouncementHandler(announcementService)
	financialHandler := presentation.NewFinancialReportHandler(financialService)

	// Set up Chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(FirebaseAuthMiddleware(authClient))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/announcements", announcementHandler.Routes())
		r.Mount("/financial-reports", financialHandler.Routes())
		r.Mount("/news", newsHandler.Routes())
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
