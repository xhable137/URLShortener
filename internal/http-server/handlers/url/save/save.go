package save

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"URLShortener/internal/lib/api/response"
	"URLShortener/internal/lib/logger/sl"
	"URLShortener/internal/lib/random"
)

// URLSaver is an interface for saving URLs.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(originalURL, shortCode string) error
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Status   string `json:"status"`
	ShortURL string `json:"short_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

func NewPostgresStorage(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.NewPostgresStorage"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		// Validate URL
		if req.URL == "" {
			log.Error("url is empty")
			render.JSON(w, r, response.Error("url is required"))
			return
		}

		_, err = url.ParseRequestURI(req.URL)
		if err != nil {
			log.Error("invalid url", sl.Err(err))
			render.JSON(w, r, response.Error("invalid url"))
			return
		}

		// Generate short code
		shortCode := random.NewRandomString(10)

		// Save URL
		err = urlSaver.SaveURL(req.URL, shortCode)
		if err != nil {
			log.Error("failed to save url", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		shortURL := "http://" + r.Host + "/" + shortCode

		log.Info("url saved", slog.String("url", req.URL), slog.String("short_code", shortCode))

		render.JSON(w, r, Response{
			Status:   "OK",
			ShortURL: shortURL,
		})
	}
}
