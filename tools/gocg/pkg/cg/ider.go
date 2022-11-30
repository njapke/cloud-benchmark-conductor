package cg

import (
	"fmt"
	"strings"
	"sync"
)

type IDer struct {
	idCount  int64
	idToName map[int64]string
	nameToID map[string]int64
	l        sync.RWMutex
}

func NewIDer() *IDer {
	return &IDer{
		idToName: make(map[int64]string),
		nameToID: make(map[string]int64),
	}
}

func (i *IDer) ID(name string) int64 {
	i.l.Lock()
	defer i.l.Unlock()

	name = strings.TrimSpace(name)

	id, ok := i.nameToID[name]
	if !ok {
		// create
		newID := i.idCount
		i.idCount++
		i.idToName[newID] = name
		i.nameToID[name] = newID
		return newID
	}
	return id
}

func (i *IDer) Name(id int64) (string, error) {
	return i.internalName(id)
}

func (i *IDer) internalName(id int64) (string, error) {
	i.l.RLock()
	defer i.l.RUnlock()

	name, ok := i.idToName[id]
	if !ok {
		return "", fmt.Errorf("unknown id %d", id)
	}
	return name, nil
}

func (i *IDer) Copy() *IDer {
	i.l.RLock()
	defer i.l.RUnlock()

	newIDToName := make(map[int64]string)
	for k, v := range i.idToName {
		newIDToName[k] = v
	}

	newNameToID := make(map[string]int64)
	for k, v := range i.nameToID {
		newNameToID[k] = v
	}

	return &IDer{
		idCount:  i.idCount,
		idToName: newIDToName,
		nameToID: newNameToID,
	}
}
