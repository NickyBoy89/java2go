package symbol

var (
	GlobalScope = &GlobalSymbols{Packages: make(map[string]*PackageScope)}
)
