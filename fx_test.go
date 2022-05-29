package fx

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFx(t *testing.T) {
	list, err := From(Infinite()).
		Map(func(v Any) (Any, error) {
			return v.(uint64) * v.(uint64), nil
		}).
		Async(10).                        // async the previous map iter, and the size of chan buf is 10
		Spawn(10, func(s Stream) Stream { // spawn 10 goroutines to cousume the iter
			return s.
				Map(func(v Any) (Any, error) {
					time.Sleep(time.Millisecond * 100)
					return v, nil
				}).
				Filter(func(v Any) (bool, error) {
					return v.(uint64)%10 == 4, nil
				}).
				FlatMap(func(v Any) ([]Any, error) {
					iv := v.(uint64)
					return []Any{iv, iv + 1, iv + 2}, nil
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
			return s.Map(func(v Any) (Any, error) {
				return v.(int) * v.(int), nil
			})
		}).
		FlatMap(func(v Any) ([]Any, error) {
			iv := v.(int)
			return []Any{iv, iv + 1}, nil
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
		Map(func(v Any) (Any, error) {
			time.Sleep(time.Millisecond * 10)
			return v.(int) * v.(int), nil
		}).
		Async(10).
		Spawn(10, func(s Stream) Stream {
			return s.
				Map(func(v Any) (Any, error) {
					time.Sleep(time.Millisecond * 100)
					return v, nil
				}).
				Filter(func(v Any) (bool, error) {
					return v.(int)%10 == 4, nil
				}).
				FlatMap(func(v Any) ([]Any, error) {
					iv := v.(int)
					return []Any{iv, iv + 1, iv + 2}, nil
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
			Map(func(v Any) (Any, error) {
				return v.(uint64) * v.(uint64), nil
			}).
			Filter(func(v Any) (bool, error) {
				if v.(uint64)%10 == 1 {
					return false, fmt.Errorf("test err: %v", v)
				}
				return true, nil
			}).
			Take(10)
		assert.NotNil(t, err)
	})

	t.Run("skip err", func(t *testing.T) {
		list, err := From(Infinite()).
			Map(func(v Any) (Any, error) {
				return v.(uint64) * v.(uint64), nil
			}).
			Filter(func(v Any) (bool, error) {
				if v.(uint64)%10 == 1 {
					return false, fmt.Errorf("test err: %v", v)
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

func TestMapReduce(t *testing.T) {
	sum, err := From(Range(1, 100)).
		Map(func(v Any) (Any, error) {
			return v.(int) * 2, nil
		}).
		Reduce(0, func(sum, v Any) (Any, error) {
			return sum.(int) + v.(int), nil
		})
	assert.Nil(t, err)
	t.Log(sum)
}

func TestRecover(t *testing.T) {
	t.Run("ignore panic", func(t *testing.T) {
		sum, err := From(Range(1, 100)).
			Map(func(v Any) (Any, error) {
				return v.(int) * 2, nil
			}).
			Filter(func(v Any) (bool, error) {
				if v.(int) == 6 {
					panic("test")
				}
				return true, nil
			}).
			Recover(func(p interface{}) error {
				t.Log("panic found and skip", p)
				return nil // ignore panic
			}).
			Async(10).
			Reduce(0, func(sum, v Any) (Any, error) {
				return sum.(int) + v.(int), nil
			})
		assert.Nil(t, err)
		assert.Equal(t, 9894, sum)
		t.Log(sum)
	})

	t.Run("fail when panic", func(t *testing.T) {
		_, err := From(Range(1, 100)).
			Map(func(v Any) (Any, error) {
				return v.(int) * 2, nil
			}).
			Filter(func(v Any) (bool, error) {
				if v.(int) == 6 {
					panic("test")
				}
				return true, nil
			}).
			Recover(func(p interface{}) error {
				if e, is := p.(error); is {
					return fmt.Errorf("panic with %w", e)
				}
				return fmt.Errorf("panic with %v", p)
			}).
			Reduce(0, func(sum, v Any) (Any, error) {
				return sum.(int) + v.(int), nil
			})
		assert.NotNil(t, err)
		t.Log(err)
	})
}

func TestAdapter(t *testing.T) {
	t.Run("fn adapter", func(t *testing.T) {
		v := 0
		sum, err := FromFn(func() (int, error) {
			if v == 10 {
				return 0, None()
			}
			v++
			return v, nil
		}).Reduce(0, func(sum, v Any) (Any, error) {
			sum = sum.(int) + v.(int)
			return sum, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 55, sum)
	})

	t.Run("fn adater with err", func(t *testing.T) {
		testErr := errors.New("test ")
		_, err := FromFn(func() (interface{}, error) {
			return nil, errors.New("test ")

		}).Reduce(0, func(sum, v Any) (Any, error) {
			sum = sum.(int) + v.(int)
			return sum, nil
		})
		assert.NotNil(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("comp iter adapter", func(t *testing.T) {
		it1 := compIterWithClose{}
		sum1, err1 := FromCompIter(&it1).
			Reduce(0, func(sum, v Any) (Any, error) {
				sum = sum.(int) + v.(int)
				return sum, nil
			})
		assert.Nil(t, err1)
		assert.Equal(t, 55, sum1)
		assert.True(t, it1.closed)

		it2 := compIterWithoutClose{}
		sum2, err2 := FromCompIter(&it2).
			Reduce(0, func(sum, v Any) (Any, error) {
				sum = sum.(int) + v.(int)
				return sum, nil
			})
		assert.Nil(t, err2)
		assert.Equal(t, 55, sum2)
	})
}

type compIterWithClose struct {
	v      int
	closed bool
}

func (i *compIterWithClose) Next() (int, error) {
	if i.v == 10 {
		return 0, None()
	}
	i.v++
	return i.v, nil
}

func (i *compIterWithClose) Close() error {
	i.closed = true
	return nil
}

type compIterWithoutClose struct {
	v int
}

func (i *compIterWithoutClose) Next() (int, error) {
	if i.v == 10 {
		return 0, None()
	}
	i.v++
	return i.v, nil
}
