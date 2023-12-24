package store

type Store interface {
	Find(key string) (interface{}, error)
	Save(key string, value interface{}) error
	Delete(key string) error
}
