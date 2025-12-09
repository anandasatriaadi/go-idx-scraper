package types

import "github.com/anandasatriaadi/go-idx-scraper/internal/db/model"

type News struct {
	model.Announcement
	AnnouncementDate *string `json:"TglPengumuman"`
	CreatedDate      *string `json:"CreatedDate"`
}

type NewsResponse struct {
	Replies []struct {
		Announcement News               `json:"pengumuman"`
		Attachment   []model.Attachment `json:"attachments"`
	}
}
