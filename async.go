package fx

import "sync"

func makeAsync(it Iterator, bs uint) Iterator {
	s := &async{
		it:      it,
		ch:      make(chan result, bs),
		closeCh: make(chan struct{}),
		wg:      sync.WaitGroup{},
	}

	s.start()

	return s
}

type async struct {
	it      Iterator
	ch      chan result
	closeCh chan struct{}
	wg      sync.WaitGroup
}

func (s *async) start() {
	s.wg.Add(1)
	go func() {
		defer func() {
			s.wg.Done()
			close(s.ch)
		}()

		consumeIter(s.it, s.ch, s.closeCh)
	}()
}

func (s *async) Next() (Elem, error) {
	r, has := <-s.ch
	if !has {
		return nil, errNone
	}
	return r.v, r.err
}

func (s *async) Close() {
	close(s.closeCh)
	s.wg.Wait()
	s.it.Close()
}

func consumeIter(it Iterator, ch chan<- result, closeCh <-chan struct{}) {
loop:
	for {
		v, err := it.Next()
		if err != nil && IsNone(err) {
			break loop
		}

		select {
		case <-closeCh:
			break loop
		case ch <- result{
			v:   v,
			err: err,
		}:
		}
	}
}
