package fx

type FlatMapFunc func(Any) ([]Any, error)

func makeFlatMap(it Iterator, fn FlatMapFunc) Iterator {
	return &flatMap{
		it: it,
		fn: fn,
	}
}

type flatMap struct {
	it    Iterator
	fn    FlatMapFunc
	elems []Any
}

func (f *flatMap) Next() (Any, error) {
	for {
		if len(f.elems) > 0 {
			e := f.elems[0]
			f.elems = f.elems[1:]
			return e, nil
		}

		v, err := f.it.Next()
		if err != nil {
			return nil, err
		}

		elems, err := f.fn(v)
		if err != nil {
			return nil, err
		}
		f.elems = elems
	}
}

func (m *flatMap) Close() {
	m.it.Close()
}
