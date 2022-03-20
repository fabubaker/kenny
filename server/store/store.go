package store

type Store struct {
	Table map[string]string
}

func (store *Store) Get(key string) string {
	return store.Table[key]
}

func (store *Store) Put(key, value string) {
	store.Table[key] = value
}

func (store *Store) Delete(key string) string {
	value := store.Table[key]
	delete(store.Table, key)

	return value
}
