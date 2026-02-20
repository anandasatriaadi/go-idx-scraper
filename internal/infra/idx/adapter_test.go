package idx

import (
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/common"
)

func TestParseAPIResponse(t *testing.T) {
	judul := "Test Announcement"
	tgl := "2024-02-05T10:00:00"
	kode := "BBCA"
	path := "path/to/file.pdf"
	apiResp := APIResponse{
		Replies: []struct {
			Announcement AnnouncementResponse `json:"pengumuman"`
			Attachment   []common.Attachment  `json:"attachments"`
		}{
			{
				Announcement: AnnouncementResponse{
					ID:              "123",
					JudulPengumuman: &judul,
					TglPengumuman:   &tgl,
					CreatedDate:     &tgl,
					KodeEmiten:      &kode,
				},
				Attachment: []common.Attachment{
					{FullSavePath: &path},
				},
			},
		},
	}

	announcements, err := ParseAPIResponse(apiResp)
	if err != nil {
		t.Fatalf("Failed to parse API response: %v", err)
	}

	if len(announcements) != 1 {
		t.Errorf("Expected 1 announcement, got %d", len(announcements))
	}

	ann := announcements[0]
	if ann.ID != "123" {
		t.Errorf("Expected ID 123, got %s", ann.ID)
	}
	if *ann.KodeEmiten != "BBCA" {
		t.Errorf("Expected KodeEmiten BBCA, got %s", *ann.KodeEmiten)
	}

	expectedTime, _ := time.Parse("2006-01-02T15:04:05", tgl)
	if !ann.TglPengumuman.Equal(expectedTime) {
		t.Errorf("Expected time %v, got %v", expectedTime, ann.TglPengumuman)
	}

	if len(ann.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(ann.Attachments))
	}
}

func TestParseAPIResponse_Error(t *testing.T) {
	tgl := "invalid-date"
	apiResp := APIResponse{
		Replies: []struct {
			Announcement AnnouncementResponse `json:"pengumuman"`
			Attachment   []common.Attachment  `json:"attachments"`
		}{
			{
				Announcement: AnnouncementResponse{
					TglPengumuman: &tgl,
				},
			},
		},
	}

	_, err := ParseAPIResponse(apiResp)
	if err == nil {
		t.Error("Expected error for invalid date, got nil")
	}
}
