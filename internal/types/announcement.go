package types

type AnnouncementResponse struct {
	ID                string  `json:"Id2" bson:"_id,omitempty"`
	EfekEmitenDire    *bool   `json:"EfekEmiten_DIRE" bson:"efek_emiten_dire"`
	EfekEmitenDinfra  *bool   `json:"EfekEmiten_DINFRA" bson:"efek_emiten_dinfra"`
	FinalId           *int    `json:"FinalId" bson:"final_id"`
	OldFinalId        *int    `json:"OldFinalId" bson:"old_final_id"`
	NoPengumuman      *string `json:"NoPengumuman" bson:"no_pengumuman"`
	JudulPengumuman   *string `json:"JudulPengumuman" bson:"judul_pengumuman"`
	JenisPengumuman   *string `json:"JenisPengumuman" bson:"jenis_pengumuman"`
	KodeEmiten        *string `json:"Kode_Emiten" bson:"kode_emiten"`
	FormId            *string `json:"Form_Id" bson:"form_id"`
	PerihalPengumuman *string `json:"PerihalPengumuman" bson:"perihal_pengumuman"`
	JmsxGroupId       *string `json:"JMSXGroupID" bson:"jmsx_group_id"`
	TglPengumuman     *string `json:"TglPengumuman"`
	CreatedDate       *string `json:"CreatedDate"`
}

type Attachment struct {
	PdfFilename      *string `json:"PDFFilename" bson:"pdf_filename"`
	FullSavePath     *string `json:"FullSavePath" bson:"full_save_path"`
	OriginalFilename *string `json:"OriginalFilename" bson:"original_filename"`
}
type APIResponse struct {
	Replies []struct {
		Announcement AnnouncementResponse `json:"pengumuman"`
		Attachment   []Attachment         `json:"attachments"`
	}
}
