package fx

type ErrorFunc func(error) error

func makeOnError(it Iterator, fn ErrorFunc) Iterator {
	return &onError{
		it: it,
		fn: fn,
	}
}

type onError struct {
	it Iterator
	fn ErrorFunc
}

func (e *onError) Next() (Any, error) {
	for {
		v, err := e.it.Next()
		if err != nil {
			if IsNone(err) {
				return nil, err
			}

			if err = e.fn(err); err != nil {
				return nil, err
			}

			continue
		}
		return v, nil
	}
}

func (e *onError) Close() {
	e.it.Close()
}
