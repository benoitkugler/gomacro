package test

type concret1 struct {
	List2 []int
	V     int
}

type concret2 struct {
	D float64
}

type itfName interface {
	isI()
}

func (concret1) isI() {}
func (concret2) isI() {}

var (
	_ itfName = concret1{}
	_ itfName = concret2{}
)

type model struct {
	Value itfName
	Dict  map[int]int
	B     string
	L     ListV
	A     int
}

type ListV []itfName

type node struct {
	Children []node
}
