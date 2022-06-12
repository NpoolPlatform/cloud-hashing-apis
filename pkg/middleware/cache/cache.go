package cache

import (
	"sync"
	"time"
)

const expireDuration = 4 * time.Hour

type entry struct {
	createAt time.Time
	value    interface{}
}

var (
	entries = map[string]*entry{}
	lock    sync.Mutex
)

func AddEntry(key string, value interface{}) {
	lock.Lock()
	defer lock.Unlock()
	entries[key] = &entry{
		createAt: time.Now(),
		value:    value,
	}
}

func DelEntry(key string) {
	lock.Lock()
	defer lock.Unlock()
	delete(entries, key)
}

func GetEntry(key string) interface{} {
	lock.Lock()
	defer lock.Unlock()
	if val, ok := entries[key]; ok {
		if time.Now().After(val.createAt.Add(expireDuration)) {
			return nil
		}
		return val.value
	}
	return nil
}
