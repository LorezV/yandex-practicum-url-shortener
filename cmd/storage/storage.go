package storage

var Repository URLRepository

func MakeRepository() URLRepository {
	return URLRepository{storage: make(map[string]URL)}
}

type URL struct {
	ID       string
	Original string
	Short    string
}

type URLRepository struct {
	storage map[string]URL
}

func (r URLRepository) GetAll() []URL {
	b := make([]URL, len(r.storage))
	i := 0
	for _, value := range r.storage {
		b[i] = value
		i++
	}
	return b
}

func (r URLRepository) Get(id string) (URL, bool) {
	val, ok := r.storage[id]
	return val, ok
}

func (r URLRepository) Add(url URL) bool {
	_, ok := r.storage[url.ID]
	if !ok {
		r.storage[url.ID] = url
	}
	return !ok
}
