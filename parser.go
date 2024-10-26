package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// FileParser struct for parsing files
type FileParser struct {
	docMap map[string]string
}

// NewFileParser initializes a new FileParser
func NewFileParser() *FileParser {
	return &FileParser{
		docMap: make(map[string]string),
	}
}

// ParseFile parses the given file for documentation blocks
func (fp *FileParser) ParseFile(filename string) {
	// Read the file content
	data, err := os.ReadFile(filename)
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

	// Clear the previous docMap for each new file
	fp.docMap = make(map[string]string)

	// Iterate through the declarations in the file
	for _, decl := range node.Decls {
		// Check for comments associated with the declaration
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Doc != nil {
				position := fset.Position(genDecl.Pos())
				fp.docMap[genDecl.Doc.Text()] = fmt.Sprintf("%s:%d", position.Filename, position.Line) // Store the documentation and position
			}
		}

		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Doc != nil {
				position := fset.Position(funcDecl.Pos())
				fp.docMap[funcDecl.Doc.Text()] = fmt.Sprintf("%s:%d", position.Filename, position.Line) // Store the documentation and position
			}
		}
	}

	// Write to JSON file
	fp.writeJSONToFile("reference.json")
}

// writeJSONToFile writes the docMap to a JSON file
func (fp *FileParser) writeJSONToFile(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	// Marshal docMap to JSON
	jsonOutput, err := json.MarshalIndent(fp.docMap, "", "  ")
	if err != nil {
		fmt.Printf("Error converting documentation to JSON: %v\n", err)
		return
	}

	// Write JSON to file
	if _, err := file.Write(jsonOutput); err != nil {
		fmt.Printf("Error writing to JSON file %s: %v\n", filename, err)
		return
	}

	fmt.Printf("Parsed documentation written to %s\n", filename)
}
