package excel

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ProcessDownloadedFiles processes Excel files.
func ProcessDownloadedFiles(sourcePath, destPath string) error {
	// Remove .crdownload files
	filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ".crdownload" {
			os.Remove(path)
		}
		return nil
	})

	files, err := filepath.Glob(filepath.Join(sourcePath, "*.xlsx"))
	if err != nil {
		return err
	}
	for _, file := range files {
		processFile(file)
	}
	return nil
}

func processFile(filePath string) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		slog.Error("Opening file", "file", filePath, "error", err)
		return
	}
	defer f.Close()

	for _, sheet := range f.GetSheetList() {
		cell, _ := f.GetCellValue(sheet, "A3")
		cell = strings.ToLower(strings.TrimSpace(cell))
		switch {
		case cell == "hidden":
			// Skip
		case strings.Contains(cell, "laporan arus kas"):
			f.SetSheetName(sheet, "CashFlow")
		case strings.Contains(cell, "laporan posisi keuangan"):
			f.SetSheetName(sheet, "Neraca")
		case strings.Contains(cell, "laporan laba rugi"):
			f.SetSheetName(sheet, "RugiLaba")
		case strings.Contains(cell, "informasi umum"):
			f.SetSheetName(sheet, "InfoUmum")
		default:
			f.DeleteSheet(sheet)
		}
	}
	f.SaveAs(filePath)
}
