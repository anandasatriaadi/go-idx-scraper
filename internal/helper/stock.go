package helper

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

// LoadCurrent loads stocks from file.
func LoadCurrent(filePath string, logger *zap.Logger) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err2 := os.WriteFile(filePath, []byte("[]"), 0644)
			if err2 != nil {
				logger.Error("Creating stocks file", zap.Error(err2))
				return nil, err2
			}
			return []string{}, nil
		}
		logger.Error("Loading stocks", zap.Error(err))
		return nil, err
	}
	var stocks []string
	if err := json.Unmarshal(data, &stocks); err != nil {
		logger.Error("Unmarshaling stocks", zap.Error(err))
		return nil, err
	}
	return stocks, nil
}
