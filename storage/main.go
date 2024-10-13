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
	crdtType   crdts.CRDT_TYPE
	heads      []crdts.SignedOperation
	value      interface{}
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

	newCell := StorageCell{operations: make([]crdts.SignedOperation, 1), crdtType: valueOpParsed.Type, heads: make([]crdts.SignedOperation, 1)}
	newCell.operations[0] = value
	newCell.heads[0] = value

	newCell.update()

	st.data[key] = newCell


	return nil
}

type GetResultDTO struct {
	Value interface{}
	Heads []crdts.SignedOperation
	Type  crdts.CRDT_TYPE
}

func (st *Storage) Get(key string) (val GetResultDTO, err error) {
	st.lock.RLock()
	defer st.lock.RUnlock()

	cell, exists := st.data[key]

	if !exists {
		return val, errors.New(fmt.Sprint("Storage cell with key ", key, " does not exist"))
	}

	return GetResultDTO{Value: cell.value, Type: cell.crdtType, Heads: cell.heads}, nil
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

	for _, pred := range valueOpParsed.Preds {
		found := false
		for _, op := range cell.operations {
			if pred == crdts.HashOperation(op) {
				found = true
			}
		}
		if !found {
			return errors.New("Attempted to append operation with unknown predecessors")
		}
	}

	cell.operations = append(cell.operations, newOp)
	cell.update()

	st.data[key] = cell

	return nil
}

func (c *StorageCell) update() {
	result := crdts.CalculateOperations(c.operations, c.crdtType)

	c.heads = result.Heads
	c.value = result.Value
}
