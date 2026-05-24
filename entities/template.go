package entities

type Template struct {
	Name         string
	ContentPath  string
	Content      map[string]interface{}
	MasterLocked bool
	Args         map[string]Arg
}
