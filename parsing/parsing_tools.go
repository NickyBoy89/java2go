package parsing

import (
	"io/fs"
	"os"
	"path/filepath"
)

const JavaExt = ".java"

func ReadSourcesInDir(directoryName string) ([]SourceFile, error) {
	sources := []SourceFile{}

	if _, err := os.Stat(directoryName); err != nil {
		return sources, err
	}

	if err := filepath.WalkDir(directoryName, fs.WalkDirFunc(
		func(path string, d fs.DirEntry, err error) error {

			// Only include java files
			if filepath.Ext(path) == JavaExt && !d.IsDir() {
				sourceCode, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				sources = append(sources, SourceFile{
					Name:   path,
					Source: sourceCode,
				})
			}

			return nil
		},
	)); err != nil {
		return nil, err
	}

	return sources, nil
}

func ParseASTs(file SourceFile) {

}
