package storage

import (
	"sync"

	"github.com/google/uuid"
)

type Storage struct {
	lock sync.Mutex
	data map[string]string // TODO: This needs to be stored in memory
}

func (st *Storage) New() (uuid.UUID, error) {
	st.lock.Lock()
	defer st.lock.Unlock()

	return uuid.NewV7()
}

func (st *Storage) Put(id string) {
	st.lock.Lock()
	st.data[id] = id
	st.lock.Unlock()
}

func Update() {

}

func Delete() {

}
