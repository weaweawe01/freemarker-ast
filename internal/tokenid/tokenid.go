package tokenid

// Name returns token name by ID.
func Name(id int) (string, bool) {
	name, ok := IDToName[id]
	return name, ok
}

// ID returns token ID by name.
func ID(name string) (int, bool) {
	id, ok := NameToID[name]
	return id, ok
}
