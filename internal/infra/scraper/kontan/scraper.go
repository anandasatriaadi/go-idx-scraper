package kontan

import (
	"context"
	"fmt"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"github.com/microcosm-cc/bluemonday"
	"github.com/tebeka/selenium"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type Browser interface {
	FetchList(ctx context.Context, url string) (string, error)
	FetchContent(ctx context.Context, url string) (content string, pagination string, err error)
}

type DefaultBrowser struct {
	driver selenium.WebDriver
}

func NewDefaultBrowser(driver selenium.WebDriver) *DefaultBrowser {
	return &DefaultBrowser{driver: driver}
}

func (b *DefaultBrowser) FetchList(ctx context.Context, url string) (string, error) {
	if err := b.driver.Get(url); err != nil {
		return "", err
	}
	// Simple sleep to let JS render
	time.Sleep(500 * time.Millisecond)

	elem, err := b.driver.FindElement(selenium.ByCSSSelector, "div.list-berita > ul")
	if err != nil {
		return "", err
	}
	return elem.GetAttribute("innerHTML")
}

func (b *DefaultBrowser) FetchContent(ctx context.Context, url string) (string, string, error) {
	if err := b.driver.Get(url); err != nil {
		return "", "", err
	}
	time.Sleep(500 * time.Millisecond)

	articleElem, err := b.driver.FindElement(selenium.ByCSSSelector, ".tmpt-desk-kon")
	if err != nil {
		return "", "", err
	}
	articleHtml, _ := articleElem.GetAttribute("innerHTML")

	var paginatedHtml string
	pagElem, err := b.driver.FindElement(selenium.ByCSSSelector, "div.pagination")
	if err == nil {
		paginatedHtml, _ = pagElem.GetAttribute("innerHTML")
	}

	return articleHtml, paginatedHtml, nil
}

type Scraper struct {
	logger    *zap.Logger
	processor *HTMLToMarkdownProcessor
	browser   Browser
}

func NewScraper(logger *zap.Logger, browser Browser) *Scraper {
	return &Scraper{
		logger:    logger,
		processor: NewHTMLToMarkdownProcessor(logger),
		browser:   browser,
	}
}

func (s *Scraper) WithBrowser(b Browser) *Scraper {
	s.browser = b
	return s
}

// Article struct internal to scraper
type Article struct {
	Title    string
	Date     string
	Category string
	Link     string
	Image    string
	Content  string
}

func (s *Scraper) Scrape(ctx context.Context, startDate, endDate time.Time, onNewsFound func(*news.News) error) error {
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		s.logger.Info("Fetching articles for date", zap.Time("date", d))

		perPage := 0
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			htmlContent, err := s.fetchArticleListHTML(ctx, d, perPage)
			if err != nil {
				s.logger.Warn("Failed to fetch article list HTML", zap.Time("date", d), zap.Int("per_page", perPage), zap.Error(err))
				break
			}

			articles := s.parseArticlesFromHTML(htmlContent)
			s.logger.Info("Parsed articles count for page", zap.Time("date", d), zap.Int("per_page", perPage), zap.Int("count", len(articles)))

			if len(articles) == 0 {
				break
			}

			for _, art := range articles {
				s.logger.Info("Fetching article content", zap.String("link", art.Link))

				articleHtml, err := s.fetchArticleContent(ctx, art.Link)
				if err != nil {
					s.logger.Warn("Failed to fetch article content", zap.String("link", art.Link), zap.Error(err))
					continue
				}

				markdown, err := s.processor.Process(articleHtml)
				if err != nil {
					s.logger.Warn("Failed to process article content", zap.String("link", art.Link), zap.Error(err))
					continue
				}

				dateParsed, err := s.parseDate(art.Date)
				if err != nil {
					s.logger.Warn("Failed to parse date", zap.String("dateStr", art.Date), zap.Error(err))
					continue
				}

				n := &news.News{
					ID:       bson.NewObjectID(),
					Title:    art.Title,
					Date:     dateParsed,
					Summary:  "",
					Content:  markdown,
					Priority: 10,
					Link:     art.Link,
				}

				if err := onNewsFound(n); err != nil {
					s.logger.Error("Failed to handle found news", zap.Error(err))
				}
			}

			if len(articles) < 20 {
				break
			}
			perPage += 20
		}
	}
	return nil
}

func (s *Scraper) fetchArticleListHTML(ctx context.Context, date time.Time, perPage int) (string, error) {
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
	htmlContent, err := s.browser.FetchList(ctx, url)
	if err != nil {
		return "", fmt.Errorf("browser fetch error for list: %w", err)
	}
	return htmlContent, nil
}

func (s *Scraper) fetchArticleContent(ctx context.Context, link string) (string, error) {
	articleHtml, paginatedHtml, err := s.browser.FetchContent(ctx, link)
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

func (s *Scraper) parseArticlesFromHTML(htmlCode string) []Article {
	if strings.TrimSpace(htmlCode) == "" {
		return nil
	}

	var articles []Article
	reader := strings.NewReader(htmlCode)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		s.logger.Warn("Failed to parse HTML", zap.Error(err))
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

func (s *Scraper) parseDate(dateStr string) (time.Time, error) {
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
		dateStr = strings.ReplaceAll(dateStr, id, en)
	}
	return time.Parse("02 January 2006", dateStr)
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
	cleanHtml := p.filterArticleContent(htmlContent)
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

func (p *HTMLToMarkdownProcessor) filterArticleContent(htmlContent string) string {
	reader := strings.NewReader(htmlContent)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		p.logger.Warn("Failed to parse article HTML for cleaning", zap.Error(err))
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
