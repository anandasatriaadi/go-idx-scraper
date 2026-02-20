package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AnnouncementHandler struct {
	service *announcement.Service
}

func NewAnnouncementHandler(service *announcement.Service) *AnnouncementHandler {
	return &AnnouncementHandler{service: service}
}

func (h *AnnouncementHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.GetByID)
	return r
}

func (h *AnnouncementHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	announcements, err := h.service.FindAll(r.Context(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

func (h *AnnouncementHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ann, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ann)
}
