package mem

import (
	"sync"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

type MemStorage struct {
	mutex *sync.RWMutex
	data  map[string]*UserRecordStorage
}

type UserRecordStorage struct {
	mutex *sync.RWMutex
	data  map[string]*models.Record
}

func NewMemStorage() *MemStorage {
	ms := &MemStorage{
		mutex: &sync.RWMutex{},
		data:  make(map[string]*UserRecordStorage),
	}

	return ms
}

// AddUserRecordStorage - create user storage in cache.
func (ms *MemStorage) AddUserRecordStorage(userID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	rs := &UserRecordStorage{
		mutex: &sync.RWMutex{},
		data:  make(map[string]*models.Record),
	}

	ms.data[userID] = rs

	return nil
}

// RemoveUserRecordStorage - remove user storage from cache.
func (ms *MemStorage) RemoveUserRecordStorage(userID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	delete(ms.data, userID)

	return nil
}
