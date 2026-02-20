package common

type Attachment struct {
	PdfFilename      *string `json:"PDFFilename" bson:"pdf_filename"`
	FullSavePath     *string `json:"FullSavePath" bson:"full_save_path"`
	OriginalFilename *string `json:"OriginalFilename" bson:"original_filename"`
}
