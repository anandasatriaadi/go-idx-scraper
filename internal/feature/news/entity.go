package news

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	StatusPending    = "pending"
	StatusSummarized = "summarized"
	StatusFailed     = "failed"
)

type News struct {
	ID                 bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt          time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `bson:"updated_at" json:"updated_at"`
	Date               time.Time     `bson:"date" json:"date"`
	Title              string        `bson:"title" json:"title"`
	Summary            string        `bson:"summary" json:"summary"`
	Content            string        `bson:"content" json:"content"`
	Link               string        `bson:"link" json:"link"`
	Priority           int           `bson:"priority" json:"priority"`
	ValueScore         int           `bson:"value_score" json:"value_score"`
	ImpactDirection    string        `bson:"impact_direction" json:"impact_direction"`
	InvestmentTakeaway string        `bson:"investment_takeaway" json:"investment_takeaway"`
	Tickers            []string      `bson:"tickers" json:"tickers"`
	Sector             string        `bson:"sector" json:"sector"`
	Subsector          string        `bson:"subsector" json:"subsector"`
	Industry           string        `bson:"industry,omitempty" json:"industry,omitempty"`
	IsIndustryWide     bool          `bson:"is_industry_wide" json:"is_industry_wide"`
	Status             string        `bson:"status,omitempty" json:"status,omitempty"`
}

type BriefingItem struct {
	Ticker             string `bson:"ticker,omitempty" json:"ticker,omitempty"`
	IssuerName         string `bson:"issuer_name,omitempty" json:"issuer_name,omitempty"`
	Headline           string `bson:"headline" json:"headline"`
	Rationale          string `bson:"rationale" json:"rationale"`
	ValueScore         int    `bson:"value_score" json:"value_score"`
	InvestmentTakeaway string `bson:"investment_takeaway" json:"investment_takeaway"`
}

type SectorHighlight struct {
	Sector    string `bson:"sector" json:"sector"`
	Summary   string `bson:"summary" json:"summary"`
	Sentiment string `bson:"sentiment" json:"sentiment"`
}

type Briefing struct {
	ID               bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	Date             time.Time         `bson:"date" json:"date"`
	Title            string            `bson:"title" json:"title"`
	MacroPulse       string            `bson:"macro_pulse" json:"macro_pulse"`
	BullishLookout   []BriefingItem    `bson:"bullish_lookout" json:"bullish_lookout"`
	BearishLookout   []BriefingItem    `bson:"bearish_lookout" json:"bearish_lookout"`
	SectorHighlights []SectorHighlight `bson:"sector_highlights" json:"sector_highlights"`
	ActionPlan       string            `bson:"action_plan" json:"action_plan"`
	RawMarkdown      string            `bson:"raw_markdown" json:"raw_markdown"`
	CreatedAt        time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time         `bson:"updated_at" json:"updated_at"`
}

// Port: Repository defines news data persistence
type Repository interface {
	Create(ctx context.Context, news *News) error
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*News, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*News, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update any) error
	ExistsByLink(ctx context.Context, link string) (bool, error)
	FindPendingSummary(ctx context.Context, limit int) ([]*News, error)
}

// Port: BriefingRepository defines briefing data persistence
type BriefingRepository interface {
	Create(ctx context.Context, b *Briefing) error
	FindByDate(ctx context.Context, date time.Time) (*Briefing, error)
	FindLatest(ctx context.Context) (*Briefing, error)
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Briefing, error)
}

// Port: Scraper defines news scraping interface
type Scraper interface {
	Scrape(ctx context.Context, startDate, endDate time.Time, onNewsFound func(*News) error) error
}
