package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

const contentTypeJSON = "application/json"
const contentType = "Content-Type"

var errUserUndefined = "user undefined"

type UserRecordStorage interface {
	AddUserRecordStorage(userID string) error
	RemoveUserRecordStorage(userID string) error
}

type Handlers struct {
	srvRecordStorage   models.RecordStorage
	srvUserStorage     models.UserStorage
	userRecordStorage  UserRecordStorage
	cacheRecordStorage models.RecordStorage
	log                *zap.Logger
}

func handlers(srvus models.UserStorage,
	srvrs models.RecordStorage,
	crs models.RecordStorage,
	log *zap.Logger) *Handlers {
	return &Handlers{
		srvRecordStorage:   srvrs,
		srvUserStorage:     srvus,
		cacheRecordStorage: crs,
		log:                log,
	}
}

func (h *Handlers) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := h.getLoginPsw(w, r)
	if err != nil {
		h.log.Error("failed to read the Register request body", zap.Error(err))
		return
	}

	user, err := h.addUser(ctx, w, u)
	if err != nil {
		h.log.Error("an error occured while registering user", zap.Error(err))
		return
	}

	// TODO jwt

	h.writeResponse(w, user)
}

func (h *Handlers) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := h.getLoginPsw(w, r)
	if err != nil {
		h.log.Error("failed to read the Login request", zap.Error(err))
		return
	}

	_, err = h.getUser(ctx, w, u)
	if err != nil {
		h.log.Error("failed to get user the Login request", zap.Error(err))
		return
	}

	// TODO add check jwt
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) Records(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, ok := userFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error(errUserUndefined)
		return
	}

	records, err := u.GetRecords(ctx, h.cacheRecordStorage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error("an error occured while retrieving records from cache storage", zap.Error(err))
		return
	}

	if len(records) != 0 {
		h.writeResponse(w, records)
		return
	}

	records, err = u.GetRecords(ctx, h.srvRecordStorage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error("an error occured while retrieving records from server storage", zap.Error(err))
		return
	}

	h.writeResponse(w, records)
}

func (h *Handlers) Record(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, ok := userFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error(errUserUndefined)
		return
	}

	recordID := chi.URLParam(r, recordIDParam)

	record, err := u.GetRecord(ctx, h.cacheRecordStorage, recordID)
	if err != nil {
		if !errors.Is(err, models.ErrRecordNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Error("an error occured while retrieving record from cache storage", zap.Error(err))
			return
		}

		record, err = u.GetRecord(ctx, h.srvRecordStorage, recordID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Error("an error occured while retrieving record from server storage", zap.Error(err))
			return
		}
	}

	h.writeResponse(w, record)
}

func (h *Handlers) AddRecord(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.addOrUpdateRecord(ctx, w, r, false)
}

func (h *Handlers) UpdateRecord(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.addOrUpdateRecord(ctx, w, r, true)
}

func (h *Handlers) addOrUpdateRecord(ctx context.Context, w http.ResponseWriter, r *http.Request, update bool) {
	u, ok := userFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error(errUserUndefined)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error("an error occured while reading update request body", zap.Error(err))
		return
	}

	if update {
		record := &models.Record{}
		if err := json.Unmarshal(b, &record); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Error("an error occured while unmarshal incoming update request", zap.Error(err))
			return
		}

		updated, err := u.UpdateRecord(ctx, h.cacheRecordStorage, record)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Error("an error occurred while updating an entry to the server storage", zap.Error(err))
			return
		}

		h.writeResponse(w, updated)
	} else {
		rdto := &models.RecordDTO{}
		if err := json.Unmarshal(b, &r); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Error("an error occured while unmarshal incoming add request", zap.Error(err))
			return
		}

		record, err := u.AddRecord(ctx, h.cacheRecordStorage, rdto)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Error("an error occurred while adding an entry to the server storage", zap.Error(err))
			return
		}

		h.writeResponse(w, record)
	}
}

func (h *Handlers) DeleteRecord(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, ok := userFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error(errUserUndefined)
		return
	}

	recordID := chi.URLParam(r, recordIDParam)

	err := u.DeleteRecord(ctx, h.cacheRecordStorage, recordID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error("an error occured while removing record from storage", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) getLoginPsw(w http.ResponseWriter, r *http.Request) (*models.UserDTO, error) {
	var u models.UserDTO

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("failed get login and password from body err: %w", err)
	}

	if err := json.Unmarshal(b, &u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("failed unmarhsal login and password err: %w", err)
	}

	if u.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("login is empty")
	}

	return &u, nil
}

func (h *Handlers) getUser(ctx context.Context, w http.ResponseWriter, u *models.UserDTO) (*models.User, error) {
	user, err := u.GetUser(ctx, h.srvUserStorage)
	if err != nil {
		if errors.Is(err, models.ErrUnknowUser) {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, fmt.Errorf("failed to authorize the user, err: %w ", err)
		}

		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("error getting user: %w ", err)
	}

	if err := h.userRecordStorage.AddUserRecordStorage(user.ID); err != nil {
		return nil, fmt.Errorf("an error occured while retrieving user cache-storage, err: %w", err)
	}

	// TODO Run sync with server

	return user, nil
}

func (h *Handlers) addUser(ctx context.Context, w http.ResponseWriter, u *models.UserDTO) (*models.User, error) {
	user, err := u.AddUser(ctx, h.srvUserStorage)
	if err != nil {
		if errors.Is(err, models.ErrLoginIsBusy) {
			w.WriteHeader(http.StatusConflict)
			return nil, models.ErrLoginIsBusy
		}

		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("an error occured while register user in server storage, err: %w", err)
	}

	if err := h.userRecordStorage.AddUserRecordStorage(user.ID); err != nil {
		return nil, fmt.Errorf("an error occured while create user cache-storage, err: %w", err)
	}

	return user, nil
}

func (h *Handlers) writeResponse(w http.ResponseWriter, resp any) {
	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error("an error occured while marshal response to json", zap.Error(err))
	}

	w.Header().Set(contentType, contentTypeJSON)

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error("an error occured while write response", zap.Error(err))
	}
}
