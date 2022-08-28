package main

import (
	"strconv"

	"github.com/NickyBoy89/java2go/symbol"
)

func ResolveFile(file SourceFile) {
	ResolveClass(file.Symbols.BaseClass, file)
	for _, subclass := range file.Symbols.BaseClass.Subclasses {
		ResolveClass(subclass, file)
	}
}

func ResolveClass(class *symbol.ClassScope, file SourceFile) {
	// Resolve all the fields in that respective class
	for _, field := range class.Fields {

		// Since a private global variable is able to be accessed in the package, it must be renamed
		// to avoid conflicts with other global variables

		packageScope := symbol.GlobalScope.FindPackage(file.Symbols.Package)

		symbol.ResolveDefinition(field, file.Symbols)

		// Rename the field if its name conflits with any keyword
		for i := 0; symbol.IsReserved(field.Name) ||
			len(packageScope.ExcludeFile(class.Class.Name).FindStaticField().ByName(field.Name)) > 0; i++ {
			field.Rename(field.Name + strconv.Itoa(i))
		}
	}

	// Resolve all the methods
	for _, method := range class.Methods {
		// Resolve the return type, as well as the body of the method
		symbol.ResolveChildren(method, file.Symbols)

		// Comparison compares the method against the found method
		// This tests for a method of the same name, but with different
		// aspects of it, so that it can be identified as a duplicate
		comparison := func(d *symbol.Definition) bool {
			// The names must match, but everything else must be different
			if method.Name != d.Name {
				return false
			}

			// Size of parameters do not match
			if len(method.Parameters) != len(d.Parameters) {
				return true
			}

			// Go through the types and check to see if they differ
			for index, param := range method.Parameters {
				if param.OriginalType != d.Parameters[index].OriginalType {
					return true
				}
			}

			// Both methods are equal, skip this method since it is likely
			// the same method that we are trying to find duplicates of
			return false
		}

		for i := 0; symbol.IsReserved(method.Name) || len(class.FindMethod().By(comparison)) > 0; i++ {
			method.Rename(method.Name + strconv.Itoa(i))
		}
		// Resolve all the paramters of the method
		for _, param := range method.Parameters {
			symbol.ResolveDefinition(param, file.Symbols)

			for i := 0; symbol.IsReserved(param.Name); i++ {
				param.Rename(param.Name + strconv.Itoa(i))
			}
		}
	}
}
