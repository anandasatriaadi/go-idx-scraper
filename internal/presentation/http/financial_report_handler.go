package http

import (
	"encoding/json"
	"net/http"

	"github.com/anandasatriaadi/go-idx-scraper/internal/application/financialreport"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FinancialReportHandler struct {
	service *financialreport.Service
}

func NewFinancialReportHandler(service *financialreport.Service) *FinancialReportHandler {
	return &FinancialReportHandler{service: service}
}

func (h *FinancialReportHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetAll)
	return r
}

func (h *FinancialReportHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.FindAll(r.Context(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}
