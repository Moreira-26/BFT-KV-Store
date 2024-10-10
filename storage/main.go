package storage

import (
	"bftkvstore/crdts"
	"errors"
	"fmt"
	"sync"
)

type Storage struct {
	lock sync.RWMutex
	data map[string]StorageCell // TODO: This needs to be stored in memory
}

type StorageCell struct {
	operations []crdts.SignedOperation
	crdtType   string
}

func Init() Storage {
	return Storage{
		data: make(map[string]StorageCell),
	}
}

func (st *Storage) Assign(key string, value crdts.SignedOperation) error {
	st.lock.Lock()
	defer st.lock.Unlock()

	_, exists := st.data[key]
	if exists {
		return errors.New("Tried to assign an already used key value")
	}

	valueOpParsed, err := crdts.ReadOperation(value)
	if err != nil {
		return errors.New("Failed to parse the given operation bytes")
	}

	newCell := StorageCell{operations: make([]crdts.SignedOperation, 1), crdtType: valueOpParsed.Type}
	newCell.operations[0] = value
	st.data[key] = newCell
	return nil
}

func (st *Storage) Get(key string) (val crdts.OpCalcResult, err error) {
	st.lock.RLock()
	defer st.lock.RUnlock()

	cell, exists := st.data[key]

	if !exists {
		return val, errors.New(fmt.Sprint("Storage cell with key", key, "does not exist"))
	}

	return crdts.CalculateOperations(cell.operations, cell.crdtType), nil
}

func (st *Storage) Append(key string, newOp crdts.SignedOperation) error {
	st.lock.Lock()
	defer st.lock.Unlock()

	cell, exists := st.data[key]
	if !exists {
		return errors.New("Tried to append to a non-existent key")
	}

	valueOpParsed, err := crdts.ReadOperation(newOp)
	if err != nil {
		return errors.New("Failed to parse the given operation bytes")
	}

	if valueOpParsed.Type != cell.crdtType {
		return errors.New("The given operation is not of the same type as the key storing")
	}

	cell.operations = append(cell.operations, newOp)
	st.data[key] = cell
	return nil
}
