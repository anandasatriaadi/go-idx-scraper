package cmd

import (
	"github.com/spf13/cobra"
)

var (
	configPath string

	rootCmd = &cobra.Command{
		Use:   "go-idx-scraper",
		Short: "IDX Scraper and Tools",
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config/config.yml", "Path to configuration file")

	rootCmd.AddCommand(ServerCmd)
	rootCmd.AddCommand(ScrapeKontanCmd)
	rootCmd.AddCommand(CheckAnnouncementsCmd)
	rootCmd.AddCommand(UpdateIssuersCmd)
	rootCmd.AddCommand(DownloadReportsCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
