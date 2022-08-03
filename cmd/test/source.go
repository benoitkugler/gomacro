package test

type S struct {
	D       string
	A, B, C int
}

type Union interface {
	isUnion()
}

func (S) isUnion() {}
