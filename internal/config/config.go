package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

// PathConfig holds path-related configuration settings.
type PathConfig struct {
	ChromeDriver string `mapstructure:"chrome_driver"`
	StockList    string `mapstructure:"stock_list"`
	Download     string `mapstructure:"download_dir"`
	Check        string `mapstructure:"check_dir"`
}

// MailConfig holds mailing-related configuration settings.
type MailConfig struct {
	SenderEmail    string   `mapstructure:"sender_email"`
	SenderPassword string   `mapstructure:"sender_password"`
	List           []string `mapstructure:"mailing_list"`
}

// DownloadConfig holds download-related configuration settings.
type DownloadConfig struct {
	Year        string `mapstructure:"year"`
	Mode        string `mapstructure:"mode"`
	MonthPeriod string `mapstructure:"month_period"`
}

// Config holds configuration settings.
type Config struct {
	Paths    PathConfig     `mapstructure:"paths"`
	Mailing  MailConfig     `mapstructure:"mailing"`
	Download DownloadConfig `mapstructure:"download"`
}

// Load reads configuration from a YAML file specified by configPath.
// It uses Viper for configuration management, allowing environment variable overrides.
// The function validates the loaded configuration and resolves relative paths to absolute paths.
// Returns a pointer to Config or an error if loading or validation fails.
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // Allow env overrides

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Validate and resolve paths
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.Paths.Download = resolvePath(cfg.Paths.Download)
	cfg.Paths.Check = resolvePath(cfg.Paths.Check)

	return &cfg, nil
}

// Validate checks the validity of the configuration.
// It ensures required fields are present, paths exist where necessary,
// email addresses are valid, and mode is one of the allowed values.
// Returns an error if validation fails.
func (c *Config) Validate() error {
	if c.Paths.ChromeDriver == "" || !fileExists(c.Paths.ChromeDriver) {
		return fmt.Errorf("invalid chrome_driver_path")
	}
	if c.Paths.StockList == "" || !fileExists(c.Paths.StockList) {
		return fmt.Errorf("invalid stock_list_path")
	}
	if len(c.Mailing.List) == 0 {
		return fmt.Errorf("mailing_list required")
	}
	for _, email := range c.Mailing.List {
		if !isValidEmail(email) {
			return fmt.Errorf("invalid email: %s", email)
		}
	}
	if c.Paths.Download == "" || c.Paths.Check == "" {
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
