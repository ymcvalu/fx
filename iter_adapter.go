package fx

import (
	"reflect"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

func FromFn(fn interface{}) Fx {
	return &fx{
		it: WrapFn(fn),
	}
}

func FromCompIter(it interface{}) Fx {
	return &fx{
		it: AdaptIter(it),
	}
}

type iterAdapter struct {
	next  reflect.Value
	close reflect.Value
}

func WrapFn(fn interface{}) Iterator {
	fnv := reflect.ValueOf(fn)
	mustBeNextFn(fnv)

	return &iterAdapter{
		next: fnv,
	}
}

func AdaptIter(v interface{}) Iterator {
	rv := reflect.ValueOf(v)
	nv := rv.MethodByName("Next")
	if !nv.IsValid() {
		panic("invalid type found because lack of Next func")
	}

	mustBeNextFn(nv)

	cv := rv.MethodByName("Close")
	if cv.IsValid() {
		mustBeCloseFn(cv)
	}

	return &iterAdapter{
		next:  nv,
		close: cv,
	}
}

func mustBeNextFn(fnv reflect.Value) {
	if fnv.Kind() != reflect.Func {
		panic("unsupport type for Next func")
	}

	fnt := fnv.Type()
	if fnt.NumIn() != 0 || fnt.NumOut() != 2 {
		panic("unsupport type for Next func")
	}

	r1t := fnt.Out(1)
	if !r1t.Implements(errType) {
		panic("unsupport type for Next func")
	}
}

func mustBeCloseFn(fnv reflect.Value) {
	if fnv.Kind() != reflect.Func {
		panic("unsupport type for Close func")
	}

	fnt := fnv.Type()
	if fnt.NumIn() != 0 {
		panic("unsupport type for Close func")
	}

}

func (it *iterAdapter) Next() (Any, error) {
	rs := it.next.Call(nil)
	if !rs[1].IsNil() {
		return nil, rs[1].Interface().(error)
	}
	return rs[0].Interface(), nil
}

func (it *iterAdapter) Close() {
	if it.close.IsValid() && !it.close.IsNil() {
		_ = it.close.Call(nil)
	}
}
