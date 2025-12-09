package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// PathConfig holds path-related configuration settings.
type PathConfig struct {
	BrowserProfile string `mapstructure:"browser_profile"`
	CheckDir       string `mapstructure:"check_dir"`
	DownloadDir    string `mapstructure:"download_dir"`
	StockList      string `mapstructure:"stock_list"`
}

// MailConfig holds mailing-related configuration settings.
type MailConfig struct {
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	List     []string `mapstructure:"list"`
}

// DownloadConfig holds download-related configuration settings.
type DownloadConfig struct {
	Year        string `mapstructure:"year"`
	Mode        string `mapstructure:"mode"`
	MonthPeriod string `mapstructure:"month_period"`
}

// DatabaseConfig holds database-related configuration settings.
type DatabaseConfig struct {
	URI string `mapstructure:"uri"`
}

// Config holds configuration settings.
type Config struct {
	Paths    PathConfig     `mapstructure:"paths"`
	Mail     MailConfig     `mapstructure:"mailing"`
	Download DownloadConfig `mapstructure:"download"`
	Database DatabaseConfig `mapstructure:"database"`
}

var (
	once     sync.Once
	instance *Config
)

// Load reads configuration from a YAML file specified by configPath.
// It uses Viper for configuration management, allowing environment variable overrides.
// The function validates the loaded configuration and resolves relative paths to absolute paths.
// Returns a pointer to Config or an error if loading or validation fails.
func Load(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		zap.L().Info("Starting to load configuration", zap.String("configPath", configPath))

		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")
		viper.AutomaticEnv() // Allow env overrides

		if e := viper.ReadInConfig(); e != nil {
			zap.L().Error("Failed to read config", zap.Error(e))
			err = fmt.Errorf("reading config: %w", e)
			return
		}
		zap.L().Info("Successfully read config file")

		var cfg Config
		if e := viper.Unmarshal(&cfg); e != nil {
			zap.L().Error("Failed to unmarshal config", zap.Error(e))
			err = fmt.Errorf("unmarshaling config: %w", e)
			return
		}
		zap.L().Info("Successfully unmarshaled config")

		// Validate and resolve paths
		if e := cfg.Validate(); e != nil {
			zap.L().Error("Config validation failed", zap.Error(e))
			err = e
			return
		}
		zap.L().Info("Config validation passed")

		cfg.Paths.DownloadDir = resolvePath(cfg.Paths.DownloadDir)
		cfg.Paths.CheckDir = resolvePath(cfg.Paths.CheckDir)
		zap.L().Info("Resolved paths", zap.String("downloadDir", cfg.Paths.DownloadDir), zap.String("checkDir", cfg.Paths.CheckDir))

		instance = &cfg
		zap.L().Info("Configuration loaded successfully")
	})
	if err != nil {
		zap.L().Error("Failed to load configuration", zap.Error(err))
		return nil, err
	}
	return instance, nil
}

func Get() *Config {
	return instance
}

// Validate checks the validity of the configuration.
// It ensures required fields are present, paths exist where necessary,
// email addresses are valid, and mode is one of the allowed values.
// Returns an error if validation fails.
func (c *Config) Validate() error {
	if c.Paths.StockList == "" {
		return fmt.Errorf("stock_list_path required")
	}
	if c.Paths.BrowserProfile == "" {
		return fmt.Errorf("browser_profile required")
	}
	if !fileExists(c.Paths.StockList) {
		if err := os.WriteFile(c.Paths.StockList, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create stock_list: %w", err)
		}
	}
	if len(c.Mail.List) == 0 {
		return fmt.Errorf("mailing_list required")
	}
	for _, email := range c.Mail.List {
		if !isValidEmail(email) {
			return fmt.Errorf("invalid email: %s", email)
		}
	}
	if c.Paths.DownloadDir == "" || c.Paths.CheckDir == "" {
		return fmt.Errorf("download/check paths required")
	}
	if c.Download.Year == "" {
		return fmt.Errorf("download_year required")
	}
	if c.Download.Mode != "TW" && c.Download.Mode != "AUDIT" {
		return fmt.Errorf("invalid download_mode")
	}
	return nil
}

// resolvePath converts a relative path to an absolute path.
// If conversion fails, it returns the original path.
func resolvePath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// fileExists checks if a file exists at the given path.
// Returns true if the file exists, false otherwise.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// isValidEmail validates an email address using a regular expression.
// Returns true if the email matches the pattern, false otherwise.
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
