package y

type option int

const (
	UseAsync option = iota
	UseDistinct
	NotNil
	NotEmpty
	Not
	Is

	// stl map
	RMap
)

type OptionContext struct {
	async    bool
	distinct bool
}
