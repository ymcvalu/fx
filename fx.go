package fx

import (
	"errors"
)

type Any interface{}

type result struct {
	v   Any
	err error
}

var (
	errNone = errors.New("end of iter")
)

func None() error {
	return errNone
}

func IsNone(err error) bool {
	return errors.Is(err, errNone)
}

type Fx interface {
	Iterator
	Stream
	Collect
}

type Iterator interface {
	Next() (Any, error)
	Close()
}

type Stream interface {
	Map(MapFunc) Fx
	FlatMap(fn FlatMapFunc) Fx
	Filter(fn FilterFunc) Fx
	Async(bs uint) Fx
	Spawn(n uint, fn SpawnFunc) Fx
	OnError(fn ErrorFunc) Fx
	Recover(fn RecoverFunc) Fx
	iter() Iterator
}

type ForEachFunc func(Any) (bool, error)
type ReduceFunc func(sum Any, v Any) (Any, error)

type Collect interface {
	ForEach(fn ForEachFunc) error
	Reduce(sum Any, fn ReduceFunc) (Any, error)
	List(initCap uint) ([]Any, error)
	Take(n uint) ([]Any, error)
}

func From(it Iterator) Fx {
	return &fx{
		it: it,
	}
}

type fx struct {
	it Iterator
}

func (f *fx) Next() (Any, error) {
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

func (f *fx) Recover(fn RecoverFunc) Fx {
	return &fx{
		it: makeOnPanic(f.it, fn),
	}
}

func (f *fx) iter() Iterator {
	return f.it
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

		if b, err := fn(v); err != nil {
			return err
		} else if !b {
			break
		}
	}
	return nil
}

func (f *fx) Reduce(init Any, fn ReduceFunc) (Any, error) {
	sum := init
	if err := f.ForEach(func(v Any) (bool, error) {
		_sum, err := fn(sum, v)
		if err != nil {
			return false, err
		}
		sum = _sum
		return true, nil
	}); err != nil {
		return nil, err
	}

	return sum, nil
}

func (f *fx) List(initCap uint) ([]Any, error) {
	list := make([]Any, 0, initCap)

	if err := f.ForEach(func(v Any) (bool, error) {
		list = append(list, v)
		return true, nil
	}); err != nil {
		return nil, err
	}

	return list, nil
}

func (f *fx) Take(n uint) ([]Any, error) {
	var (
		i    uint = 0
		list      = make([]Any, 0, n)
	)

	if err := f.ForEach(func(v Any) (bool, error) {
		if i >= n {
			return false, nil
		}
		list = append(list, v)
		i++
		return true, nil
	}); err != nil {
		return nil, err
	}

	return list, nil
}
