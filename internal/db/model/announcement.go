package model

import (
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/types"
)

// Announcement represents the main document to be saved in MongoDB.
// It includes embedded Attachments.
//
//go:generate go run ../../../generator/mongogen/main.go -type=Announcement -collection=announcements
type Announcement struct {
	ID                string             `json:"id" bson:"_id,omitempty"`
	EfekEmitenDire    *bool              `json:"efek_emiten_dire" bson:"efek_emiten_dire"`
	EfekEmitenDinfra  *bool              `json:"efek_emiten_dinfra" bson:"efek_emiten_dinfra"`
	FinalId           *int               `json:"final_id" bson:"final_id"`
	OldFinalId        *int               `json:"old_final_id" bson:"old_final_id"`
	NoPengumuman      *string            `json:"no_pengumuman" bson:"no_pengumuman"`
	TglPengumuman     *time.Time         `json:"tgl_pengumuman" bson:"tgl_pengumuman"`
	JudulPengumuman   *string            `json:"judul_pengumuman" bson:"judul_pengumuman"`
	JenisPengumuman   *string            `json:"jenis_pengumuman" bson:"jenis_pengumuman"`
	KodeEmiten        *string            `json:"kode_emiten" bson:"kode_emiten"`
	CreatedDate       *time.Time         `json:"created_date" bson:"created_date"`
	FormId            *string            `json:"form_id" bson:"form_id"`
	PerihalPengumuman *string            `json:"perihal_pengumuman" bson:"perihal_pengumuman"`
	JmsxGroupId       *string            `json:"jmsx_group_id" bson:"jmsx_group_id"`
	Attachments       []types.Attachment `json:"attachments" bson:"attachments"`
}

// ParseAPIResponse converts APIResponse to a slice of Announcements, parsing dates and setting attachments.
func ParseAPIResponse(apiResp types.APIResponse) ([]*Announcement, error) {
	var announcements []*Announcement
	for _, reply := range apiResp.Replies {
		trimmedTicker := strings.Trim(*reply.Announcement.KodeEmiten, " ")
		ann := &Announcement{
			ID:                reply.Announcement.ID,
			EfekEmitenDire:    reply.Announcement.EfekEmitenDire,
			EfekEmitenDinfra:  reply.Announcement.EfekEmitenDinfra,
			FinalId:           reply.Announcement.FinalId,
			OldFinalId:        reply.Announcement.OldFinalId,
			NoPengumuman:      reply.Announcement.NoPengumuman,
			JudulPengumuman:   reply.Announcement.JudulPengumuman,
			JenisPengumuman:   reply.Announcement.JenisPengumuman,
			KodeEmiten:        &trimmedTicker,
			FormId:            reply.Announcement.FormId,
			PerihalPengumuman: reply.Announcement.PerihalPengumuman,
			JmsxGroupId:       reply.Announcement.JmsxGroupId,
			Attachments:       reply.Attachment,
		}
		// Parse TglPengumuman assuming YYYY-MM-DD format
		if reply.Announcement.TglPengumuman != nil {
			t, err := time.Parse("2006-01-02T15:04:05", *reply.Announcement.TglPengumuman)
			if err != nil {
				return nil, err
			}
			ann.TglPengumuman = &t
		}
		// Parse CreatedDate assuming RFC3339 format
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
