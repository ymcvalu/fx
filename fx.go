package fx

import (
	"errors"
)

type Elem interface{}

type result struct {
	v   Elem
	err error
}

var (
	errNone = errors.New("end of iter")
)

func IsNone(err error) bool {
	return errors.Is(err, errNone)
}

type Fx interface {
	Iterator
	Stream
	Collect
}

type Iterator interface {
	Next() (Elem, error)
	Close()
}

type Stream interface {
	Map(MapFunc) Fx
	FlatMap(fn FlatMapFunc) Fx
	Filter(fn FilterFunc) Fx
	Async(bs uint) Fx
	Spawn(n uint, fn SpawnFunc) Fx
	OnError(fn ErrorFunc) Fx
	iter() Iterator
}

type ForEachFunc func(e Elem) error

type Collect interface {
	List(initCap uint) ([]Elem, error)
	ForEach(fn ForEachFunc) error
	Take(n uint) ([]Elem, error)
}

func From(it Iterator) Fx {
	return &fx{
		it: it,
	}
}

type fx struct {
	it Iterator
}

func (f *fx) Next() (Elem, error) {
	return f.it.Next()
}

func (f *fx) Close() {
	f.it.Close()
	f.it = nil
}

func (f *fx) Map(fn MapFunc) Fx {
	return &fx{
		it: makeMapping(f.it, fn),
	}
}

func (f *fx) FlatMap(fn FlatMapFunc) Fx {
	return &fx{
		it: makeFlatMap(f.it, fn),
	}
}

func (f *fx) Filter(fn FilterFunc) Fx {
	return &fx{
		it: makeFilter(f.it, fn),
	}
}

func (f *fx) Async(bs uint) Fx {
	return &fx{
		it: makeAsync(f.it, bs),
	}
}

func (f *fx) Spawn(n uint, fn SpawnFunc) Fx {
	return &fx{
		it: makeSpawn(f.it, n, fn),
	}
}

func (f *fx) OnError(fn ErrorFunc) Fx {
	return &fx{
		it: makeOnError(f.it, fn),
	}
}

func (f *fx) iter() Iterator {
	return f.it
}

func (f *fx) List(initCap uint) ([]Elem, error) {
	list := make([]Elem, 0, initCap)

	if err := f.ForEach(func(e Elem) error {
		list = append(list, e)
		return nil
	}); err != nil {
		return nil, err
	}

	return list, nil
}

func (f *fx) Take(n uint) ([]Elem, error) {
	defer f.Close()

	list := make([]Elem, 0, n)
	for i := uint(0); i < n; i++ {
		v, err := f.Next()
		if err != nil {
			if IsNone(err) {
				break
			}
			return nil, err
		}
		list = append(list, v)
	}

	return list, nil
}

func (f *fx) ForEach(fn ForEachFunc) error {
	defer f.Close()

	for {
		v, err := f.Next()
		if err != nil {
			if IsNone(err) {
				break
			}
			return err
		}

		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}
