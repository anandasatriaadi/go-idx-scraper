package email

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"gopkg.in/gomail.v2"
)

func FindDownloadedStocks(config config.Config) []string {
	var foundStocks []string
	files, err := os.ReadDir(config.Paths.Download)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".xlsx" {
			parts := strings.Split(strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())), "-")
			foundStocks = append(foundStocks, parts[len(parts)-1])
		}
	}

	return foundStocks
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Sync to ensure write is complete
	return destFile.Sync()
}

func moveFile(src, dst string) error {
	// Try to rename (move) the file first
	err := os.Rename(src, dst)
	if err == nil {
		return nil // Successful move
	}

	// If rename fails, try copy and delete
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

func MoveFiles(config config.Config) error {
	files, err := filepath.Glob(filepath.Join(config.Paths.Download, "*.xlsx"))
	if err != nil {
		return fmt.Errorf("Error getting files: %v", err)
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		destPath := filepath.Join(config.Paths.Check, fileName)
		err := moveFile(file, destPath)
		if err != nil {
			log.Printf("Error moving file %s: %v", file, err)
		} else {
			log.Printf("Successfully moved file: %s", fileName)
		}
	}

	return nil
}

func GenerateMailContent(foundStocks []string, romanPeriod string, config config.Config) string {
	now := time.Now()
	modeText := "Tahunan"
	if config.Download.Mode != "AUDIT" {
		modeText = "Tri Wulan " + romanPeriod
	}

	content := fmt.Sprintf(`
    <div style="background-color: #f8f8f8; color: #262626; padding: 2rem">
        <h1 style="margin: 0 auto; text-align: center">
            New %s %s Stock Report
        </h1>
        <h2 style="margin: 0 auto 1rem; padding-bottom: 1rem; text-align: center">
            %02d/%02d/%d %02d:%02d:%02d
        </h2>
        <h3 style="margin: 0 auto 1rem; text-align: center; color: #969696;">
            Please check Google Drive, files has been minimized and uploaded
        </h3>
        <table style="margin: auto; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.15); background-color: #fff;">
            <tr style="font-weight: bold">
    `, modeText, config.Download.Year, now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second())

	for i, stock := range foundStocks {
		var url string
		if config.Download.Mode == "AUDIT" {
			url = fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/FinancialStatement-%s-Tahunan-%s.xlsx", config.Download.Year, "Audit", stock, config.Download.Year, stock)
		} else {
			url = fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/FinancialStatement-%s-%s-%s.xlsx", config.Download.Year, config.Download.Mode, stock, config.Download.Year, romanPeriod, stock)
		}

		content += fmt.Sprintf(`
            <td style="text-align: center; padding: 0.5rem 2rem">
                <a style="display: block; text-decoration: none; color: rgb(50, 106, 211); padding: 0.5rem 2rem;" href="%s">%s</a>
            </td>
        `, url, stock)

		if (i+1)%4 == 0 {
			content += `
                </tr>
                <tr style="font-weight: bold">
            `
		}
	}

	content += `
            </tr>
        </table>
        <div style="color: #767676; text-align: center; margin: 2rem 0 0">
            Adi Family Server &copy; 2023
        </div>
    </div>
    `

	return content
}

func SendMail(content string, romanPeriod string, config config.Config) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "Stock Info <stockinfo@annd.dev>")
	m.SetHeader("To", config.Mailing.List...)
	m.SetHeader("Subject", fmt.Sprintf("New %s %s Stock Report 📈 - %02d:%02d:%02d",
		func() string {
			if config.Download.Mode == "AUDIT" {
				return "Tahunan"
			}
			return "Tri Wulan " + romanPeriod
		}(),
		config.Download.Year,
		time.Now().Hour(),
		time.Now().Minute(),
		time.Now().Second(),
	))
	m.SetBody("text/html", content)

	d := gomail.NewDialer("smtp.gmail.com", 587, config.Mailing.SenderEmail, config.Mailing.SenderPassword)

	// Retry sending the email up to 3 times
	for i := 0; i < 3; i++ {
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Error sending email (attempt %d): %v", i+1, err)
			if i == 2 {
				return fmt.Errorf("Failed to send email after 3 attempts")
			}
			time.Sleep(5 * time.Second) // Wait 5 seconds before retrying
		} else {
			fmt.Println("Mail Sent")
			return nil
		}
	}

	return nil
}
