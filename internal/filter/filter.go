package filter

type Filter struct {
	name    string
	Ignore  []string
	Include []string
	Exclude []string
	Type    FilterType
	Cmd     []string
	Args    []string
	OnDir   bool
	Server  *Server
	Command *Command
}

type Server struct {
	*Filter
	Port       int
	Persistent bool
}

type Command struct {
	*Filter
	PathFlag    string
	OkExitCodes []int
}

type Tidier interface {
	Tidy(string) (string, error)
}

type Linter interface {
	Lint(string) (string, error)
}

func NewServer(
	name string,
	ignore, include, exclude []string,
	typ FilterType,
	cmd string,
	args []string,
	onDir bool,
	port int,
	persistent bool,
) *Filter {
	return &Filter{}
}

func NewCommand(
	name string,
	ignore, include, exclude []string,
	typ FilterType,
	cmd string,
	args []string,
	onDir bool,
	pathFlag string,
	okExitCodes []int,
) *Filter {
	return &Filter{}
}
