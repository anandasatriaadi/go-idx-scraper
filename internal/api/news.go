package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NewsRepository interface {
	FindAll(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOptions]) ([]*model.News, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.News, error)
}

func NewsRoutes(repo NewsRepository) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		filter := bson.M{}

		if startDateStr := r.URL.Query().Get("date_gte"); startDateStr != "" {
			if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
				filter["date"] = bson.M{"$gte": startDate}
			}
		}

		if endDateStr := r.URL.Query().Get("date_lte"); endDateStr != "" {
			if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
				if filter["date"] == nil {
					filter["date"] = bson.M{"$lte": endDate}
				} else {
					filter["date"].(bson.M)["$lte"] = endDate
				}
			}
		}

		if priorityStr := r.URL.Query().Get("priority"); priorityStr != "" {
			if priority, err := strconv.Atoi(priorityStr); err == nil {
				filter["priority"] = priority
			}
		}

		if source := r.URL.Query().Get("source"); source != "" {
			filter["link"] = source
		}

		news, err := repo.FindAll(r.Context(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(news)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		n, err := repo.FindByID(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(n)
	})

	return r
}
