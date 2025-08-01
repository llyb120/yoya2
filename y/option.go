package y

type option int

const (
	UseAsync option = iota
	UseDistinct
	NotNil
	NotEmpty
	Not
	Is
)

type OptionContext struct {
	async    bool
	distinct bool
}
