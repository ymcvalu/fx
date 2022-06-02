package fx

import (
	"reflect"
	"sync"
)

type SpawnFunc func(Stream) Stream

func makeSpawn(it Iterator, n uint, fn SpawnFunc) Iterator {
	if n == 0 {
		n = 1
	}

	s := &spawn{
		it:      it,
		ch:      make(chan result),
		closeCh: make(chan struct{}),
		wg:      sync.WaitGroup{},
		fn:      fn,
	}

	chs := s.dispatch(n)
	s.fanin(chs)

	return s
}

type spawn struct {
	it      Iterator
	ch      chan result
	closeCh chan struct{}
	wg      sync.WaitGroup
	fn      SpawnFunc
}

func (s *spawn) Next() (Any, error) {
	r, has := <-s.ch
	if !has {
		return nil, errNone
	}
	return r.v, r.err
}

func (s *spawn) Close() {
	close(s.closeCh)
	s.wg.Wait()
	s.it.Close()
}

func (s *spawn) dispatch(n uint) []<-chan result {
	chs := make([]<-chan result, 0, n)

	safe := makeSaferead(s.it)
	for i := uint(0); i < n; i++ {
		ch := make(chan result)

		s.wg.Add(1)
		go func() {
			defer func() {
				s.wg.Done()
				close(ch)
			}()

			it := s.fn(&fx{
				it: safe,
			}).iter()
			defer it.Close()

			consumeIter(it, ch, s.closeCh)
		}()

		chs = append(chs, ch)
	}

	return chs
}

func (s *spawn) fanin(chs []<-chan result) {
	s.wg.Add(1)
	go func() {
		defer func() {
			s.wg.Done()
			close(s.ch)
		}()

		cases := make([]reflect.SelectCase, 0, len(chs))
		for _, c := range chs {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(c),
			})
		}

	loop:
		for len(cases) > 0 {
			i, v, ok := reflect.Select(cases)
			if !ok {
				cases = append(cases[:i], cases[i+1:]...)
				continue
			}
			select {
			case s.ch <- v.Interface().(result):
			case <-s.closeCh:
				break loop
			}
		}
	}()
}
