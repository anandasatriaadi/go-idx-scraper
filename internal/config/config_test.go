package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create temp config
	yaml := `
paths:
  chrome_driver: /bin/echo
  stock_list: /dev/null
  download_dir: /tmp
  check_dir: /tmp
mailing:
  mailing_list:
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
