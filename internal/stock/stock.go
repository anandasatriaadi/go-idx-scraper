package stock

import (
	"encoding/json"
	"log/slog"
	"os"
)

// LoadCurrent loads stocks from file.
func LoadCurrent(filePath string) []string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err2 := os.WriteFile(filePath, []byte("[]"), 0644)
			if err2 != nil {
				slog.Error("Creating stocks file", "error", err2)
				return nil
			}
			return []string{}
		}
		slog.Error("Loading stocks", "error", err)
		return nil
	}
	var stocks []string
	json.Unmarshal(data, &stocks)
	return stocks
}
