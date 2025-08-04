package y

type Tuple2[A, B any] struct {
	a A
	b B
}

func (t Tuple2[A, B]) Alpha() A {
	return t.a
}

func (t Tuple2[A, B]) Beta() B {
	return t.b
}

func T[A, B any](a A, b B) Tuple2[A, B] {
	return Tuple2[A, B]{
		a: a,
		b: b,
	}
}

type Tuple3[A, B, C any] struct {
	a A
	b B
	c C
}

func (t Tuple3[A, B, C]) Alpha() A {
	return t.a
}

func (t Tuple3[A, B, C]) Beta() B {
	return t.b
}

func (t Tuple3[A, B, C]) Gamma() C {
	return t.c
}

func T3[A, B, C any](a A, b B, c C) Tuple3[A, B, C] {
	return Tuple3[A, B, C]{
		a: a,
		b: b,
		c: c,
	}
}

type Tuple4[A, B, C, D any] struct {
	a A
	b B
	c C
	d D
}

func (t Tuple4[A, B, C, D]) Alpha() A {
	return t.a
}

func (t Tuple4[A, B, C, D]) Beta() B {
	return t.b
}

func (t Tuple4[A, B, C, D]) Gamma() C {
	return t.c
}

func (t Tuple4[A, B, C, D]) Delta() D {
	return t.d
}

func T4[A, B, C, D any](a A, b B, c C, d D) Tuple4[A, B, C, D] {
	return Tuple4[A, B, C, D]{
		a: a,
		b: b,
		c: c,
		d: d,
	}
}
