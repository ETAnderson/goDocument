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
	docMap map[string][]DocEntry // Store documentation entries for each file
}

// DocEntry holds a comment and its location
type DocEntry struct {
	Comment  string `json:"comment"`
	Location string `json:"location"`
}

// NewFileParser initializes a new FileParser
func NewFileParser() *FileParser {
	return &FileParser{
		docMap: make(map[string][]DocEntry),
	}
}

// ParseFile parses the given file for documentation blocks
func (fp *FileParser) ParseFile(filename string) {
	// Create a new file set and parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", filename, err)
		return
	}

	// Iterate through the declarations in the file
	for _, decl := range node.Decls {
		var entry DocEntry
		// Use type assertion to check if decl is of type *ast.GenDecl
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Doc != nil {
			position := fset.Position(genDecl.Pos())
			entry = DocEntry{
				Comment:  genDecl.Doc.Text(),
				Location: fmt.Sprintf("%s:%d", position.Filename, position.Line),
			}
			fp.docMap[filename] = append(fp.docMap[filename], entry)
		}

		// Use type assertion to check if decl is of type *ast.FuncDecl
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Doc != nil {
			position := fset.Position(funcDecl.Pos())
			entry = DocEntry{
				Comment:  funcDecl.Doc.Text(),
				Location: fmt.Sprintf("%s:%d", position.Filename, position.Line),
			}
			fp.docMap[filename] = append(fp.docMap[filename], entry)
		}
	}

	// Write the docMap to reference.json
	if err := fp.writeJSONToFile("reference.json"); err != nil {
		fmt.Printf("Error writing reference.json: %v\n", err)
	}
}

// writeJSONToFile writes the docMap to a JSON file
func (fp *FileParser) writeJSONToFile(filename string) error {
	file, err := os.Create(filename) // Overwrite the file if it exists
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert the docMap to JSON
	jsonOutput, err := json.MarshalIndent(fp.docMap, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON output to the file
	_, err = file.Write(jsonOutput)
	return err
}
