package announcement

import (
	"context"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/common"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Announcement struct {
	ID                string              `json:"id" bson:"_id,omitempty"`
	EfekEmitenDire    *bool               `json:"efek_emiten_dire" bson:"efek_emiten_dire"`
	EfekEmitenDinfra  *bool               `json:"efek_emiten_dinfra" bson:"efek_emiten_dinfra"`
	FinalId           *int                `json:"final_id" bson:"final_id"`
	OldFinalId        *int                `json:"old_final_id" bson:"old_final_id"`
	NoPengumuman      *string             `json:"no_pengumuman" bson:"no_pengumuman"`
	TglPengumuman     *time.Time          `json:"tgl_pengumuman" bson:"tgl_pengumuman"`
	JudulPengumuman   *string             `json:"judul_pengumuman" bson:"judul_pengumuman"`
	JenisPengumuman   *string             `json:"jenis_pengumuman" bson:"jenis_pengumuman"`
	KodeEmiten        *string             `json:"kode_emiten" bson:"kode_emiten"`
	CreatedDate       *time.Time          `json:"created_date" bson:"created_date"`
	FormId            *string             `json:"form_id" bson:"form_id"`
	PerihalPengumuman *string             `json:"perihal_pengumuman" bson:"perihal_pengumuman"`
	JmsxGroupId       *string             `json:"jmsx_group_id" bson:"jmsx_group_id"`
	Attachments       []common.Attachment `json:"attachments" bson:"attachments"`
}

type Repository interface {
	Create(ctx context.Context, announcement *Announcement) error
	FindAll(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) ([]*Announcement, error)
	FindByID(ctx context.Context, id string) (*Announcement, error)
	Exists(ctx context.Context, id string) (bool, error)
}
