package fx

import "sync"

func makeSaferead(it Iterator) Iterator {
	return &saferead{
		it: it,
		mu: sync.Mutex{},
	}
}

type saferead struct {
	it Iterator
	mu sync.Mutex
}

func (s *saferead) Next() (Any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.it.Next()
}

func (s *saferead) Close() {
	// undo
}
