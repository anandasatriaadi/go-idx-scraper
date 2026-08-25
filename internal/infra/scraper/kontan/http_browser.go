package kontan

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// HTTPBrowser implements Browser using a lightweight, allocation-efficient net/http client with goquery
type HTTPBrowser struct {
	client *http.Client
}

// NewHTTPBrowser creates a high-performance HTTP browser adapter with optimized connection pooling
func NewHTTPBrowser() *HTTPBrowser {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     30,
		IdleConnTimeout:     60 * time.Second,
		DisableCompression: false,
		ForceAttemptHTTP2:   true,
	}
	return &HTTPBrowser{
		client: &http.Client{
			Timeout:   12 * time.Second,
			Transport: transport,
		},
	}
}

// FetchList retrieves and parses the article list HTML from Kontan index page
func (h *HTTPBrowser) FetchList(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating list request: %w", err)
	}
	h.setHeaders(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing list request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parsing list html: %w", err)
	}

	listElem := doc.Find("div.list-berita > ul")
	if listElem.Length() == 0 {
		// Fallback for alternate list selector
		listElem = doc.Find("div.list-berita")
	}

	html, err := listElem.Html()
	if err != nil {
		return "", fmt.Errorf("extracting list html: %w", err)
	}

	return html, nil
}

// FetchContent retrieves the article body and optional pagination from Kontan article page
func (h *HTTPBrowser) FetchContent(ctx context.Context, url string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("creating article request: %w", err)
	}
	h.setHeaders(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("executing article request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("parsing article html: %w", err)
	}

	// Extract primary article body
	articleNode := doc.Find(".tmpt-desk-kon")
	if articleNode.Length() == 0 {
		articleNode = doc.Find("[itemprop='articleBody']")
	}
	if articleNode.Length() == 0 {
		articleNode = doc.Find(".detail-berita")
	}

	articleHtml, err := articleNode.Html()
	if err != nil || strings.TrimSpace(articleHtml) == "" {
		return "", "", fmt.Errorf("article content element not found for %s", url)
	}

	// Check pagination
	var paginatedHtml string
	pagElem := doc.Find("div.pagination")
	if pagElem.Length() > 0 {
		paginatedHtml, _ = pagElem.Html()
	}

	return articleHtml, paginatedHtml, nil
}

func (h *HTTPBrowser) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
}
