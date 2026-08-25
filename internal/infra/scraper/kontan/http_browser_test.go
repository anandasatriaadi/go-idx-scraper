package kontan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPBrowser_FetchListAndContent(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/list" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<html>
					<body>
						<div class="list-berita">
							<ul>
								<li><a href="/news/item-1">News Item 1</a></li>
								<li><a href="/news/item-2">News Item 2</a></li>
							</ul>
						</div>
					</body>
				</html>
			`)
			return
		}

		if r.URL.Path == "/article" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<html>
					<body>
						<div class="tmpt-desk-kon" itemprop="articleBody">
							<p>This is the test article body content.</p>
						</div>
						<div class="pagination">
							<a href="/article?page=2">2</a>
						</div>
					</body>
				</html>
			`)
			return
		}

		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	browser := NewHTTPBrowser()
	ctx := context.Background()

	// Test FetchList
	listHTML, err := browser.FetchList(ctx, mockServer.URL+"/list")
	if err != nil {
		t.Fatalf("FetchList failed: %v", err)
	}
	if len(listHTML) == 0 {
		t.Errorf("Expected non-empty list HTML")
	}

	// Test FetchContent
	content, pagination, err := browser.FetchContent(ctx, mockServer.URL+"/article")
	if err != nil {
		t.Fatalf("FetchContent failed: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("Expected non-empty article content")
	}
	if len(pagination) == 0 {
		t.Errorf("Expected non-empty pagination HTML")
	}
}
