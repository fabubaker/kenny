package main

type Store struct {
	table map[string]string
}

func (store *Store) Get(key string) string {
	return store.table[key]
}

func (store *Store) Put(key, value string) {
	store.table[key] = value
}
