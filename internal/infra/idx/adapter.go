package idx

import (
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/common"
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
