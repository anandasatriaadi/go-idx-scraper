package model

import (
	"time"
)

// Announcement represents the main document to be saved in MongoDB.
// It includes embedded Attachments.
//
//go:generate mongogen -type=Announcement -collection=announcements
type Announcement struct {
	ID                  string       `json:"Id2" bson:"_id,omitempty"` // MongoDB document ID
	IssuerEffectDire    *bool        `json:"EfekEmiten_DIRE" bson:"issuerEffectDire"`
	IssuerEffectDinfra  *bool        `json:"EfekEmiten_DINFRA" bson:"issuerEffectDinfra"`
	FinalId             *int         `json:"FinalId" bson:"finalId"`
	OldFinalId          *int         `json:"OldFinalId" bson:"oldFinalId"`
	AnnouncementNumber  *string      `json:"NoPengumuman" bson:"announcementNumber"`
	AnnouncementDate    *time.Time   `json:"TglPengumuman" bson:"announcementDate"`
	AnnouncementTitle   *string      `json:"JudulPengumuman" bson:"announcementTitle"`
	AnnouncementType    *string      `json:"JenisPengumuman" bson:"announcementType"`
	IssuerCode          *string      `json:"Kode_Emiten" bson:"issuerCode"`
	CreatedDate         *time.Time   `json:"CreatedDate" bson:"createdDate"`
	FormId              *string      `json:"Form_Id" bson:"formId"`
	AnnouncementSubject *string      `json:"PerihalPengumuman" bson:"announcementSubject"`
	JmsxGroupId         *string      `json:"JMSXGroupID" bson:"jmsxGroupId"`
	Attachments         []Attachment `json:"attachments" bson:"attachments"`
}

// Attachment represents each attachment in the announcement.
type Attachment struct {
	PdfFilename      *string `json:"PDFFilename" bson:"pdfFilename"`
	FullSavePath     *string `json:"FullSavePath" bson:"fullSavePath"`
	OriginalFilename *string `json:"OriginalFilename" bson:"originalFilename"`
}
