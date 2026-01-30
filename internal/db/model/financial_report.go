package model

//go:generate go run ../../../generator/mongogen/main.go -type=FinancialReport -collection=financial_reports
type FinancialReport struct {
	Year         int    `bson:"year"`
	Quarter      int    `bson:"quarter"`
	IssuerCode   string `bson:"issuer_code"`
	ReportURL    string `bson:"report_url"`
	DownloadedAt int64  `bson:"downloaded_at"`
}
