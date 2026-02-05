package cmd

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/finreport"
	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/news"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	handlers "github.com/anandasatriaadi/go-idx-scraper/internal/presentation/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

var ServerCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

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

func startServer() {
	// Initialize Logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Initialize Config
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Warn("Failed to load config, some features might be disabled", zap.Error(err))
	}

	// Initialize Firebase app
	ctx := context.Background()
	opt := option.WithCredentialsFile("path/to/your/firebase-credentials.json") // Update with actual path or config
	firebaseApp, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to get Firebase auth client: %v", err)
	}

	// Initialize Database
	dbClient, err := mongo.NewClient(logger)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	database := dbClient.Database(viper.GetString("database.db_names"))

	// Infrastructure: Repositories
	newsRepo := mongo.NewNewsRepository(database)
	announcementRepo := mongo.NewAnnouncementRepository(database)
	financialRepo := mongo.NewFinancialReportRepository(database)

	// Application: Services
	newsService := news.NewService(newsRepo, logger, cfg)
	announcementService := announcement.NewService(announcementRepo, logger)
	financialService := finreport.NewService(financialRepo, logger)

	// Presentation: Handlers
	newsHandler := handlers.NewNewsHandler(newsService)
	announcementHandler := handlers.NewAnnouncementHandler(announcementService)
	financialHandler := handlers.NewFinancialReportHandler(financialService)

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
