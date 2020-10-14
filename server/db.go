package server

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type memoryDB struct {
	items map[string]string
	mu    sync.RWMutex
}

func newDB() memoryDB {
	dbFile, err := os.Open("cypher-db.json")
	if err != nil {
		return memoryDB{items: map[string]string{}}
	}

	items := map[string]string{}
	if err := json.NewDecoder(dbFile).Decode(&items); err != nil {
		fmt.Println("Could not decode Json file", err.Error())
		return memoryDB{items: map[string]string{}}
	}

	return memoryDB{items: items}
}

func (m *memoryDB) set(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = value

}

func (m *memoryDB) get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, found := m.items[key]
	return value, found
}

func (m *memoryDB) delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}

func (m *memoryDB) save() {
	f, err := os.Create("cypher-db.json")
	if err != nil {
		fmt.Println("Could not create DB file", err.Error())
	}

	if err := json.NewEncoder(f).Encode(&m.items); err != nil {
		fmt.Println("Could not decode DB file", err.Error())
	}

	fmt.Println("Successfully saved items to DB file")
}
