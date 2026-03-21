// Package idx implements the IDX API adapter for the announcement feature.
// It serves as an adapter (External Service implementation) in the Hexagonal Architecture,
// implementing the IDXDataProvider port defined in the announcement feature.
package idx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/common"
	"github.com/tebeka/selenium"
	"go.uber.org/zap"
)

type AnnouncementResponse struct {
	JudulPengumuman   *string `json:"JudulPengumuman" bson:"judul_pengumuman"`
	TglPengumuman     *string `json:"TglPengumuman"`
	EfekEmitenDinfra  *bool   `json:"EfekEmiten_DINFRA" bson:"efek_emiten_dinfra"`
	FinalId           *int    `json:"FinalId" bson:"final_id"`
	OldFinalId        *int    `json:"OldFinalId" bson:"old_final_id"`
	NoPengumuman      *string `json:"NoPengumuman" bson:"no_pengumuman"`
	KodeEmiten        *string `json:"Kode_Emiten" bson:"kode_emiten"`
	CreatedDate       *string `json:"CreatedDate"`
	EfekEmitenDire    *bool   `json:"EfekEmiten_DIRE" bson:"efek_emiten_dire"`
	FormId            *string `json:"Form_Id" bson:"form_id"`
	PerihalPengumuman *string `json:"PerihalPengumuman" bson:"perihal_pengumuman"`
	JmsxGroupId       *string `json:"JMSXGroupID" bson:"jmsx_group_id"`
	JenisPengumuman   *string `json:"JenisPengumuman" bson:"jenis_pengumuman"`
	ID                string  `json:"Id2" bson:"_id,omitempty"`
}

type APIResponse struct {
	Replies []struct {
		Announcement AnnouncementResponse `json:"pengumuman"`
		Attachment   []common.Attachment  `json:"attachments"`
	}
}

func ParseAPIResponse(apiResp APIResponse) ([]*announcement.Announcement, error) {
	var announcements []*announcement.Announcement
	for _, reply := range apiResp.Replies {
		// trim spaces
		kodeEmiten := reply.Announcement.KodeEmiten
		/*
			trimmedTicker := strings.Trim(*reply.Announcement.KodeEmiten, " ")
		*/

		ann := &announcement.Announcement{
			ID:                reply.Announcement.ID,
			EfekEmitenDire:    reply.Announcement.EfekEmitenDire,
			EfekEmitenDinfra:  reply.Announcement.EfekEmitenDinfra,
			FinalId:           reply.Announcement.FinalId,
			OldFinalId:        reply.Announcement.OldFinalId,
			NoPengumuman:      reply.Announcement.NoPengumuman,
			JudulPengumuman:   reply.Announcement.JudulPengumuman,
			JenisPengumuman:   reply.Announcement.JenisPengumuman,
			KodeEmiten:        kodeEmiten, // Keeping pointer logic
			FormId:            reply.Announcement.FormId,
			PerihalPengumuman: reply.Announcement.PerihalPengumuman,
			JmsxGroupId:       reply.Announcement.JmsxGroupId,
			Attachments:       reply.Attachment,
		}

		// Parse TglPengumuman
		if reply.Announcement.TglPengumuman != nil {
			t, err := time.Parse("2006-01-02T15:04:05", *reply.Announcement.TglPengumuman)
			if err != nil {
				return nil, err
			}
			ann.TglPengumuman = &t
		}
		// Parse CreatedDate
		if reply.Announcement.CreatedDate != nil {
			t, err := time.Parse("2006-01-02T15:04:05", *reply.Announcement.CreatedDate)
			if err != nil {
				return nil, err
			}
			ann.CreatedDate = &t
		}
		announcements = append(announcements, ann)
	}
	return announcements, nil
}

// IDXProvider implements the announcement.IDXDataProvider port.
// This is an External Service Adapter in Hexagonal Architecture.
type IDXProvider struct {
	logger *zap.Logger
	driver selenium.WebDriver
}

// NewIDXProvider creates a new IDX provider adapter.
func NewIDXProvider(logger *zap.Logger, driver selenium.WebDriver) announcement.IDXDataProvider {
	return &IDXProvider{
		logger: logger,
		driver: driver,
	}
}

// Fetch retrieves announcements from the IDX API for a given date range.
func (p *IDXProvider) Fetch(ctx context.Context, dateFrom, dateTo string) ([]*announcement.Announcement, error) {
	url := fmt.Sprintf(
		`https://www.idx.co.id/primary/ListedCompany/GetAnnouncement?kodeEmiten=&emitenType=s&indexFrom=0&pageSize=500&dateFrom=%s&dateTo=%s&lang=id&keyword=`,
		dateFrom, dateTo,
	)

	p.logger.Info("Fetching announcements from IDX", zap.String("url", url))

	if err := p.driver.Get(url); err != nil {
		return nil, fmt.Errorf("failed to navigate to IDX: %w", err)
	}

	time.Sleep(1 * time.Second)

	body, err := p.driver.FindElement(selenium.ByTagName, "body")
	if err != nil {
		return nil, fmt.Errorf("failed to find body element: %w", err)
	}

	data, err := body.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from body: %w", err)
	}

	var resp APIResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		p.logger.Error("Failed to unmarshal API response", zap.Error(err), zap.String("data", data))
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	announcements, err := ParseAPIResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	p.logger.Info("Successfully fetched announcements", zap.Int("count", len(announcements)))
	return announcements, nil
}
