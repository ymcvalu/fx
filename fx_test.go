package fx

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFx(t *testing.T) {
	list, err := From(Infinite()).
		Map(func(e Elem) (Elem, error) {
			return e.(uint64) * e.(uint64), nil
		}).
		Async(10).                        // async the previous map iter, and the size of chan buf is 10
		Spawn(10, func(s Stream) Stream { // spawn 10 goroutines to cousume the iter
			return s.
				Map(func(e Elem) (Elem, error) {
					time.Sleep(time.Millisecond * 100)
					return e, nil
				}).
				Filter(func(e Elem) (bool, error) {
					return e.(uint64)%10 == 4, nil
				}).
				FlatMap(func(e Elem) ([]Elem, error) {
					v := e.(uint64)
					return []Elem{v, v + 1, v + 2}, nil
				})
		}).
		OnError(func(error) error {
			// log...
			return nil // skip the err
		}).
		Take(100)
	assert.Nil(t, err)
	assert.Len(t, list, 100)
	t.Log(list)
}

func TestNotClose(t *testing.T) {
	iter := From(Range(1, 10)).
		Spawn(10, func(s Stream) Stream {
			return s.Map(func(e Elem) (Elem, error) {
				return e.(int) * e.(int), nil
			})
		}).
		FlatMap(func(e Elem) ([]Elem, error) {
			v := e.(int)
			return []Elem{v, v + 1}, nil
		})

	list := make([]int, 0)
	for {
		e, err := iter.Next()
		if IsNone(err) {
			break
		}
		assert.Nil(t, err)
		list = append(list, e.(int))
	}

	_, err := iter.Next()
	assert.NotNil(t, err)
	assert.Equal(t, errNone, err)
	assert.Len(t, list, 18)
	t.Log(list)
}

func TestRange(t *testing.T) {
	start := time.Now()
	list, err := From(Range(1, 100)).
		Map(func(e Elem) (Elem, error) {
			time.Sleep(time.Millisecond * 10)
			return e.(int) * e.(int), nil
		}).
		Async(10).
		Spawn(10, func(s Stream) Stream {
			return s.
				Map(func(e Elem) (Elem, error) {
					time.Sleep(time.Millisecond * 100)
					return e, nil
				}).
				Filter(func(e Elem) (bool, error) {
					return e.(int)%10 == 4, nil
				}).
				FlatMap(func(e Elem) ([]Elem, error) {
					v := e.(int)
					return []Elem{v, v + 1, v + 2}, nil
				})
		}).
		List(100)
	timeUsed := time.Since(start)
	t.Logf("time used: %v", timeUsed)
	assert.Nil(t, err)
	t.Log(list)
}

func TestErr(t *testing.T) {
	t.Run("throw err", func(t *testing.T) {
		_, err := From(Infinite()).
			Map(func(e Elem) (Elem, error) {
				return e.(uint64) * e.(uint64), nil
			}).
			Filter(func(e Elem) (bool, error) {
				if e.(uint64)%10 == 1 {
					return false, fmt.Errorf("test err: %v", e)
				}
				return true, nil
			}).
			Take(10)
		assert.NotNil(t, err)
	})

	t.Run("skip err", func(t *testing.T) {
		list, err := From(Infinite()).
			Map(func(e Elem) (Elem, error) {
				return e.(uint64) * e.(uint64), nil
			}).
			Filter(func(e Elem) (bool, error) {
				if e.(uint64)%10 == 1 {
					return false, fmt.Errorf("test err: %v", e)
				}
				return true, nil
			}).
			OnError(func(err error) error {
				t.Log("err found: ", err)
				return nil
			}).
			Take(10)
		assert.Nil(t, err)
		assert.Len(t, list, 10)
	})
}
