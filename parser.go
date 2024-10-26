package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

// FileParser struct for parsing files
type FileParser struct {
	// You can define any fields if necessary
}

// NewFileParser initializes a new FileParser
func NewFileParser() *FileParser {
	return &FileParser{}
}

// ParseFile parses the given file for documentation blocks
func (fp *FileParser) ParseFile(filename string) {
	// Read the file content
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return
	}

	// Create a new file set and parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, data, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", filename, err)
		return
	}

	// Create a map to hold documentation blocks
	docMap := make(map[string]string)

	// Iterate through the declarations in the file
	for _, decl := range node.Decls {
		// Use type assertion to check if decl is of type *ast.GenDecl
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			// Check for comments associated with the declaration
			if genDecl.Doc != nil {
				position := fset.Position(genDecl.Pos())
				docMap[genDecl.Doc.Text()] = fmt.Sprintf("%s:%d", position.Filename, position.Line) // Store the documentation and position
			}
		}

		// Use type assertion to check if decl is of type *ast.FuncDecl
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			// Check for comments associated with the function declaration
			if funcDecl.Doc != nil {
				position := fset.Position(funcDecl.Pos())
				docMap[funcDecl.Doc.Text()] = fmt.Sprintf("%s:%d", position.Filename, position.Line) // Store the documentation and position
			}
		}
	}

	// Convert the map to JSON
	jsonOutput, err := json.MarshalIndent(docMap, "", "  ")
	if err != nil {
		fmt.Printf("Error converting documentation to JSON: %v\n", err)
		return
	}

	// Print the JSON output
	fmt.Printf("Parsed Documentation for %s:\n%s\n", filename, jsonOutput)
}
