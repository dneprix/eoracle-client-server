package storage

type Storage interface {
	Add(key string, value string)
	Get(key string) (string, bool)
	Delete(key string) bool
	GetAll() []KeyValue
}
