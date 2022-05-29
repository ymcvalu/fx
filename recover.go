package fx

type RecoverFunc func(p interface{}) error

func makeOnPanic(it Iterator, fn RecoverFunc) Iterator {
	return &onPanic{
		it: it,
		fn: fn,
	}
}

type onPanic struct {
	it Iterator
	fn RecoverFunc
}

func (p *onPanic) Next() (Any, error) {
	for {
		result, has := p.next()
		if has {
			return result.v, result.err
		}
	}
}

func (p *onPanic) next() (r result, b bool) {
	defer func() {
		if pe := recover(); pe != nil {
			err := p.fn(pe)
			if err != nil {
				// return err
				r = result{err: err}
				b = true
			}
		}
		// ignore and skip
	}()

	v, err := p.it.Next()
	return result{v: v, err: err}, true
}

func (p *onPanic) Close() {
	p.it.Close()
}
