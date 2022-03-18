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

func (store *Store) Delete(key string) string {
	value := store.table[key]
	delete(store.table, key)

	return value
}
