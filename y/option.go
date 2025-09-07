package y

type option int

const (
	UseAsync option = iota
	UseDistinct
	UsePanic
	NotNil
	NotEmpty
	Not
	Is

	// stl map
	RMap
	// isFlatFlex
	isFlatFlex
)

type OptionContext struct {
	async    bool
	distinct bool
}
