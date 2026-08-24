package xbrl

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FactValue struct {
	Value    float64 `bson:"value" json:"value"`
	Unit     string  `bson:"unit,omitempty" json:"unit,omitempty"`
	Decimals int     `bson:"decimals,omitempty" json:"decimals,omitempty"`
}

type FactMap map[string]map[string]FactValue // [Tag][ContextRef] -> FactValue

type StatementMetadata struct {
	Sector             string  `bson:"sector" json:"sector"`
	Subsector          string  `bson:"subsector" json:"subsector"`
	Industry           string  `bson:"industry" json:"industry"`
	Subindustry        string  `bson:"subindustry" json:"subindustry"`
	AccountingStandard string  `bson:"accounting_standard" json:"accounting_standard"`
	Currency           string  `bson:"currency" json:"currency"`
	RoundingLevel      string  `bson:"rounding_level" json:"rounding_level"`
	RoundingMultiplier float64 `bson:"rounding_multiplier" json:"rounding_multiplier"`
	AuditStatus        string  `bson:"audit_status" json:"audit_status"`
	AuditorOpinion     string  `bson:"auditor_opinion,omitempty" json:"auditor_opinion,omitempty"`
	AuditorName        string  `bson:"auditor_name,omitempty" json:"auditor_name,omitempty"`
	SourceFile         string  `bson:"source_file,omitempty" json:"source_file,omitempty"`
	ConversionRate     float64 `bson:"conversion_rate,omitempty" json:"conversion_rate,omitempty"`
}

type CoreFinancials struct {
	SharesOutstanding  float64 `bson:"shares_outstanding" json:"shares_outstanding"`
	TotalAssets        float64 `bson:"total_assets" json:"total_assets"`
	CashAndEquivalents float64 `bson:"cash_and_equivalents" json:"cash_and_equivalents"`
	CurrentAssets      float64 `bson:"current_assets" json:"current_assets"`
	TotalLiabilities   float64 `bson:"total_liabilities" json:"total_liabilities"`
	CurrentLiabilities float64 `bson:"current_liabilities" json:"current_liabilities"`
	WorkingCapital     float64 `bson:"working_capital" json:"working_capital"`
	ShortTermDebt      float64 `bson:"short_term_debt" json:"short_term_debt"`
	LongTermDebt       float64 `bson:"long_term_debt" json:"long_term_debt"`
	TotalDebt          float64 `bson:"total_debt" json:"total_debt"`
	TotalEquity        float64 `bson:"total_equity" json:"total_equity"`
	RetainedEarnings   float64 `bson:"retained_earnings" json:"retained_earnings"`

	Revenue            float64 `bson:"revenue" json:"revenue"`
	CostOfRevenue      float64 `bson:"cost_of_revenue" json:"cost_of_revenue"`
	GrossProfit        float64 `bson:"gross_profit" json:"gross_profit"`
	OperatingIncome    float64 `bson:"operating_income" json:"operating_income"`
	EBITDA             float64 `bson:"ebitda" json:"ebitda"`
	FinanceCosts       float64 `bson:"finance_costs" json:"finance_costs"`
	NetIncome          float64 `bson:"net_income" json:"net_income"`
	NetIncomeParent    float64 `bson:"net_income_parent" json:"net_income_parent"`

	OperatingCashFlow  float64 `bson:"operating_cash_flow" json:"operating_cash_flow"`
	InvestingCashFlow  float64 `bson:"investing_cash_flow" json:"investing_cash_flow"`
	FinancingCashFlow  float64 `bson:"financing_cash_flow" json:"financing_cash_flow"`
	CapEx              float64 `bson:"capex" json:"capex"`
	FreeCashFlow       float64 `bson:"free_cash_flow" json:"free_cash_flow"`
	DividendsPaid      float64 `bson:"dividends_paid" json:"dividends_paid"`
}

type ComputedRatios struct {
	ROIC                  float64 `bson:"roic" json:"roic"`
	ROE                   float64 `bson:"roe" json:"roe"`
	ROA                   float64 `bson:"roa" json:"roa"`
	GrossMarginPct        float64 `bson:"gross_margin_pct" json:"gross_margin_pct"`
	OperatingMarginPct    float64 `bson:"operating_margin_pct" json:"operating_margin_pct"`
	NetMarginPct          float64 `bson:"net_margin_pct" json:"net_margin_pct"`
	DebtToEquity          float64 `bson:"debt_to_equity" json:"debt_to_equity"`
	NetDebt               float64 `bson:"net_debt" json:"net_debt"`
	InterestCoverageRatio float64 `bson:"interest_coverage_ratio" json:"interest_coverage_ratio"`
	CurrentRatio          float64 `bson:"current_ratio" json:"current_ratio"`
	FCFConversionPct      float64 `bson:"fcf_conversion_pct" json:"fcf_conversion_pct"`
	PiotroskiFScore       int     `bson:"piotroski_f_score" json:"piotroski_f_score"`
	AltmanZScore          float64 `bson:"altman_z_score" json:"altman_z_score"`
}

type ValuationMetrics struct {
	NormalizedEPS        float64 `bson:"normalized_eps" json:"normalized_eps"`
	NormalizedBVPS       float64 `bson:"normalized_bvps" json:"normalized_bvps"`
	RevenuePerShare      float64 `bson:"revenue_per_share" json:"revenue_per_share"`
	CashPerShare         float64 `bson:"cash_per_share" json:"cash_per_share"`
	FreeCashFlowPerShare float64 `bson:"free_cash_flow_per_share" json:"free_cash_flow_per_share"`
	MarketCap            float64 `bson:"market_cap" json:"market_cap"`
	EnterpriseValue      float64 `bson:"enterprise_value" json:"enterprise_value"`
	GrahamNumber         float64 `bson:"graham_number" json:"graham_number"`
	DCFFairValue         float64 `bson:"dcf_fair_value" json:"dcf_fair_value"`
	CurrentPrice         float64 `bson:"current_price" json:"current_price"`
	MarginOfSafetyPct    float64 `bson:"margin_of_safety_pct" json:"margin_of_safety_pct"`
	PERatio              float64 `bson:"pe_ratio" json:"pe_ratio"`
	PBRatio              float64 `bson:"pb_ratio" json:"pb_ratio"`
	PSRatio              float64 `bson:"ps_ratio" json:"ps_ratio"`
	PFCFRatio            float64 `bson:"p_fcf_ratio" json:"p_fcf_ratio"`
	EVToEBIT             float64 `bson:"ev_to_ebit" json:"ev_to_ebit"`
	EVToEBITDA           float64 `bson:"ev_to_ebitda" json:"ev_to_ebitda"`
	EarningsYieldPct     float64 `bson:"earnings_yield_pct" json:"earnings_yield_pct"`
	QuickRatio           float64 `bson:"quick_ratio" json:"quick_ratio"`
}

type Statement struct {
	ID             bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	Ticker         string            `bson:"ticker" json:"ticker"`
	CompanyName    string            `bson:"company_name" json:"company_name"`
	Year           int               `bson:"year" json:"year"`
	Period         string            `bson:"period" json:"period"`
	PeriodType     string            `bson:"period_type" json:"period_type"`
	PeriodEndDate  time.Time         `bson:"period_end_date" json:"period_end_date"`
	StatementType  string            `bson:"statement_type" json:"statement_type"`
	Metadata       StatementMetadata `bson:"metadata" json:"metadata"`
	Core           CoreFinancials    `bson:"core" json:"core"`
	Facts          FactMap           `bson:"facts" json:"facts"`
	ComputedRatios ComputedRatios    `bson:"computed_ratios" json:"computed_ratios"`
	Valuation      ValuationMetrics  `bson:"valuation" json:"valuation"`
	CreatedAt      time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time         `bson:"updated_at" json:"updated_at"`
}

type Repository interface {
	Upsert(ctx context.Context, s *Statement) error
	FindByTickerAndPeriod(ctx context.Context, ticker string, year int, period string) (*Statement, error)
	FindHistoricalByTicker(ctx context.Context, ticker string, limit int) ([]*Statement, error)
	FindAllWithFilter(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Statement, error)
}
