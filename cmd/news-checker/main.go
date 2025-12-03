package main

import (
	"context"
	"log"
	"os"

	"github.com/chromedp/chromedp"
)

// --- Main Logic ---

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.UserDataDir("/Users/ananda/gitrepo/stock-downloder/chromedp-profile"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.4758.102 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var data string
	// get page data as string
	if err := chromedp.Run(ctx, getPageData(`https://www.idx.co.id/primary/ListedCompany/GetAnnouncement?kodeEmiten=&emitenType=s&indexFrom=0&pageSize=100&dateFrom=19010101&dateTo=20251014&lang=id&keyword=`, &data)); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("data.json", []byte(data), 0o644); err != nil {
		log.Fatal(err)
	}

	log.Println("Starting report generation process...")
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
