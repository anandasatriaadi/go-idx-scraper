package xbrl

import (
	"io"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/xbrl/parser"
)

// ParseInstanceZip opens an instance.zip or inlineXBRL.zip file and parses the contents
func ParseInstanceZip(zipPath string) (*domain.Statement, error) {
	return parser.ParseInstanceZip(zipPath)
}

// ParseInstanceXML parses an XBRL instance stream into a domain Statement
func ParseInstanceXML(r io.Reader) (*domain.Statement, error) {
	return parser.ParseInstanceXML(r)
}
