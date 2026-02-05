package browser

import (
	"context"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/chromedp/chromedp"
)

func SetupChromeDp(cfg *config.Config) (ctx context.Context, cancel context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.Flag("disable-extensions", true),
		chromedp.UserDataDir(cfg.Paths.BrowserProfile),
		chromedp.Flag("log-level", "1"),
		chromedp.Flag("safebrowsing-disable-download-protection", true),
		chromedp.Flag("safebrowsing-disable-extension-blacklist", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"),
		chromedp.Flag("credentials_enable_service", false),
		chromedp.Flag("profile.password_manager_enabled", false),
	)

	// Create allocator context (manages browser process)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create task context (for running actions)
	ctx, cdpCancel := chromedp.NewContext(allocCtx)

	// Combined cancel function to clean up both contexts
	cancel = func() {
		cdpCancel()
		allocCancel()
	}
	return
}
