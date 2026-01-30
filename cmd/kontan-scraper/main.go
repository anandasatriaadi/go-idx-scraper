package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/chromedp/chromedp"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

// Article struct (updated to include content)
type Article struct {
	Title    string
	Date     string
	Category string
	Link     string
	Image    string
	Content  string // Markdown from ".tmpt-desk-kon"
}

// ContentProcessor defines an interface for processing article content
type ContentProcessor interface {
	Process(htmlContent string) (string, error)
}

// HTMLToMarkdownProcessor implements ContentProcessor
type HTMLToMarkdownProcessor struct {
	sanitizer *bluemonday.Policy
	logger    *zap.Logger
}

// NewHTMLToMarkdownProcessor creates a new instance of HTMLToMarkdownProcessor
func NewHTMLToMarkdownProcessor(logger *zap.Logger) *HTMLToMarkdownProcessor {
	return &HTMLToMarkdownProcessor{
		sanitizer: bluemonday.UGCPolicy(),
		logger:    logger,
	}
}

// Process filters, sanitizes, and converts HTML to Markdown
func (p *HTMLToMarkdownProcessor) Process(htmlContent string) (string, error) {
	cleanHtml := filterArticleContent(htmlContent, p.logger)
	if strings.TrimSpace(cleanHtml) == "" {
		return "", fmt.Errorf("no clean content after filtering")
	}

	sanitizedHtml := p.sanitizer.Sanitize(cleanHtml)

	markdown, err := htmltomarkdown.ConvertString(sanitizedHtml)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	return markdown, nil
}

func main() {
	logger, err := initializeLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		return
	}

	// Connect to MongoDB
	ctx := context.Background()
	db, err := db.New(logger)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	repo := model.NewNewsRepository(db.GetDatabase("idx"))

	// Parse start and end dates from command line arguments
	now := time.Now()
	var startDate, endDate time.Time
	if len(os.Args) == 2 {
		startDate = now
		endDate = now
	} else if len(os.Args) == 3 {
		startDate, err = time.Parse("2006-01-02", os.Args[2])
		if err != nil {
			logger.Fatal("Failed to parse startDate", zap.Error(err))
		}
		endDate = startDate
	} else if len(os.Args) == 4 {
		startDate, err = time.Parse("2006-01-02", os.Args[2])
		if err != nil {
			logger.Fatal("Failed to parse startDate", zap.Error(err))
		}
		endDate, err = time.Parse("2006-01-02", os.Args[3])
		if err != nil {
			logger.Fatal("Failed to parse endDate", zap.Error(err))
		}
	} else {
		logger.Fatal("Invalid number of arguments. Usage: <program> <config_file> [startDate] [endDate]")
	}

	// Setup Chromedp context
	chromeCtx, cancel := br.SetupChromeDp(cfg)
	defer cancel()

	processor := NewHTMLToMarkdownProcessor(logger)

	// Loop over each date in the range
	articleIndex := 0
	var ids []bson.ObjectID
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		logger.Info("Fetching articles for date", zap.Time("date", d))

		// Step 1: Extract and parse the article list HTML for the date, handling pagination
		var allArticles []Article
		perPage := 0
		for {
			htmlContent, err := fetchArticleListHTML(chromeCtx, d, perPage)
			if err != nil {
				logger.Error("Failed to fetch article list HTML", zap.Time("date", d), zap.Int("per_page", perPage), zap.Error(err))
				break
			}

			logger.Debug("Raw HTML Content from list", zap.String("html", htmlContent))

			// Step 2: Parse articles from the HTML
			articles := parseArticlesFromHTML(htmlContent, logger)
			logger.Info("Parsed articles count for page", zap.Time("date", d), zap.Int("per_page", perPage), zap.Int("count", len(articles)))

			// Append to all articles
			allArticles = append(allArticles, articles...)

			// If less than 20 articles, no more pages
			if len(articles) < 20 {
				break
			}

			perPage += 20
		}

		logger.Info("Total articles for date", zap.Time("date", d), zap.Int("count", len(allArticles)))

		// Step 3: For each article, fetch full content from ".tmpt-desk-kon"
		for _, art := range allArticles {
			articleIndex++
			logger.Info("Fetching article content", zap.Int("index", articleIndex), zap.String("link", art.Link))

			articleHtml, err := fetchArticleContent(chromeCtx, art.Link)
			if err != nil {
				logger.Error("Failed to fetch article content", zap.String("link", art.Link), zap.Error(err))
				continue
			}

			markdown, err := processor.Process(articleHtml)
			if err != nil {
				logger.Warn("Failed to process article content", zap.String("link", art.Link), zap.Error(err))
				continue
			}

			// Update article content
			art.Content = markdown

			// Parse date
			dateParsed, err := parseDate(art.Date)
			if err != nil {
				logger.Warn("Failed to parse date", zap.String("dateStr", art.Date), zap.Error(err))
				continue
			}

			// Create News model
			news := &model.News{
				Id:       bson.NewObjectID(),
				Title:    art.Title,
				Date:     dateParsed,
				Summary:  "",
				Content:  art.Content,
				Priority: 10,
				Link:     art.Link,
			}

			// Save to DB
			_, err = repo.Create(ctx, news)
			if err != nil {
				logger.Error("Failed to save news to DB", zap.String("link", art.Link), zap.Error(err))
				continue
			}

			ids = append(ids, news.Id)
		}
	}

	if len(ids) > 0 {
		err = helper.SummarizeNews(ctx, logger, ids, repo)
		if err != nil {
			logger.Error("Failed to summarize news", zap.Error(err))
		}
	}
}

// parseDate parses the date string into time.Time
func parseDate(dateStr string) (time.Time, error) {
	// Assuming format like "02 Januari 2023" - adjust as needed
	// Map Indonesian months to English for parsing
	monthMap := map[string]string{
		"Januari":   "January",
		"Februari":  "February",
		"Maret":     "March",
		"April":     "April",
		"Mei":       "May",
		"Juni":      "June",
		"Juli":      "July",
		"Agustus":   "August",
		"September": "September",
		"Oktober":   "October",
		"November":  "November",
		"Desember":  "December",
	}
	for id, en := range monthMap {
		dateStr = strings.Replace(dateStr, id, en, -1)
	}
	return time.Parse("02 January 2006", dateStr)
}

// initializeLogger sets up the zap logger
func initializeLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

// loadConfig loads the configuration from the provided file
func loadConfig() (*config.Config, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("no config file provided, usage: %s <config_file>", os.Args[0])
	}
	configPath := os.Args[1]
	return config.Load(configPath)
}

// fetchArticleListHTML fetches the HTML content of the article list for a specific date and per_page
func fetchArticleListHTML(ctx context.Context, date time.Time, perPage int) (string, error) {
	var htmlContent string
	var perPageStr string
	if perPage == 0 {
		perPageStr = ""
	} else {
		perPageStr = fmt.Sprintf("%d", perPage)
	}
	day := fmt.Sprintf("%02d", date.Day())
	month := fmt.Sprintf("%02d", int(date.Month()))
	year := date.Year()
	url := fmt.Sprintf("https://www.kontan.co.id/search/indeks?kanal=investasi&tanggal=%s&bulan=%s&tahun=%d&pos=indeks&per_page=%s", day, month, year, perPageStr)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(200*time.Millisecond), // Adjust for dynamic loading
		chromedp.InnerHTML("div.list-berita > ul", &htmlContent),
	)
	if err != nil {
		return "", fmt.Errorf("chromedp run error for list: %w", err)
	}
	return htmlContent, nil
}

// fetchArticleContent fetches the HTML content of a single article
func fetchArticleContent(ctx context.Context, link string) (string, error) {
	var articleHtml string
	var paginatedHtml string
	err := chromedp.Run(ctx,
		chromedp.Navigate(link),
		chromedp.Sleep(200*time.Millisecond), // Adjust for dynamic loading
		chromedp.InnerHTML(".tmpt-desk-kon", &articleHtml),
		chromedp.InnerHTML("div.pagination", &paginatedHtml),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch article content: %w", err)
	}

	if strings.TrimSpace(articleHtml) == "" {
		return "", fmt.Errorf("no content found for article")
	}

	if strings.TrimSpace(paginatedHtml) != "" {
		return "", fmt.Errorf("paginated page found, skipping")
	}

	return articleHtml, nil
}

// parseArticlesFromHTML (unchanged from previous version)
func parseArticlesFromHTML(htmlCode string, logger *zap.Logger) []Article {
	if strings.TrimSpace(htmlCode) == "" {
		logger.Warn("No HTML content to parse")
		return nil
	}

	var articles []Article
	reader := strings.NewReader(htmlCode)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		logger.Error("Failed to parse HTML", zap.Error(err))
		return nil
	}

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h1 > a").Text()
		link, _ := s.Find("h1 > a").Attr("href")
		if !strings.HasPrefix(link, "http") {
			link = "https:" + link
		}
		category := strings.TrimSpace(s.Find(".ket > div.fs14 span:first-child").Text())
		date := strings.TrimSpace(strings.Split(s.Find(".ket > div.fs14").Children().Eq(1).Text(), "|")[1])
		imgSrc, _ := s.Find(".pic > img").Attr("src")
		dataSrc, _ := s.Find(".pic > img").Attr("data-src")
		image := imgSrc
		if image == "" {
			image = dataSrc
		}
		if !strings.HasPrefix(image, "http") && image != "" {
			image = "https://foto.kontan.co.id" + image
		}
		if title != "" && link != "" {
			articles = append(articles, Article{
				Title:    title,
				Date:     date,
				Category: category,
				Link:     link,
				Image:    image,
			})
		}
	})

	return articles
}

func filterArticleContent(htmlContent string, logger *zap.Logger) string {
	reader := strings.NewReader(htmlContent)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		logger.Error("Failed to parse article HTML for cleaning", zap.Error(err))
		return htmlContent // Fallback to original
	}

	// Remove empty or non-content <p> (spammers)
	doc.Find("h2:has(span), p:has(span), p:has(a)").FilterFunction(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		text = strings.ToLower(text)
		return text == "" || strings.Contains(text, "baca juga") || strings.Contains(text, "selanjutnya") || strings.Contains(text, "menarik dibaca") || strings.Contains(text, "google news")
	}).Remove()

	var body string
	var fallbackText []string
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" { // Skip truly empty <p>
			fallbackText = append(fallbackText, text)
		}
	})
	if len(fallbackText) > 0 {
		body = "<p>" + strings.Join(fallbackText, "</p><p>") + "</p>"
	} else {
		body = htmlContent // Ultimate fallback
	}

	return strings.TrimSpace(body)
}
