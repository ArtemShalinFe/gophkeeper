package mem

import (
	"context"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/google/uuid"
)

func (ms *MemStorage) List(ctx context.Context, userID string) ([]*models.Record, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	us, ok := ms.data[userID]
	if !ok {
		return nil, models.ErrUserStorageNotFound
	}

	us.mutex.RLock()
	defer us.mutex.RUnlock()

	rs := make([]*models.Record, len(us.data))
	i := 0
	for _, r := range us.data {
		rs[i] = r
		i++
	}

	return rs, nil
}

func (ms *MemStorage) Get(ctx context.Context, userID string, recordID string) (*models.Record, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	us, ok := ms.data[userID]
	if !ok {
		return nil, models.ErrUserStorageNotFound
	}

	us.mutex.RLock()
	defer us.mutex.RUnlock()

	r, ok := us.data[recordID]
	if !ok {
		return nil, models.ErrRecordNotFound
	}

	return r, nil
}

func (ms *MemStorage) Add(ctx context.Context, userID string, record *models.RecordDTO) (*models.Record, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	us, ok := ms.data[userID]
	if !ok {
		return nil, models.ErrUserStorageNotFound
	}

	us.mutex.Lock()
	defer us.mutex.Unlock()
	id := uuid.New().String()

	now := time.Now()

	r := &models.Record{
		ID:          id,
		Owner:       userID,
		Description: record.Description,
		Type:        record.Type,
		Created:     now,
		Modified:    now,
		Data:        record.Data,
		Hashsum:     record.Hashsum,
		Metainfo:    record.Metainfo,
		Version:     1,
	}

	us.data[id] = r

	return r, nil
}

func (ms *MemStorage) Update(ctx context.Context, userID string, record *models.Record) (*models.Record, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	us, ok := ms.data[userID]
	if !ok {
		return nil, models.ErrUserStorageNotFound
	}

	us.mutex.Lock()
	defer us.mutex.Unlock()

	record.Modified = time.Now()
	us.data[record.ID] = record

	return record, nil
}

func (ms *MemStorage) Delete(ctx context.Context, userID string, recordID string) error {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	us, ok := ms.data[userID]
	if !ok {
		return models.ErrUserStorageNotFound
	}

	us.mutex.Lock()
	defer us.mutex.Unlock()
	r, ok := us.data[recordID]
	if !ok {
		return models.ErrRecordNotFound
	}

	r.Deleted = true
	r.Modified = time.Now()

	return nil
}
