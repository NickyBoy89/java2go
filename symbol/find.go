package symbol

// Finder represents an object that can search through its contents for a given
// list of definitions that match a certian criteria
type Finder interface {
	ByName(name string) []*Definition
	ByOriginalName(originalName string) []*Definition
}
