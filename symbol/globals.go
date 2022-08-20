package symbol

var (
	// The global symbol table
	GlobalScope = &GlobalSymbols{Packages: make(map[string]*PackageScope)}
)

// A GlobalSymbols represents a global view of all the packages in the parsed source
type GlobalSymbols struct {
	// Every package's path associatedd with its definition
	Packages map[string]*PackageScope
}

func (gs *GlobalSymbols) String() string {
	result := ""
	for packageName := range gs.Packages {
		result += packageName + "\n"
	}
	return result
}

// FindPackage looks up a package's path in the global scope, and returns it
func (gs *GlobalSymbols) FindPackage(name string) *PackageScope {
	return gs.Packages[name]
}
