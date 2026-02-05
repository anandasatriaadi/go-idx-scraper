package kontan

import (
	"strings"
	"testing"
	"time"

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
	scraper := NewScraper(logger)
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

func TestParseDate(t *testing.T) {
	dateStr := "05 Februari 2024"
	logger := zap.NewNop()
	scraper := NewScraper(logger)
	parsed, err := scraper.parseDate(dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}

	expected := time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, parsed)
	}
}
