package store

type Store struct {
	Table map[string]map[string]string
}

func (store *Store) Get(key string, fields []string) map[string]string {
	subset := make(map[string]string)

	if len(fields) == 0 {
		return store.Table[key]
	}

	for _, field := range fields {
		value, ok := store.Table[key][field]
		if ok {
			subset[field] = value
		}
	}

	return subset
}

func (store *Store) Put(key string, values map[string]string) {
	existing, ok := store.Table[key]
	if !ok {
		store.Table[key] = values
		return
	}

	for field, value := range values {
		existing[field] = value
	}
}

func (store *Store) Delete(key string) {
	delete(store.Table, key)
}
