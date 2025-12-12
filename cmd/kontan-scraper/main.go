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
	"github.com/chromedp/chromedp"
	"github.com/microcosm-cc/bluemonday"
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

	// Setup Chromedp context
	ctx, cancel := br.SetupChromeDp(cfg)
	defer cancel()

	processor := NewHTMLToMarkdownProcessor(logger)

	// Step 1: Extract and parse the article list HTML
	htmlContent, err := fetchArticleListHTML(ctx)
	if err != nil {
		logger.Error("Failed to fetch article list HTML", zap.Error(err))
		return
	}

	logger.Debug("Raw HTML Content from list", zap.String("html", htmlContent))

	// Step 2: Parse articles from the HTML
	articles := parseArticlesFromHTML(htmlContent, logger)
	logger.Info("Parsed articles count", zap.Int("count", len(articles)))

	// Step 3: For each article, fetch full content from ".tmpt-desk-kon"
	for i, art := range articles {
		logger.Info("Fetching article content", zap.Int("index", i+1), zap.String("link", art.Link))

		articleHtml, err := fetchArticleContent(ctx, art.Link)
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
		articles[i].Content = markdown

		// Output article summary (or save to file/DB)
		outputArticle(i+1, articles[i])
	}
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

// fetchArticleListHTML fetches the HTML content of the article list
func fetchArticleListHTML(ctx context.Context) (string, error) {
	var htmlContent string
	url := "https://www.kontan.co.id/search/indeks?kanal=keuangan&tanggal=08&bulan=12&tahun=2025&pos=indeks&per_page="
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("div.list-berita > ul"),
		chromedp.Sleep(1*time.Second), // Adjust for dynamic loading
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

// outputArticle prints the article details
func outputArticle(index int, article Article) {
	fmt.Printf("\n--- Article %d ---\nTitle: %s\nDate: %s\nCategory: %s\nLink: %s\nImage: %s\nContent:\n%s\n",
		index, article.Title, article.Date, article.Category, article.Link, article.Image, article.Content)
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
		ketText := s.Find(".ket").Text()
		parts := strings.Split(strings.TrimSpace(ketText), "|")
		category := strings.TrimSpace(parts[0])
		date := ""
		if len(parts) > 1 {
			date = strings.TrimSpace(parts[1])
		}
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
