package fx

type FilterFunc func(Any) (bool, error)

func makeFilter(it Iterator, fn FilterFunc) Iterator {
	return &filter{
		it: it,
		fn: fn,
	}
}

type filter struct {
	it Iterator
	fn FilterFunc
}

func (f *filter) Next() (Any, error) {
	for {
		v, err := f.it.Next()
		if err != nil {
			return nil, err
		}

		if b, err := f.fn(v); err != nil {
			return nil, err
		} else if b {
			return v, nil
		}
	}
}

func (f *filter) Close() {
	f.it.Close()
}
