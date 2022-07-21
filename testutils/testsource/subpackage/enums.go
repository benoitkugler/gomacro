package subpackage

type Enum int

const (
	A Enum = iota
	B
	C
)

// gomacro:SQL add unique
type StructWithComment struct {
	A int
}
