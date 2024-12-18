package ISyncMap

type ISyncMap[KEY comparable, VALUE any] interface {
	Get(KEY) (VALUE, error)
	Set(KEY, VALUE)
	Delete(KEY)
	GetKeys() (keys []KEY)
}
