package entities

type Definition struct {
	Templates map[string]Template
	Vars      map[string]Var
	Jsons     map[string]Json
}
