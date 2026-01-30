package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/spf13/viper"
)

// PathConfig holds path settings.
type PathConfig struct {
	BrowserProfile string `mapstructure:"browser_profile"`
	CheckDir       string `mapstructure:"check_dir"`
	DownloadDir    string `mapstructure:"download_dir"`
	IssuerList     string `mapstructure:"issuer_list"`
}

// MailConfig holds mailing settings.
type MailConfig struct {
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	List     []string `mapstructure:"list"`
}

// DownloadConfig holds download settings.
type DownloadConfig struct {
	Year        string `mapstructure:"year"`
	Mode        string `mapstructure:"mode"`
	MonthPeriod string `mapstructure:"month_period"`
}

// DatabaseConfig holds database settings.
type DatabaseConfig struct {
	URI string `mapstructure:"uri"`
}

// Config holds configuration settings.
type Config struct {
	Paths            PathConfig     `mapstructure:"paths"`
	Mail             MailConfig     `mapstructure:"mailing"`
	Download         DownloadConfig `mapstructure:"download"`
	Database         DatabaseConfig `mapstructure:"database"`
	OpenrouterApiKey string         `mapstructure:"openrouter_api_key"`
}

var (
	once     sync.Once
	instance *Config
)

// Load loads and validates config from YAML file.
func Load(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")
		viper.AutomaticEnv() // Allow env overrides

		if e := viper.ReadInConfig(); e != nil {
			err = fmt.Errorf("reading config: %w", e)
			return
		}

		var cfg Config
		if e := viper.Unmarshal(&cfg); e != nil {
			err = fmt.Errorf("unmarshaling config: %w", e)
			return
		}

		// Validate config
		if e := cfg.Validate(); e != nil {
			err = e
			return
		}

		cfg.Paths.DownloadDir = resolvePath(cfg.Paths.DownloadDir)
		cfg.Paths.CheckDir = resolvePath(cfg.Paths.CheckDir)

		instance = &cfg
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func Get() *Config {
	return instance
}

// Validate checks config validity and returns error if invalid.
func (c *Config) Validate() error {
	if c.Paths.IssuerList == "" {
		return fmt.Errorf("issuer_list required")
	}
	if c.Paths.BrowserProfile == "" {
		return fmt.Errorf("browser_profile required")
	}
	if !fileExists(c.Paths.IssuerList) {
		if err := os.WriteFile(c.Paths.IssuerList, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create issuer_list: %w", err)
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

// resolvePath converts relative path to absolute.
func resolvePath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// fileExists checks if file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// isValidEmail validates email format.
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
