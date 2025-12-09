package db

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

type cfg struct {
	URI                    string
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
	MaxPoolSize            uint64
	MinPoolSize            uint64
	MaxConnIdleTime        time.Duration
	HeartbeatInterval      time.Duration
	RetryWrites            bool
	RetryReads             bool
	Compressors            []string
	TLSConfig              *tls.Config
	AppName                string
}

// DefaultConfig returns a production-ready config with sensible defaults.
// Override via env vars or struct fields.
func DefaultConfig() *cfg {
	config := config.Get()
	if config == nil {
		zap.L().Panic("Config not loaded")
	}
	return &cfg{
		URI:                    config.Database.URI,
		ConnectTimeout:         10 * time.Second,
		ServerSelectionTimeout: 5 * time.Second,
		MaxPoolSize:            100,
		MinPoolSize:            10,
		MaxConnIdleTime:        30 * time.Second,
		HeartbeatInterval:      10 * time.Second,
		RetryWrites:            true,
		RetryReads:             true,
		Compressors:            []string{"snappy"},
		AppName:                "go-idx-scraper",
	}
}

// MongoDB wraps the MongoDB client for better encapsulation and testing.
type MongoDB struct {
	client *mongo.Client
	logger *zap.Logger
}

var instance *MongoDB
var instanceErr error
var once sync.Once

// New creates and connects a new MongoDB instance.
// Call Close() to disconnect. Use in main or as a dependency.
func New() (*MongoDB, error) {
	once.Do(func() {
		cfg := DefaultConfig()
		logger := zap.L() // Fallback to default logger

		// Build client options from config
		clientOpts := options.Client().
			ApplyURI(cfg.URI).
			SetConnectTimeout(cfg.ConnectTimeout).
			SetServerSelectionTimeout(cfg.ServerSelectionTimeout).
			SetMaxPoolSize(cfg.MaxPoolSize).
			SetMinPoolSize(cfg.MinPoolSize).
			SetMaxConnIdleTime(cfg.MaxConnIdleTime).
			SetHeartbeatInterval(cfg.HeartbeatInterval).
			SetRetryWrites(cfg.RetryWrites).
			SetRetryReads(cfg.RetryReads).
			SetCompressors(cfg.Compressors).
			SetAppName(cfg.AppName)

		if cfg.TLSConfig != nil {
			clientOpts.SetTLSConfig(cfg.TLSConfig)
		}

		// Connect
		client, err := mongo.Connect(clientOpts)
		if err != nil {
			logger.Error("Failed to connect to MongoDB", zap.Error(err))
			instanceErr = err
			return
		}

		mdb := &MongoDB{client: client, logger: logger}

		// Verify connection
		if err := mdb.HealthCheck(context.Background()); err != nil {
			logger.Error("MongoDB health check failed", zap.Error(err))
			client.Disconnect(context.Background()) // Cleanup on failure
			instanceErr = err
			return
		}

		logger.Info("Successfully connected to MongoDB", zap.String("uri", cfg.URI))
		instance = mdb
	})
	return instance, instanceErr
}

// HealthCheck pings MongoDB to verify connectivity.
// Use for health endpoints (e.g., /health in HTTP servers).
func (m *MongoDB) HealthCheck(ctx context.Context) error {
	if m.client == nil {
		return errors.New("MongoDB client not initialized")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := m.client.Ping(pingCtx, readpref.Primary()); err != nil {
		m.logger.Error("MongoDB ping failed", zap.Error(err))
		return err
	}
	return nil
}

// Client returns the underlying MongoDB client for advanced operations.
// Use sparingly; prefer GetCollection for common tasks.
func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

// GetCollection returns a collection from the specified database.
// Convenience method for CRUD operations.
func (m *MongoDB) GetCollection(databaseName, collectionName string) *mongo.Collection {
	if m.client == nil {
		m.logger.Error("MongoDB client not initialized")
		return nil
	}
	return m.client.Database(databaseName).Collection(collectionName)
}

// GetDatabase returns a database handle.
// Useful for operations spanning multiple collections.
func (m *MongoDB) GetDatabase(databaseName string) *mongo.Database {
	if m.client == nil {
		m.logger.Error("MongoDB client not initialized")
		return nil
	}
	return m.client.Database(databaseName)
}

// Close disconnects the MongoDB client gracefully.
// Call with a context (e.g., from signal handling) for production shutdown.
func (m *MongoDB) Close(ctx context.Context) error {
	if m.client == nil {
		return nil
	}
	closeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := m.client.Disconnect(closeCtx)
	m.client = nil
	if err != nil {
		m.logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		return err
	}
	m.logger.Info("Disconnected from MongoDB")
	return nil
}
