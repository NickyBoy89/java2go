package symbol

// FileScope represents the scope in a single source file, that can contain one
// or more source classes
type FileScope struct {
	// The global package that the file is located in
	Package string
	// Every external package that is imported into the file
	// Formatted as map[ImportedType: full.package.path]
	Imports map[string]string
	// The base class that is in the file
	BaseClass *ClassScope
}

// FindClass searches through a file to find if a given class has been defined
// at its root class, or within any of the subclasses
func (fs *FileScope) FindClass(name string) *Definition {
	if def := fs.BaseClass.FindClass(name); def != nil {
		return def
	}
	for _, subclass := range fs.BaseClass.Subclasses {
		if def := subclass.FindClass(name); def != nil {
			return def
		}
	}
	return nil
}
