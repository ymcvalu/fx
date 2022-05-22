package fx

func Infinite() Iterator {
	return &infinite{}
}

type infinite struct {
	v uint64
}

func (i *infinite) Next() (Any, error) {
	v := i.v
	i.v++
	return v, nil
}

func (i *infinite) Close() {
}
