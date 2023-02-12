package symbol

var (
	// GlobalScope represents the global symbol table, and contains a mapping
	// between the package's path, and its symbols
	//
	// Example:
	// "net.java.math" -> Symbols { Vectors, Cos }
	GlobalScope = &GlobalSymbols{Packages: make(map[string]*PackageScope)}
)

// AddSymbolsToPackage adds a given file's symbols to the global package scope
func AddSymbolsToPackage(symbols *FileScope) {
	if _, exist := GlobalScope.Packages[symbols.Package]; !exist {
		GlobalScope.Packages[symbols.Package] = NewPackageScope()
	}
	GlobalScope.Packages[symbols.Package].Files[symbols.BaseClass.Class.Name] = symbols
}

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
