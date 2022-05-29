# fx

a function extension package, with synchronous and asynchronous support.

> Using type parameters as method receiver is unsupported, liking `func [T] (T) Xxx(...)`, and type's method must have no type parameters, so i can only use the fucking `interface{}`.

## Usage

### map and reduce

```go
  sum, err := From(Range(1, 100)).
    Map(func(v Any) (Any, error) {
      return v.(int) * 2, nil
    }).
    Reduce(0, func(sum, v Any) (Any, error) {
      return sum.(int) + v.(int), nil
    })
```

### async and spawn

```go
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
```

- async: make the previous iter execute asynchronously
- spawn: dispatch to multiple groutines and fan-in
