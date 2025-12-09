package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"github.com/anandasatriaadi/go-idx-scraper/internal/types"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

// --- Main Logic ---

func main() {
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	zap.ReplaceGlobals(logger)
	if len(os.Args) < 2 {
		log.Fatalf("no config file provided. Usage: %s <config_file>", os.Args[0])
	}
	configPath := os.Args[1]
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	ctx, cancel := browser.SetupChromeDp(cfg)
	defer cancel()

	var data string
	// get page data as string
	if err := chromedp.Run(ctx, getPageData(`https://www.idx.co.id/primary/ListedCompany/GetAnnouncement?kodeEmiten=&emitenType=s&indexFrom=0&pageSize=100&dateFrom=19010101&dateTo=20251014&lang=id&keyword=`, &data)); err != nil {
		log.Panic(err)
	}
	if err := os.WriteFile("data.json", []byte(data), 0o644); err != nil {
		log.Panic(err)
	}

	database, err := db.New()

	var resp types.NewsResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		log.Panic(err)
	}

	// for each response Replies, set announcement to slice
	var announcements []*model.Announcement
	for _, ann := range resp.Replies {
		n := &ann.Announcement.Announcement
		temp := strings.Trim(*n.IssuerCode, " ")
		n.IssuerCode = &temp // Parse and assign AnnouncementDate if present
		if *ann.Announcement.AnnouncementDate != "" {
			if parsed, err := time.Parse("2006-01-02T15:04:05", *ann.Announcement.AnnouncementDate); err == nil {
				n.AnnouncementDate = &parsed
			} // Optionally handle parsing errors (e.g., log or return an error)
		}

		// Parse and assign CreatedDate if present
		if *ann.Announcement.CreatedDate != "" {
			if parsed, err := time.Parse("2006-01-02T15:04:05", *ann.Announcement.CreatedDate); err == nil {
				n.CreatedDate = &parsed
			} // Optionally handle parsing errors
		}
		n.Attachments = append(n.Attachments, ann.Attachment...)
		announcements = append(announcements, n)
	}

	aRepo := model.NewAnnouncementRepository(database.GetDatabase("idx"))
	aRepo.CreateMany(ctx, announcements)
	// log.Printf("Total Announcements: %v\n", resp)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	zap.L().Info("Memory usage", zap.Float64("MB", float64(m.Alloc)/(1024*1024)))

	zap.L().Info("Starting report generation process...")
}

// getPageData navigates to the URL and retrieves the page data as a string.
//
// Assumes the page content is JSON text in the body.
func getPageData(urlstr string, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Evaluate(`document.body.innerText`, res),
	}
}
