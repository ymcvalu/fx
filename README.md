# fx

a simple and easy function extension package, with synchronous and asynchronous support.

## Usage

```go
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
```

- async: make the previous iter execute asynchronously
- spawn: fan-out and fan-in
