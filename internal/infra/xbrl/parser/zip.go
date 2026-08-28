package parser

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strings"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ParseInstanceZip opens an instance.zip or inlineXBRL.zip file and parses the contents
func ParseInstanceZip(zipPath string) (*domain.Statement, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip file: %w", err)
	}
	defer r.Close()

	// 1. Try finding standalone .xbrl or .xml instance file
	for _, f := range r.File {
		nameLower := strings.ToLower(f.Name)
		if (strings.HasSuffix(nameLower, ".xbrl") || strings.HasSuffix(nameLower, ".xml")) && !strings.Contains(nameLower, "taxonomy") {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("reading file in zip: %w", err)
			}
			defer rc.Close()
			stmt, err := ParseInstanceXML(rc)
			if err != nil {
				return nil, err
			}
			stmt.Metadata.SourceFile = filepath.Base(zipPath)
			return stmt, nil
		}
	}

	// 2. Try parsing inline XBRL HTML files inside zip (e.g. 1000000.html, 1210000.html)
	stmt := &domain.Statement{
		Metadata: domain.StatementMetadata{
			RoundingMultiplier: 1.0,
			SourceFile:         filepath.Base(zipPath),
		},
		Facts: make(domain.FactMap),
	}
	hasParsedHTML := false
	for _, f := range r.File {
		nameLower := strings.ToLower(f.Name)
		if strings.HasSuffix(nameLower, ".html") || strings.HasSuffix(nameLower, ".htm") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			s, err := ParseInstanceXML(rc)
			rc.Close()
			if err == nil {
				hasParsedHTML = true
				MergeStatements(stmt, s)
			}
		}
	}

	if hasParsedHTML && (stmt.Ticker != "" || stmt.CompanyName != "") {
		FinalizeCoreFinancials(stmt)
		return stmt, nil
	}

	return nil, fmt.Errorf("no valid .xbrl, .xml, or inline XBRL .html found in zip archive: %s", zipPath)
}

// MergeStatements merges source statement fields into target statement
func MergeStatements(target, source *domain.Statement) {
	if target.Ticker == "" {
		target.Ticker = source.Ticker
	}
	if target.CompanyName == "" {
		target.CompanyName = source.CompanyName
	}
	if target.Year == 0 {
		target.Year = source.Year
	}
	if target.Period == "" {
		target.Period = source.Period
	}
	if target.PeriodType == "" {
		target.PeriodType = source.PeriodType
	}
	if target.Metadata.Sector == "" {
		target.Metadata.Sector = source.Metadata.Sector
	}
	if target.Metadata.Industry == "" {
		target.Metadata.Industry = source.Metadata.Industry
	}
	if target.Metadata.Currency == "" {
		target.Metadata.Currency = source.Metadata.Currency
	}
	if target.Metadata.ConversionRate == 0 {
		target.Metadata.ConversionRate = source.Metadata.ConversionRate
	}

	// Merge core
	if target.Core.TotalAssets == 0 {
		target.Core.TotalAssets = source.Core.TotalAssets
	}
	if target.Core.CashAndEquivalents == 0 {
		target.Core.CashAndEquivalents = source.Core.CashAndEquivalents
	}
	if target.Core.TotalLiabilities == 0 {
		target.Core.TotalLiabilities = source.Core.TotalLiabilities
	}
	if target.Core.TotalEquity == 0 {
		target.Core.TotalEquity = source.Core.TotalEquity
	}
	if target.Core.Revenue == 0 {
		target.Core.Revenue = source.Core.Revenue
	}
	if target.Core.GrossProfit == 0 {
		target.Core.GrossProfit = source.Core.GrossProfit
	}
	if target.Core.OperatingIncome == 0 {
		target.Core.OperatingIncome = source.Core.OperatingIncome
	}
	if target.Core.NetIncome == 0 {
		target.Core.NetIncome = source.Core.NetIncome
	}

	// Merge facts
	for k, v := range source.Facts {
		if target.Facts[k] == nil {
			target.Facts[k] = make(map[string]domain.FactValue)
		}
		for ctxRef, fv := range v {
			target.Facts[k][ctxRef] = fv
		}
	}
}
