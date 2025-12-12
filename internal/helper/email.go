package helper

import (
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"gopkg.in/gomail.v2"
)

// getModeText returns the mode text based on config and romanPeriod.
func getModeText(config *config.Config, romanPeriod string) string {
	if config.Download.Mode == "AUDIT" {
		return "Tahunan"
	}
	return "Tri Wulan " + romanPeriod
}

// buildStockURL constructs the URL for a given stock.
func buildStockURL(stock string, config *config.Config, romanPeriod string) string {
	if config.Download.Mode == "AUDIT" {
		return fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/FinancialStatement-%s-Tahunan-%s.xlsx", config.Download.Year, "Audit", stock, config.Download.Year, stock)
	}
	return fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/FinancialStatement-%s-%s-%s.xlsx", config.Download.Year, config.Download.Mode, stock, config.Download.Year, romanPeriod, stock)
}

func GenerateNewReportEmail(stocks []string, romanPeriod string, config *config.Config) string {
	now := time.Now()
	modeText := getModeText(config, romanPeriod)

	// Prepare data for template
	type StockData struct {
		URL  string
		Name string
	}

	var stockList []StockData
	for _, stock := range stocks {
		stockList = append(stockList, StockData{
			URL:  buildStockURL(stock, config, romanPeriod),
			Name: stock,
		})
	}

	// Group stocks into rows of 4
	var rows [][]StockData
	var end int
	for i := 0; i < len(stockList); i += 4 {
		end = i + 4
		end = min(len(stockList), end)
		rows = append(rows, stockList[i:end])
	}

	data := struct {
		ModeText  string
		Year      string
		DateTime  string
		HasStocks bool
		StockRows [][]StockData
	}{
		ModeText:  modeText,
		Year:      config.Download.Year,
		DateTime:  fmt.Sprintf("%02d/%02d/%d %02d:%02d:%02d", now.Day(), int(now.Month()), now.Year(), now.Hour(), now.Minute(), now.Second()),
		HasStocks: len(stocks) > 0,
		StockRows: rows,
	}

	const htmlTemplate = `
	<div style="background-color: #f8f8f8; color: #262626; padding: 2rem">
		<h1 style="margin: 0 auto; text-align: center">
			New {{.ModeText}} {{.Year}} Stock Report
		</h1>
		<h2 style="margin: 0 auto 1rem; padding-bottom: 1rem; text-align: center">
			{{.DateTime}}
		</h2>
		<h3 style="margin: 0 auto 1rem; text-align: center; color: #969696;">
			Please check Google Drive, files has been minimized and uploaded
		</h3>
		{{if .HasStocks}}
		<table style="margin: auto; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.15); background-color: #fff;">
			{{range .StockRows}}
			<tr style="font-weight: bold">
				{{range .}}
				<td style="text-align: center; padding: 0.5rem 2rem">
					<a style="display: block; text-decoration: none; color: rgb(50, 106, 211); padding: 0.5rem 2rem;" href="{{.URL}}">{{.Name}}</a>
				</td>
				{{end}}
			</tr>
			{{end}}
		</table>
		{{else}}
		<p style="text-align: center; color: #969696;">No new stocks found.</p>
		{{end}}
		<div style="color: #767676; text-align: center; margin: 2rem 0 0">
			Adi Family Server &copy; 2023
		</div>
	</div>
	`

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		log.Printf("Error parsing email template: %v", err)
		return ""
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Error executing email template: %v", err)
		return ""
	}

	return buf.String()
}

func SendMail(content string, romanPeriod string, config *config.Config) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "Stock Info <stockinfo@annd.dev>")
	m.SetHeader("To", config.Mail.List...)
	now := time.Now()
	modeText := getModeText(config, romanPeriod)
	subject := fmt.Sprintf("New %s %s Stock Report 📈 - %02d:%02d:%02d",
		modeText,
		config.Download.Year,
		now.Hour(),
		now.Minute(),
		now.Second(),
	)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	d := gomail.NewDialer("smtp.gmail.com", 587, config.Mail.Username, config.Mail.Password)

	// Retry sending the email up to 3 times
	for i := range 3 {
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Error sending email (attempt %d): %v", i+1, err)
			if i == 2 {
				return fmt.Errorf("failed to send email after 3 attempts")
			}
			time.Sleep(5 * time.Second) // Wait 5 seconds before retrying
		} else {
			fmt.Println("Mail Sent")
			return nil
		}
	}

	return nil
}
