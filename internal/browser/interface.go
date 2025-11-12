package browser

import (
	"github.com/tebeka/selenium"
)

// Service represents a browser driver service
type Service interface {
	Stop() error
}

// Browser represents a web browser session
type Browser interface {
	Get(url string) error
	Title() (string, error)
	Quit() error
}

// SeleniumService wraps selenium.Service to implement Service interface
type SeleniumService struct {
	*selenium.Service
}

// Stop implements the Service interface
func (s *SeleniumService) Stop() error {
	if s.Service != nil {
		return s.Service.Stop()
	}
	return nil
}

// SeleniumBrowser wraps selenium.WebDriver to implement Browser interface
type SeleniumBrowser struct {
	selenium.WebDriver
}

// Get implements the Browser interface
func (b *SeleniumBrowser) Get(url string) error {
	return b.WebDriver.Get(url)
}

// Title implements the Browser interface
func (b *SeleniumBrowser) Title() (string, error) {
	return b.WebDriver.Title()
}

// Quit implements the Browser interface
func (b *SeleniumBrowser) Quit() error {
	return b.WebDriver.Quit()
}
