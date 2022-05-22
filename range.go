package fx

func Range(from, to int) Iterator {
	return &rng{
		from: from,
		to:   to,
	}
}

type rng struct {
	from int
	to   int
}

func (r *rng) Next() (Any, error) {
	if r.from >= r.to {
		return 0, errNone
	}

	v := r.from
	r.from++
	return v, nil
}

func (r *rng) Close() {
	r.from = r.to
}
