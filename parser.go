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
	docMap map[string][]FunctionInfo // Holds documentation and function info
}

// FunctionInfo holds information about functions
type FunctionInfo struct {
	Name       string   `json:"name"`
	Params     []string `json:"params"`
	ParamTypes []string `json:"param_types"`
	ReturnType string   `json:"return_type"`
	Doc        string   `json:"doc"`
}

// NewFileParser initializes a new FileParser
func NewFileParser() *FileParser {
	return &FileParser{
		docMap: make(map[string][]FunctionInfo),
	}
}

// ParseFile parses the given file for documentation blocks and function signatures
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

	// Iterate through the declarations in the file
	for _, decl := range node.Decls {
		// Check for general declarations
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Doc != nil {
			// Currently not handling TypeSpec; you can add your logic here if needed.
		}

		// Check for function declarations
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			functionInfo := FunctionInfo{
				Name: funcDecl.Name.Name,
				Doc:  funcDecl.Doc.Text(),
			}

			// Extract parameters
			if funcDecl.Type.Params != nil {
				for _, param := range funcDecl.Type.Params.List {
					for _, name := range param.Names {
						functionInfo.Params = append(functionInfo.Params, name.Name)
						functionInfo.ParamTypes = append(functionInfo.ParamTypes, fmt.Sprint(param.Type))
					}
				}
			}

			// Extract return types
			if funcDecl.Type.Results != nil {
				for _, result := range funcDecl.Type.Results.List {
					functionInfo.ReturnType = fmt.Sprint(result.Type)
				}
			}

			fp.docMap[filename] = append(fp.docMap[filename], functionInfo)
		}
	}

	// Write the parsed information to JSON file
	fp.writeJSONToFile("reference.json")
}

// writeJSONToFile writes the documentation to a JSON file
func (fp *FileParser) writeJSONToFile(filename string) {
	jsonOutput, err := json.MarshalIndent(fp.docMap, "", "  ")
	if err != nil {
		fmt.Printf("Error converting documentation to JSON: %v\n", err)
		return
	}

	// Write to the file
	if err := os.WriteFile(filename, jsonOutput, 0644); err != nil {
		fmt.Printf("Error writing to JSON file: %v\n", err)
	} else {
		fmt.Println("Parsed documentation written to", filename)
	}
}
