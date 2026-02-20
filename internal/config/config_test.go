package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create temp config
	yaml := `
paths:
  browser_profile: /tmp/browser
  issuer_list: /dev/null
  download_dir: /tmp
  check_dir: /tmp
mailing:
  list:
    - test@example.com
download:
  year: 2023
  mode: TW
  month_period: 3`
	file, _ := os.CreateTemp("", "config.yaml")
	defer os.Remove(file.Name())
	file.WriteString(yaml)
	file.Close()

	cfg, err := Load(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Download.Mode != "TW" {
		t.Errorf("expected TW, got %s", cfg.Download.Mode)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Paths: PathConfig{
					IssuerList:     "/dev/null",
					BrowserProfile: "/tmp",
					DownloadDir:    "/tmp",
					CheckDir:       "/tmp",
				},
				Mail: MailConfig{
					List: []string{"test@example.com"},
				},
				Download: DownloadConfig{
					Year: "2023",
					Mode: "TW",
				},
			},
			wantErr: false,
		},
		{
			name: "missing issuer_list",
			config: Config{
				Paths: PathConfig{
					BrowserProfile: "/tmp",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			config: Config{
				Paths: PathConfig{
					IssuerList:     "/dev/null",
					BrowserProfile: "/tmp",
					DownloadDir:    "/tmp",
					CheckDir:       "/tmp",
				},
				Mail: MailConfig{
					List: []string{"invalid-email"},
				},
				Download: DownloadConfig{
					Year: "2023",
					Mode: "TW",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			config: Config{
				Paths: PathConfig{
					IssuerList:     "/dev/null",
					BrowserProfile: "/tmp",
					DownloadDir:    "/tmp",
					CheckDir:       "/tmp",
				},
				Mail: MailConfig{
					List: []string{"test@example.com"},
				},
				Download: DownloadConfig{
					Year: "2023",
					Mode: "INVALID",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
