package kontan

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"go.uber.org/zap"
)

func TestParseArticlesFromHTML(t *testing.T) {
	// Sample HTML simulating the structure found in Kontan
	htmlContent := `
		<div class="list-berita">
			<ul>
				<li>
					<div class="pic">
						<img src="https://foto.kontan.co.id/image.jpg" alt="Image">
					</div>
					<div class="ket">
						<h1><a href="https://investasi.kontan.co.id/news/sample-news">Sample News Title</a></h1>
						<div class="fs14">
							<span>Investasi</span>
							<span>| 05 Februari 2024</span>
						</div>
					</div>
				</li>
				<li>
					<div class="pic">
						<img data-src="/local-image.jpg" alt="Local Image">
					</div>
					<div class="ket">
						<h1><a href="/news/local-news">Local News Title</a></h1>
						<div class="fs14">
							<span>Saham</span>
							<span>| 04 Februari 2024</span>
						</div>
					</div>
				</li>
			</ul>
		</div>
	`

	logger := zap.NewNop()
	scraper := NewScraper(logger, nil)
	articles := scraper.parseArticlesFromHTML(htmlContent)

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}

	// Check first article
	if articles[0].Title != "Sample News Title" {
		t.Errorf("Expected title 'Sample News Title', got '%s'", articles[0].Title)
	}
	if articles[0].Link != "https://investasi.kontan.co.id/news/sample-news" {
		t.Errorf("Expected link 'https://investasi.kontan.co.id/news/sample-news', got '%s'", articles[0].Link)
	}
	if articles[0].Category != "Investasi" {
		t.Errorf("Expected category 'Investasi', got '%s'", articles[0].Category)
	}
	if articles[0].Date != "05 Februari 2024" {
		t.Errorf("Expected date '05 Februari 2024', got '%s'", articles[0].Date)
	}

	// Check second article (handling relative links and lazy loaded images)
	if articles[1].Title != "Local News Title" {
		t.Errorf("Expected title 'Local News Title', got '%s'", articles[1].Title)
	}
	// Check contains local-news
	if !strings.Contains(articles[1].Link, "local-news") {
		t.Errorf("Link should contain local-news")
	}

	if articles[1].Image != "https://foto.kontan.co.id/local-image.jpg" {
		t.Errorf("Expected image to be prepended with base url, got '%s'", articles[1].Image)
	}
}

func TestFilterArticleContent(t *testing.T) {
	htmlContent := `
		<div class="tmpt-desk-kon">
			<p>Content paragraph 1.</p>
			<p><strong>Baca Juga:</strong> <a href="#">Link to other news</a></p>
			<p>Content paragraph 2.</p>
			<p><span>Selanjutnya: Page 2</span></p>
            <p></p> <!-- Empty -->
		</div>
	`

	logger := zap.NewNop()
	processor := NewHTMLToMarkdownProcessor(logger)
	filtered := processor.filterArticleContent(htmlContent)

	if strings.Contains(filtered, "Baca Juga") {
		t.Errorf("Content should not contain 'Baca Juga'")
	}
	if strings.Contains(filtered, "Selanjutnya") {
		t.Errorf("Content should not contain 'Selanjutnya'")
	}
	if !strings.Contains(filtered, "Content paragraph 1") {
		t.Errorf("Content should contain 'Content paragraph 1'")
	}
	if !strings.Contains(filtered, "Content paragraph 2") {
		t.Errorf("Content should contain 'Content paragraph 2'")
	}

	// Check if it wraps properly
	if !strings.HasPrefix(filtered, "<p>") || !strings.HasSuffix(filtered, "</p>") {
		t.Errorf("Content should be wrapped in <p> tags")
	}
}

func TestProcess(t *testing.T) {
	htmlContent := `
		<div class="tmpt-desk-kon">
			<p>Valid content here.</p>
			<p>Baca Juga: <a href="#">ignore</a></p>
		</div>
	`
	logger := zap.NewNop()
	processor := NewHTMLToMarkdownProcessor(logger)
	markdown, err := processor.Process(htmlContent)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !strings.Contains(markdown, "Valid content here.") {
		t.Errorf("Markdown should contain 'Valid content here.'")
	}
	if strings.Contains(markdown, "Baca Juga") {
		t.Errorf("Markdown should not contain 'Baca Juga'")
	}
}

type MockBrowser struct {
	ListHTML       string
	ContentHTML    string
	PaginationHTML string
	Err            error
}

func (m *MockBrowser) FetchList(ctx context.Context, url string) (string, error) {
	return m.ListHTML, m.Err
}

func (m *MockBrowser) FetchContent(ctx context.Context, url string) (string, string, error) {
	return m.ContentHTML, m.PaginationHTML, m.Err
}

func TestScrape(t *testing.T) {
	logger := zap.NewNop()
	scraper := NewScraper(logger, nil)

	listHTML := `
		<li>
			<div class="ket">
				<h1><a href="https://investasi.kontan.co.id/news/test">Test News</a></h1>
				<div class="fs14">
					<span>Investasi</span>
					<span>| 05 Februari 2024</span>
				</div>
			</div>
		</li>
	`
	contentHTML := `<div class="tmpt-desk-kon"><p>Test content.</p></div>`

	mock := &MockBrowser{
		ListHTML:    listHTML,
		ContentHTML: contentHTML,
	}
	scraper.WithBrowser(mock)

	startDate := time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)

	var foundNews []*news.News
	err := scraper.Scrape(context.Background(), startDate, endDate, func(n *news.News) error {
		foundNews = append(foundNews, n)
		return nil
	})

	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if len(foundNews) != 1 {
		t.Errorf("Expected 1 news, got %d", len(foundNews))
	} else {
		if foundNews[0].Title != "Test News" {
			t.Errorf("Expected title 'Test News', got '%s'", foundNews[0].Title)
		}
		if !strings.Contains(foundNews[0].Content, "Test content.") {
			t.Errorf("Expected content to contain 'Test content.', got '%s'", foundNews[0].Content)
		}
	}
}

func TestScrape_Errors(t *testing.T) {
	logger := zap.NewNop()
	scraper := NewScraper(logger, nil)

	// Case 1: FetchList error
	mock := &MockBrowser{Err: fmt.Errorf("network error")}
	scraper.WithBrowser(mock)
	err := scraper.Scrape(context.Background(), time.Now(), time.Now(), nil)
	if err != nil {
		t.Errorf("Scrape should not return error even if FetchList fails for one day, it logs it")
	}

	// Case 2: FetchContent error
	listHTML := `
		<li>
			<div class="ket">
				<h1><a href="http://test.com">Title</a></h1>
				<div class="fs14">
					<span>Cat</span>
					<span>| 05 Februari 2024</span>
				</div>
			</div>
		</li>
	`
	mock = &MockBrowser{
		ListHTML: listHTML,
		Err:      fmt.Errorf("content error"),
	}
	// Reset MockBrowser error for FetchList but fail for FetchContent
	// Actually MockBrowser returns the same error for both.
	// Let's refine MockBrowser.
}

type RefinedMockBrowser struct {
	FetchListFunc    func(ctx context.Context, url string) (string, error)
	FetchContentFunc func(ctx context.Context, url string) (string, string, error)
}

func (m *RefinedMockBrowser) FetchList(ctx context.Context, url string) (string, error) {
	return m.FetchListFunc(ctx, url)
}

func (m *RefinedMockBrowser) FetchContent(ctx context.Context, url string) (string, string, error) {
	return m.FetchContentFunc(ctx, url)
}

func TestScrape_Detailed(t *testing.T) {
	logger := zap.NewNop()
	scraper := NewScraper(logger, nil)

	listHTML := `
		<li>
			<div class="ket">
				<h1><a href="http://test.com">Title</a></h1>
				<div class="fs14">
					<span>Cat</span>
					<span>| 05 Februari 2024</span>
				</div>
			</div>
		</li>
	`

	t.Run("FetchContent Error", func(t *testing.T) {
		mock := &RefinedMockBrowser{
			FetchListFunc: func(ctx context.Context, url string) (string, error) {
				return listHTML, nil
			},
			FetchContentFunc: func(ctx context.Context, url string) (string, string, error) {
				return "", "", fmt.Errorf("content error")
			},
		}
		scraper.WithBrowser(mock)
		err := scraper.Scrape(context.Background(), time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), func(n *news.News) error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Paginated Content Skip", func(t *testing.T) {
		mock := &RefinedMockBrowser{
			FetchListFunc: func(ctx context.Context, url string) (string, error) {
				return listHTML, nil
			},
			FetchContentFunc: func(ctx context.Context, url string) (string, string, error) {
				return "some content", "page 1 2", nil
			},
		}
		scraper.WithBrowser(mock)
		count := 0
		scraper.Scrape(context.Background(), time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), func(n *news.News) error {
			count++
			return nil
		})
		if count != 0 {
			t.Errorf("Expected 0 news due to pagination skip, got %d", count)
		}
	})

	t.Run("Empty content after filtering", func(t *testing.T) {
		mock := &RefinedMockBrowser{
			FetchListFunc: func(ctx context.Context, url string) (string, error) {
				return listHTML, nil
			},
			FetchContentFunc: func(ctx context.Context, url string) (string, string, error) {
				return "   ", "", nil
			},
		}
		scraper.WithBrowser(mock)
		count := 0
		scraper.Scrape(context.Background(), time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), func(n *news.News) error {
			count++
			return nil
		})
		// If content is empty after filtering, Process returns error, and it's skipped
		if count != 0 {
			t.Errorf("Expected 0 news due to empty content, got %d", count)
		}
	})
}

func TestParseDate(t *testing.T) {
	dateStr := "05 Februari 2024"
	logger := zap.NewNop()
	scraper := NewScraper(logger, nil)
	parsed, err := scraper.parseDate(dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}

	expected := time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, parsed)
	}
}
