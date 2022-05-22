package fx

type MapFunc func(Elem) (Elem, error)

func makeMapping(it Iterator, fn MapFunc) Iterator {
	return &mapping{
		it: it,
		fn: fn,
	}
}

type mapping struct {
	it Iterator
	fn MapFunc
}

func (m *mapping) Next() (Elem, error) {
	v, err := m.it.Next()
	if err != nil {
		return nil, err
	}
	return m.fn(v)
}

func (m *mapping) Close() {
	m.it.Close()
}
