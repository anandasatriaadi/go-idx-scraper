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
	// Initialize Ticker/Year/Period from filename if available
	base := filepath.Base(stmt.Metadata.SourceFile)
	parts := strings.Split(base, "-")
	if len(parts) >= 4 {
		if y, err := strconv.Atoi(parts[1]); err == nil && y > 2000 {
			stmt.Year = y
		}
		pUpper := strings.ToUpper(parts[2])
		switch pUpper {
		case "I", "TW1", "Q1":
			stmt.Period = "Q1"
		case "II", "TW2", "Q2":
			stmt.Period = "Q2"
		case "III", "TW3", "Q3":
			stmt.Period = "Q3"
		case "IV", "TW4", "Q4":
			stmt.Period = "Q4"
		case "TAHUNAN", "AUDIT", "FY":
			stmt.Period = "FY"
		default:
			stmt.Period = parts[2]
		}
		stmt.Ticker = strings.ToUpper(strings.TrimSuffix(parts[3], filepath.Ext(parts[3])))
	}

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
		case strings.Contains(key, "jumlah saham") || strings.Contains(key, "number of shares") || strings.Contains(key, "outstanding shares") || strings.Contains(key, "disetor penuh") || strings.Contains(key, "modal saham"):
			if shares, err := parseNumericValue(val); err == nil && shares > 0 {
				if stmt.Core.SharesOutstanding <= 1 {
					stmt.Core.SharesOutstanding = shares
				}
			}
		case strings.Contains(key, "periode laporan") || strings.Contains(key, "period of financial"):
			stmt.PeriodType = val
			valLower := strings.ToLower(val)
			if strings.Contains(val, "Kuartal I") || strings.Contains(val, "First") || strings.Contains(valLower, "maret") || strings.Contains(valLower, "march") || strings.Contains(val, "-03-") {
				stmt.Period = "Q1"
			} else if strings.Contains(val, "Kuartal II") || strings.Contains(val, "Second") || strings.Contains(valLower, "juni") || strings.Contains(valLower, "june") || strings.Contains(val, "-06-") {
				stmt.Period = "Q2"
			} else if strings.Contains(val, "Kuartal III") || strings.Contains(val, "Third") || strings.Contains(valLower, "september") || strings.Contains(val, "-09-") {
				stmt.Period = "Q3"
			} else if strings.Contains(val, "Tahunan") || strings.Contains(val, "Annual") || strings.Contains(val, "Audit") || strings.Contains(valLower, "desember") || strings.Contains(valLower, "december") || strings.Contains(val, "-12-") {
				stmt.Period = "FY"
			}
		case (strings.Contains(key, "tanggal akhir periode tahun berjalan") || strings.Contains(key, "current period end date") || strings.Contains(key, "tanggal akhir periode")) && !strings.Contains(key, "sebelumnya") && !strings.Contains(key, "prior") && !strings.Contains(key, "lalu"):
			if t, err := time.Parse("2006-01-02", val); err == nil {
				stmt.PeriodEndDate = t
				stmt.Year = t.Year()
				switch t.Month() {
				case time.March:
					if stmt.Period == "" || stmt.Period == "FY" {
						stmt.Period = "Q1"
					}
				case time.June:
					if stmt.Period == "" || stmt.Period == "FY" {
						stmt.Period = "Q2"
					}
				case time.September:
					if stmt.Period == "" || stmt.Period == "FY" {
						stmt.Period = "Q3"
					}
				case time.December:
					stmt.Period = "FY"
				}
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
	sheet := findMatchingSheet(f, []string{"1110000", "1210000", "4220000", "4210000", "neraca", "balance", "posisi keuangan"})
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
		case label == "kas" || strings.Contains(label, "kas dan setara kas") || strings.Contains(label, "cash and cash equivalents") || strings.Contains(label, "giro pada bank indonesia"):
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
	sheet := findMatchingSheet(f, []string{"1311000", "1321000", "4312000", "4322000", "4311000", "rugilaba", "income", "profit", "laba rugi"})
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
		case strings.Contains(label, "penjualan dan pendapatan") || strings.Contains(label, "sales and revenue") || strings.Contains(label, "pendapatan bunga") || strings.Contains(label, "interest income") || strings.Contains(label, "pendapatan operasional"):
			if stmt.Core.Revenue == 0 {
				stmt.Core.Revenue = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "beban pokok penjualan") || strings.Contains(label, "cost of sales and revenue") || strings.Contains(label, "beban bunga") || strings.Contains(label, "interest expense"):
			if stmt.Core.CostOfRevenue == 0 {
				stmt.Core.CostOfRevenue = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "jumlah laba bruto") || strings.Contains(label, "gross profit") || strings.Contains(label, "pendapatan bunga bersih") || strings.Contains(label, "net interest income"):
			if stmt.Core.GrossProfit == 0 {
				stmt.Core.GrossProfit = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "laba (rugi) usaha") || strings.Contains(label, "operating profit") || strings.Contains(label, "operating income") || strings.Contains(label, "laba operasional"):
			if stmt.Core.OperatingIncome == 0 {
				stmt.Core.OperatingIncome = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "beban keuangan") || strings.Contains(label, "finance costs"):
			if stmt.Core.FinanceCosts == 0 {
				stmt.Core.FinanceCosts = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "laba (rugi) yang dapat diatribusikan ke entitas induk") || strings.Contains(label, "profit (loss) attributable to parent entity"):
			if stmt.Core.NetIncomeParent == 0 {
				stmt.Core.NetIncomeParent = val * stmt.Metadata.RoundingMultiplier
			}
			if stmt.Core.NetIncome == 0 {
				stmt.Core.NetIncome = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "jumlah laba (rugi)") || strings.Contains(label, "total profit (loss)") || strings.Contains(label, "laba (rugi) periode berjalan") || strings.Contains(label, "laba (rugi) tahun berjalan") || strings.Contains(label, "profit (loss) for the period") || strings.Contains(label, "laba bersih"):
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
	sheet := findMatchingSheet(f, []string{"1510000", "4510000", "4520000", "cashflow", "arus kas", "cash"})
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
		case strings.Contains(label, "aktivitas operasi") || strings.Contains(label, "operating activities"):
			if strings.Contains(label, "kas neto") || strings.Contains(label, "net cash") || strings.Contains(label, "jumlah") {
				if stmt.Core.OperatingCashFlow == 0 {
					stmt.Core.OperatingCashFlow = val * stmt.Metadata.RoundingMultiplier
				}
			}
		case strings.Contains(label, "aktivitas investasi") || strings.Contains(label, "investing activities"):
			if strings.Contains(label, "kas neto") || strings.Contains(label, "net cash") || strings.Contains(label, "jumlah") {
				if stmt.Core.InvestingCashFlow == 0 {
					stmt.Core.InvestingCashFlow = val * stmt.Metadata.RoundingMultiplier
				}
			}
		case strings.Contains(label, "aktivitas pendanaan") || strings.Contains(label, "financing activities"):
			if strings.Contains(label, "kas neto") || strings.Contains(label, "net cash") || strings.Contains(label, "jumlah") {
				if stmt.Core.FinancingCashFlow == 0 {
					stmt.Core.FinancingCashFlow = val * stmt.Metadata.RoundingMultiplier
				}
			}
		case strings.Contains(label, "perolehan aset tetap") || strings.Contains(label, "payments for property, plant and equipment") || strings.Contains(label, "penambahan aset tetap") || strings.Contains(label, "pembelian aset tetap"):
			if stmt.Core.CapEx == 0 {
				stmt.Core.CapEx = val * stmt.Metadata.RoundingMultiplier
			}
		case strings.Contains(label, "pembayaran dividen") || strings.Contains(label, "dividends paid"):
			if stmt.Core.DividendsPaid == 0 {
				stmt.Core.DividendsPaid = val * stmt.Metadata.RoundingMultiplier
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
