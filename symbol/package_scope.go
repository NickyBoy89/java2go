package symbol

// PackageScope represents a single package, which can contain one or more files
type PackageScope struct {
	// Maps the file's name to its definitions
	Files map[string]*FileScope
}

func (ps *PackageScope) ExcludeFile(excludedFileName string) *PackageScope {
	newScope := &PackageScope{Files: make(map[string]*FileScope)}
	for fileName, fileScope := range ps.Files {
		if fileName != excludedFileName {
			newScope.Files[fileName] = fileScope
		}
	}
	return newScope
}

func (ps *PackageScope) FindStaticField() Finder {
	pf := PackageFieldFinder(*ps)
	return &pf
}

type PackageFieldFinder PackageScope

func (pf *PackageFieldFinder) By(criteria func(d *Definition) bool) []*Definition {
	results := []*Definition{}
	for _, file := range pf.Files {
		for _, field := range file.BaseClass.Fields {
			if criteria(field) {
				results = append(results, field)
			}
		}
	}
	return results
}

func (ps *PackageFieldFinder) ByName(name string) []*Definition {
	return ps.By(func(d *Definition) bool {
		return d.Name == name
	})
}

func (ps *PackageFieldFinder) ByOriginalName(originalName string) []*Definition {
	return ps.By(func(d *Definition) bool {
		return d.Name == originalName
	})
}

func (ps *PackageScope) AddSymbolsFromFile(symbols *FileScope) {
	ps.Files[symbols.BaseClass.Class.Name] = symbols
}

// FindClass searches for a class in the given package and returns a scope for it
// the class may be the subclass of another class
func (ps *PackageScope) FindClass(name string) *ClassScope {
	for _, fileScope := range ps.Files {
		if fileScope.BaseClass.Class.OriginalName == name {
			return fileScope.BaseClass
		}
		for _, subclass := range fileScope.BaseClass.Subclasses {
			class := subclass.FindClass(name)
			if class != nil {
				return fileScope.BaseClass
			}
		}
	}
	return nil
}
