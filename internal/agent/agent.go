package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

const recordIDParam = "recordID"

type Agent struct {
	httpServer *http.Server
	log        *zap.Logger
}

func InitAgent(
	ctx context.Context,
	cfg config.AgentCfg,
	srvus models.UserStorage,
	srvrs models.RecordStorage,
	crs models.RecordStorage,
	log *zap.Logger) *Agent {
	h := handlers(srvus, srvrs, crs, log)
	a := &Agent{
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: initRouter(h),
		},
		log: log,
	}

	return a
}

func (a *Agent) ListenAndServe() error {
	if err := a.httpServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("an error occured while server listen and serve, err: %w", err)
		}
	}
	return nil
}

func (a *Agent) Shutdown(ctx context.Context) error {
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("an error occured while server shutdown, err: %w", err)
	}
	return nil
}

func initRouter(h *Handlers) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(h.requestLogger)

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			h.Register(r.Context(), w, r)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			h.Login(r.Context(), w, r)
		})

		r.Group(func(r chi.Router) {
			// r.Use(h.JwtMiddleware)

			var recordPath = "/records"
			var recordPathWithParam = path.Join(recordPath, "/{", recordIDParam, "}")

			r.Get(recordPath, func(w http.ResponseWriter, r *http.Request) {
				h.Records(r.Context(), w, r)
			})

			r.Get(recordPathWithParam, func(w http.ResponseWriter, r *http.Request) {
				h.Record(r.Context(), w, r)
			})

			r.Post(recordPath, func(w http.ResponseWriter, r *http.Request) {
				h.AddRecord(r.Context(), w, r)
			})

			r.Put(recordPath, func(w http.ResponseWriter, r *http.Request) {
				h.UpdateRecord(r.Context(), w, r)
			})

			r.Delete(recordPath, func(w http.ResponseWriter, r *http.Request) {
				h.DeleteRecord(r.Context(), w, r)
			})
		})
	})

	return router
}

func (h *Handlers) requestLogger(hr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewResponseLoggerWriter(w)
		var buf bytes.Buffer
		tee := io.TeeReader(r.Body, &buf)
		body, err := io.ReadAll(tee)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error("an error occured while logger read request body", zap.Error(err))
			return
		}
		r.Body = io.NopCloser(&buf)

		start := time.Now()
		hr.ServeHTTP(rw, r)
		duration := time.Since(start)

		h.log.Info("HTTP request",
			zap.String("method", r.Method),
			zap.Any("header", r.Header),
			zap.String("body", string(body)),
			zap.String("uri", r.RequestURI),
			zap.Duration("duration", duration),
			zap.Int("status", rw.responseData.status),
			zap.Int("size", rw.responseData.size),
		)
	})
}

type responseData struct {
	status int
	size   int
}

// ResponseLoggerWriter - stores the size of the response body and the response code.
type ResponseLoggerWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// NewResponseLoggerWriter - Object Constructor.
func NewResponseLoggerWriter(w http.ResponseWriter) *ResponseLoggerWriter {
	return &ResponseLoggerWriter{
		ResponseWriter: w,
		responseData:   &responseData{},
	}
}

func (r *ResponseLoggerWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("response logger writer err: %w", err)
	}

	r.responseData.size += size

	return size, nil
}

func (r *ResponseLoggerWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
