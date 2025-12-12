package helper

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"go.uber.org/zap"
)

func FindDownloadedStocks(config *config.Config) []string {
	var foundStocks []string
	files, err := os.ReadDir(config.Paths.DownloadDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".xlsx" {
			parts := strings.Split(strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())), "-")
			foundStocks = append(foundStocks, parts[len(parts)-1])
		}
	}

	return foundStocks
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Sync to ensure write is complete
	return destFile.Sync()
}

func moveFile(src, dst string) error {
	// Try to rename (move) the file first
	err := os.Rename(src, dst)
	if err == nil {
		return nil // Successful move
	}

	// If rename fails, try copy and delete
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

func MoveFiles(logger *zap.Logger, config *config.Config) error {
	files, err := filepath.Glob(filepath.Join(config.Paths.DownloadDir, "*.xlsx"))
	if err != nil {
		return fmt.Errorf("Error getting files: %v", err)
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		destPath := filepath.Join(config.Paths.CheckDir, fileName)
		err := moveFile(file, destPath)
		if err != nil {
			logger.Info("Error moving file", zap.String("file", file), zap.Error(err))
		} else {
			logger.Info("Successfully moved file", zap.String("file", fileName))
		}
	}

	return nil
}
