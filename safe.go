package fx

import "sync"

func makeSafe(it Iterator) Iterator {
	return &safe{
		it: it,
		mu: sync.Mutex{},
	}
}

type safe struct {
	it Iterator
	mu sync.Mutex
}

func (s *safe) Next() (Any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.it.Next()
}

func (s *safe) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.it.Close()
}
