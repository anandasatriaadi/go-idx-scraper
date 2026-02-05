package helper

import (
	"fmt"
	"html/template"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/announcement"
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
		StockRows [][]StockData
		HasStocks bool
	}{
		ModeText:  modeText,
		Year:      config.Download.Year,
		DateTime:  fmt.Sprintf("%02d/%02d/%d %02d:%02d:%02d", now.Day(), int(now.Month()), now.Year(), now.Hour(), now.Minute(), now.Second()),
		HasStocks: len(stocks) > 0,
		StockRows: rows,
	}

	const htmlTemplate = `
	<!DOCTYPE html>
	<html lang="en">
			<head>
					<meta charset="UTF-8" />
					<meta name="viewport" content="width=device-width, initial-scale=1.0" />
					<title>New Stock Report</title>
					<style>
							/* Resets for Email Clients */
							body {
									margin: 0;
									padding: 0;
									min-width: 100%;
									background-color: #f0f0f0;
							}
							table {
									border-spacing: 0;
									font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
									color: #000000;
							}
							td {
									padding: 0;
							}
							img {
									border: 0;
							}
							.wrapper {
									width: 100%;
									table-layout: fixed;
									-webkit-text-size-adjust: 100%;
									-ms-text-size-adjust: 100%;
							}

							/* Modern Container */
							.report-card {
									max-width: 600px;
									margin: 0 auto;
									background-color: #ffffff;
									border: 1px solid #000000;
									box-shadow: none;
							}

							/* Utility Classes */
							.report-header {
									background-color: #000000;
									color: #ffffff;
									padding: 12px 24px;
									font-size: 14px;
									font-weight: 800;
									letter-spacing: 1px;
									text-transform: uppercase;
									text-align: center;
							}

							.date-time {
									color: #666666;
									font-size: 11px;
									font-family: monospace;
									letter-spacing: -0.5px;
									text-align: center;
									margin-bottom: 4px;
							}

							.subtitle {
									font-size: 12px;
									color: #888888;
									text-align: center;
									margin-bottom: 16px;
							}

							.stock-list {
									display: flex;
									flex-wrap: wrap;
							}

							.stock-item {
									flex: 0 0 25%;
									border-right: 1px dashed #e0e0e0;
							}

							.stock-link {
									display: block;
									text-decoration: none;
									color: #000000;
									padding: 8px 16px;
									text-align: center;
									font-weight: 700;
							}

							.stock-link:hover {
									color: #444444;
									text-decoration: underline;
							}

							.no-stocks {
									text-align: center;
									color: #888888;
									padding: 20px;
							}

							.footer {
									color: #767676;
									text-align: center;
									margin: 20px 0;
									font-size: 12px;
							}
					</style>
			</head>
			<body>
					<!-- This container simulates the email body -->
					<div style="background-color: #f0f0f0; padding: 40px 10px">
							<center class="wrapper">
									<!-- Global Header -->
									<div
											style="max-width: 600px; margin: 0 auto 20px auto; text-align: left"
									>
											<h1
													style="
															color: #000000;
															margin: 0;
															font-size: 24px;
															font-weight: 800;
															letter-spacing: -1px;
													"
											>
													ADI FAMILY MARKET<span style="font-weight: 300">WATCH</span>
											</h1>
											<p style="margin: 4px 0 0 0; font-size: 12px; color: #666666">
													New Stock Report
											</p>
									</div>

									<!-- Report Content -->
									<div class="report-card">
											<div class="report-header">
													New {{.ModeText}} {{.Year}} Stock Report
											</div>
											<div style="padding: 20px 24px">
													<span class="date-time">{{.DateTime}}</span>
													<p class="subtitle">
															Please check Google Drive, files has been minimized and
															uploaded
													</p>
													{{if .HasStocks}}
													<div class="stock-list">
															{{range .StockRows}}
															{{range .}}
															<div class="stock-item">
																	<a class="stock-link" href="{{.URL}}">{{.Name}}</a>
															</div>
															{{end}}
															{{end}}
													</div>
													{{else}}
													<p class="no-stocks">No new stocks found.</p>
													{{end}}
											</div>
									</div>
									<div class="footer">Adi Family Server &copy; 2023</div>
							</center>
					</div>
			</body>
	</html>
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

// AttachmentData represents data for a single attachment in the template.
type AttachmentData struct {
	URL      string
	Filename string
}

// ItemData represents data for a single announcement item in the template.
type ItemData struct {
	Date                 string
	Title                string
	MainLink             string
	SecondaryAttachments []AttachmentData
	HasBorder            bool
}

// GroupData represents data for a group of announcements by ticker.
type GroupData struct {
	Ticker string
	Items  []ItemData
}

// TemplateData represents the top-level data for the announcement email template.
type TemplateData struct {
	Groups []GroupData
}

// groupAnnouncements groups announcements by KodeEmiten and prepares data for the template.
func groupAnnouncements(announcements []*announcement.Announcement) []GroupData {
	grouped := make(map[string][]*announcement.Announcement)
	for _, ann := range announcements {
		if ann.KodeEmiten != nil {
			grouped[*ann.KodeEmiten] = append(grouped[*ann.KodeEmiten], ann)
		}
	}

	var groups []GroupData
	var tickers []string
	for ticker := range grouped {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)
	for _, ticker := range tickers {
		anns := grouped[ticker]
		var items []ItemData
		for i, ann := range anns {
			dateStr := ""
			if ann.TglPengumuman != nil {
				// Format as "13 DEC 2025" (uppercase month, matching JS example)
				dateStr = strings.ToUpper(ann.TglPengumuman.Format("02 Jan 2006"))
			}
			title := ""
			if ann.JudulPengumuman != nil {
				title = *ann.JudulPengumuman
			}
			mainLink := "#"
			var secondary []AttachmentData
			if len(ann.Attachments) > 0 {
				mainLink = *ann.Attachments[0].FullSavePath
				for j, att := range ann.Attachments {
					if j == 0 {
						continue
					}
					secondary = append(secondary, AttachmentData{
						URL:      *att.FullSavePath,
						Filename: *att.OriginalFilename,
					})
				}
			}
			items = append(items, ItemData{
				Date:                 dateStr,
				Title:                title,
				MainLink:             mainLink,
				HasBorder:            i < len(anns)-1, // Border for all but last item
				SecondaryAttachments: secondary,
			})
		}
		groups = append(groups, GroupData{
			Ticker: ticker,
			Items:  items,
		})
	}
	return groups
}

// GenerateAnnouncementEmail generates HTML content for an announcement email using the provided announcements.
func GenerateAnnouncementEmail(announcements []*announcement.Announcement) (string, error) {
	groups := groupAnnouncements(announcements)
	data := TemplateData{Groups: groups}

	const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Announcement Notification</title>
    <style>
        /* Resets for Email Clients */
        body { margin: 0; padding: 0; min-width: 100%; background-color: #f0f0f0; }
        table { border-spacing: 0; font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; color: #000000; }
        td { padding: 0; }
        img { border: 0; }
        .wrapper { width: 100%; table-layout: fixed; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; }

        /* Modern Container (Now applied per Group) */
        .group-card {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border: 1px solid #000000; /* Sharp black border */
            box-shadow: none;
            margin-bottom: 20px;
        }

        /* Utility Classes */
        .ticker-header {
            background-color: #000000;
            color: #ffffff;
            padding: 12px 24px;
            font-size: 14px;
            font-weight: 800;
            letter-spacing: 1px;
            text-transform: uppercase;
        }

        .date {
            color: #666666;
            font-size: 11px;
            font-family: monospace;
            letter-spacing: -0.5px;
            display: block;
            margin-bottom: 4px;
        }

        .main-link {
            font-size: 16px; /* Slightly smaller to accommodate list view */
            font-weight: 700;
            color: #000000;
            text-decoration: none;
            display: block;
            margin-bottom: 8px;
            line-height: 1.4;
        }
        .main-link:hover {
            color: #444444;
            text-decoration: underline;
        }

        .secondary-attachments {
            margin-top: 8px;
            padding-top: 8px;
            border-top: 1px dashed #e0e0e0;
        }
        .secondary-label {
            font-size: 10px;
            text-transform: uppercase;
            color: #888888;
            margin-bottom: 4px;
            display: block;
            font-weight: 600;
        }
        .attachment-link {
            font-size: 12px;
            color: #555555;
            text-decoration: none;
            display: block;
            padding: 2px 0;
            /* Removed white-space: nowrap; overflow: hidden; text-overflow: ellipsis; to allow wrapping */
        }
        .attachment-link:hover {
            color: #000000;
            background-color: #f9f9f9;
        }
        .icon-clip {
            display: inline-block;
            margin-right: 6px;
            font-size: 10px;
            color: #000000;
        }
    </style>
</head>
<body>
    <!-- This container simulates the email body -->
    <div style="background-color: #f0f0f0; padding: 40px 10px;">
        <center class="wrapper">

            <!-- Global Header -->
            <div style="max-width: 600px; margin: 0 auto 20px auto; text-align: left;">
                <h1 style="color: #000000; margin: 0; font-size: 24px; font-weight: 800; letter-spacing: -1px;">ADI FAMILY MARKET<span style="font-weight:300;">WATCH</span></h1>
                <p style="margin: 4px 0 0 0; font-size: 12px; color: #666666;">Daily Stock Announcements</p>
            </div>

            <!-- Dynamic Content Container -->
            <div id="email-content">
                {{range .Groups}}
                <div class="group-card">
                    <table width="100%" cellpadding="0" cellspacing="0">
                        <!-- Header Row for the Group -->
                        <tr>
                            <td class="ticker-header">
                                {{.Ticker}}
                            </td>
                        </tr>
                        {{range .Items}}
                        <tr>
                            <td style="padding: 20px 24px;{{if .HasBorder}} border-bottom: 1px solid #eeeeee;{{end}}">
                                <span class="date">{{.Date}}</span>

                                <a href="{{.MainLink}}" class="main-link" target="_blank">
                                    {{.Title}}
                                </a>

                                {{if .SecondaryAttachments}}
                                <div class="secondary-attachments">
                                    <span class="secondary-label">Attachments</span>
                                    {{range .SecondaryAttachments}}
                                    <a href="{{.URL}}" class="attachment-link" target="_blank">
                                        <span class="icon-clip">↳</span> {{.Filename}}
                                    </a>
                                    {{end}}
                                </div>
                                {{end}}
                            </td>
                        </tr>
                        {{end}}
                    </table>
                </div>
                {{end}}
            </div>

        </center>
    </div>
</body>
</html>
	`

	tmpl, err := template.New("announcement").Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing announcement template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing announcement template: %w", err)
	}

	return buf.String(), nil
}

// SendAnnouncementMail sends an announcement email with the provided content.
func SendAnnouncementMail(content string, config *config.Config) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "Stock Info <stockinfo@annd.dev>")
	m.SetHeader("To", config.Mail.List...)
	now := time.Now().In(time.FixedZone("GMT+8", 8*60*60))
	subject := fmt.Sprintf("Daily Stock Announcements - %02d/%02d/%d - %02d:%02d:%02d",
		now.Day(), int(now.Month()), now.Year(), now.Hour(), now.Minute(), now.Second(),
	)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	d := gomail.NewDialer("smtp.gmail.com", 587, config.Mail.Username, config.Mail.Password)

	// Retry sending the email up to 3 times
	for i := range 3 {
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Error sending announcement email (attempt %d): %v", i+1, err)
			if i == 2 {
				return fmt.Errorf("failed to send announcement email after 3 attempts")
			}
			time.Sleep(5 * time.Second) // Wait 5 seconds before retrying
		} else {
			fmt.Println("Announcement Mail Sent")
			return nil
		}
	}

	return nil
}
