package xbrl

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/xuri/excelize/v2"
)

// ParseAnyFiling automatically detects and parses .zip, .xbrl, .xml, or .xlsx filings
func ParseAnyFiling(filePath string) (*domain.Statement, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".zip":
		return ParseInstanceZip(filePath)
	case ".xbrl", ".xml":
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("opening xbrl file %s: %w", filePath, err)
		}
		defer f.Close()
		stmt, err := ParseInstanceXML(f)
		if err != nil {
			return nil, err
		}
		stmt.Metadata.SourceFile = filepath.Base(filePath)
		return stmt, nil
	case ".xlsx":
		return ParseExcelStatement(filePath)
	default:
		return nil, fmt.Errorf("unsupported filing file extension %s for %s", ext, filePath)
	}
}

// ParseExcelStatement parses standard IDX XBRL-derived Excel workbooks
func ParseExcelStatement(filePath string) (*domain.Statement, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening excel file %s: %w", filePath, err)
	}
	defer f.Close()

	stmt := &domain.Statement{
		Metadata: domain.StatementMetadata{
			RoundingMultiplier: 1.0,
			SourceFile:         filepath.Base(filePath),
		},
		Facts: make(domain.FactMap),
	}

	// Parse InfoUmum / Sheet 1000000
	parseExcelGeneralInfo(f, stmt)

	// Parse Balance Sheet / Neraca / Sheet 1110000 / 1210000
	parseExcelBalanceSheet(f, stmt)

	// Parse Income Statement / RugiLaba / Sheet 1311000 / 1321000
	parseExcelIncomeStatement(f, stmt)

	// Parse Cash Flow / Sheet 1510000 / CashFlow
	parseExcelCashFlow(f, stmt)

	finalizeCoreFinancials(stmt)

	return stmt, nil
}

func parseExcelGeneralInfo(f *excelize.File, stmt *domain.Statement) {
	sheets := f.GetSheetList()
	var targetSheet string
	for _, s := range sheets {
		nameLower := strings.ToLower(s)
		if strings.Contains(nameLower, "1000000") || strings.Contains(nameLower, "infoumum") || strings.Contains(nameLower, "general") {
			targetSheet = s
			break
		}
	}
	if targetSheet == "" && len(sheets) > 0 {
		targetSheet = sheets[0]
	}

	rows, err := f.GetRows(targetSheet)
	if err != nil {
		return
	}

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(row[0]))
		val := ""
		if len(row) >= 2 {
			val = strings.TrimSpace(row[1])
		}
		if val == "" && len(row) >= 3 {
			val = strings.TrimSpace(row[2])
		}

		switch {
		case strings.Contains(key, "kode entitas") || strings.Contains(key, "entity code") || strings.Contains(key, "ticker"):
			if stmt.Ticker == "" {
				stmt.Ticker = strings.ToUpper(val)
			}
		case strings.Contains(key, "nama entitas") || strings.Contains(key, "entity name"):
			if stmt.CompanyName == "" {
				stmt.CompanyName = val
			}
		case strings.Contains(key, "sektor") || strings.Contains(key, "sector"):
			if stmt.Metadata.Sector == "" {
				stmt.Metadata.Sector = val
			}
		case strings.Contains(key, "industri") || strings.Contains(key, "industry"):
			if stmt.Metadata.Industry == "" {
				stmt.Metadata.Industry = val
			}
		case strings.Contains(key, "mata uang") || strings.Contains(key, "currency"):
			if strings.Contains(strings.ToUpper(val), "USD") || strings.Contains(strings.ToLower(val), "dollar") {
				stmt.Metadata.Currency = "USD"
			} else {
				stmt.Metadata.Currency = "IDR"
			}
		case strings.Contains(key, "pembulatan") || strings.Contains(key, "rounding"):
			stmt.Metadata.RoundingLevel = val
			if strings.Contains(strings.ToLower(val), "thousand") || strings.Contains(strings.ToLower(val), "ribu") {
				stmt.Metadata.RoundingMultiplier = 1000
			} else if strings.Contains(strings.ToLower(val), "million") || strings.Contains(strings.ToLower(val), "juta") {
				stmt.Metadata.RoundingMultiplier = 1000000
			} else if strings.Contains(strings.ToLower(val), "billion") || strings.Contains(strings.ToLower(val), "miliar") {
				stmt.Metadata.RoundingMultiplier = 1000000000
			}
		case strings.Contains(key, "kurs konversi") || strings.Contains(key, "conversion rate"):
			rate, _ := parseNumericValue(val)
			stmt.Metadata.ConversionRate = rate
		case strings.Contains(key, "periode laporan") || strings.Contains(key, "period of financial"):
			stmt.PeriodType = val
			if strings.Contains(val, "Kuartal I") || strings.Contains(val, "First") {
				stmt.Period = "Q1"
			} else if strings.Contains(val, "Kuartal II") || strings.Contains(val, "Second") {
				stmt.Period = "Q2"
			} else if strings.Contains(val, "Kuartal III") || strings.Contains(val, "Third") {
				stmt.Period = "Q3"
			} else {
				stmt.Period = "FY"
			}
		case (strings.Contains(key, "tanggal akhir periode tahun berjalan") || strings.Contains(key, "current period end date") || strings.Contains(key, "tanggal akhir periode")) && !strings.Contains(key, "sebelumnya") && !strings.Contains(key, "prior") && !strings.Contains(key, "lalu"):
			if t, err := time.Parse("2006-01-02", val); err == nil {
				stmt.PeriodEndDate = t
				stmt.Year = t.Year()
			}
		}
	}

	// Fallback to filename parsing if Ticker/Year are still missing
	if stmt.Ticker == "" || stmt.Year == 0 {
		base := filepath.Base(stmt.Metadata.SourceFile)
		parts := strings.Split(base, "-")
		if len(parts) >= 4 {
			if y, err := strconv.Atoi(parts[1]); err == nil {
				stmt.Year = y
			}
			stmt.Period = parts[2]
			stmt.Ticker = strings.ToUpper(strings.TrimSuffix(parts[3], filepath.Ext(parts[3])))
		}
	}
}

func parseExcelBalanceSheet(f *excelize.File, stmt *domain.Statement) {
	sheet := findMatchingSheet(f, []string{"1110000", "1210000", "neraca", "balance"})
	if sheet == "" {
		return
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return
	}

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		label := strings.ToLower(strings.TrimSpace(row[0]))
		val := findFirstNumericCell(row)

		switch {
		case (strings.Contains(label, "jumlah aset") || strings.Contains(label, "total assets")) && !strings.Contains(label, "lancar") && !strings.Contains(label, "current"):
			stmt.Core.TotalAssets = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "kas dan setara kas") || strings.Contains(label, "cash and cash equivalents"):
			if stmt.Core.CashAndEquivalents == 0 {
				stmt.Core.CashAndEquivalents = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "jumlah aset lancar") || strings.Contains(label, "total current assets"):
			stmt.Core.CurrentAssets = val * stmt.Metadata.RoundingMultiplier
		case (strings.Contains(label, "jumlah liabilitas") || strings.Contains(label, "total liabilities")) && !strings.Contains(label, "jangka") && !strings.Contains(label, "current"):
			stmt.Core.TotalLiabilities = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "jumlah liabilitas jangka pendek") || strings.Contains(label, "total current liabilities"):
			stmt.Core.CurrentLiabilities = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "utang bank jangka pendek") || strings.Contains(label, "short-term bank loans"):
			stmt.Core.ShortTermDebt = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "utang bank jangka panjang") || strings.Contains(label, "long-term bank loans"):
			stmt.Core.LongTermDebt = val * stmt.Metadata.RoundingMultiplier
		case (strings.Contains(label, "jumlah ekuitas") || strings.Contains(label, "total equity")):
			if stmt.Core.TotalEquity == 0 {
				stmt.Core.TotalEquity = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "saldo laba yang belum dicadangkan") || strings.Contains(label, "unappropriated retained earnings"):
			stmt.Core.RetainedEarnings = val * stmt.Metadata.RoundingMultiplier
		}
	}
}

func parseExcelIncomeStatement(f *excelize.File, stmt *domain.Statement) {
	sheet := findMatchingSheet(f, []string{"1311000", "1321000", "rugilaba", "income", "profit"})
	if sheet == "" {
		return
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return
	}

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		label := strings.ToLower(strings.TrimSpace(row[0]))
		val := findFirstNumericCell(row)

		switch {
		case strings.Contains(label, "penjualan dan pendapatan") || strings.Contains(label, "sales and revenue") || strings.Contains(label, "pendapatan bunga"):
			if stmt.Core.Revenue == 0 {
				stmt.Core.Revenue = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "beban pokok penjualan") || strings.Contains(label, "cost of sales and revenue"):
			stmt.Core.CostOfRevenue = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "jumlah laba bruto") || strings.Contains(label, "gross profit"):
			stmt.Core.GrossProfit = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "laba (rugi) usaha") || strings.Contains(label, "operating profit") || strings.Contains(label, "operating income"):
			stmt.Core.OperatingIncome = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "beban keuangan") || strings.Contains(label, "finance costs"):
			stmt.Core.FinanceCosts = val * stmt.Metadata.RoundingMultiplier
		case strings.Contains(label, "laba (rugi) periode berjalan") || strings.Contains(label, "profit (loss) for the period") || strings.Contains(label, "laba bersih"):
			if stmt.Core.NetIncome == 0 {
				stmt.Core.NetIncome = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "jumlah rata-rata tertimbang saham") || strings.Contains(label, "weighted average number of shares"):
			if stmt.Core.SharesOutstanding <= 1 {
				stmt.Core.SharesOutstanding = val
			}
		}
	}
}

func parseExcelCashFlow(f *excelize.File, stmt *domain.Statement) {
	sheet := findMatchingSheet(f, []string{"1510000", "cashflow", "arus kas", "cash"})
	if sheet == "" {
		return
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return
	}

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		label := strings.ToLower(strings.TrimSpace(row[0]))
		val := findFirstNumericCell(row)

		switch {
		case strings.Contains(label, "kas neto yang diperoleh dari (digunakan untuk) aktivitas operasi") || strings.Contains(label, "net cash flows from (used in) operating activities"):
			if stmt.Core.OperatingCashFlow == 0 {
				stmt.Core.OperatingCashFlow = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "perolehan aset tetap") || strings.Contains(label, "payments for property, plant and equipment") || strings.Contains(label, "penambahan aset tetap"):
			if stmt.Core.CapEx == 0 {
				stmt.Core.CapEx = val * stmt.Metadata.RoundingMultiplier
			}
		}
	}
}

func findMatchingSheet(f *excelize.File, candidates []string) string {
	sheets := f.GetSheetList()
	for _, cand := range candidates {
		for _, s := range sheets {
			if strings.Contains(strings.ToLower(s), cand) {
				return s
			}
		}
	}
	return ""
}

func findFirstNumericCell(row []string) float64 {
	for i := 1; i < len(row); i++ {
		cell := strings.TrimSpace(row[i])
		if cell == "" || cell == "-" {
			continue
		}
		// Handle parentheses for negative numbers e.g. (1,234)
		isNeg := false
		if strings.HasPrefix(cell, "(") && strings.HasSuffix(cell, ")") {
			isNeg = true
			cell = strings.TrimPrefix(strings.TrimSuffix(cell, ")"), "(")
		}
		cell = strings.ReplaceAll(cell, ",", "")
		cell = strings.ReplaceAll(cell, " ", "")
		if val, err := strconv.ParseFloat(cell, 64); err == nil {
			if isNeg {
				val = -val
			}
			return val
		}
	}
	return 0
}
