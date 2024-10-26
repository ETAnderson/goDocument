package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

// FileParser struct
type FileParser struct {
	docMap map[string]FileData // Map to hold file names and their corresponding doc comments and function details
}

// FileData represents the structure of the data stored for each file
type FileData struct {
	Package   string           `json:"package"`
	Imports   []string         `json:"imports"`
	Functions []FunctionDetail `json:"functions"`
}

// FunctionDetail represents the structure of function details
type FunctionDetail struct {
	Name        string   `json:"name"`
	Docs        string   `json:"docs"` // Associate documentation with functions
	Params      []string `json:"params"`
	ParamTypes  []string `json:"param_types"`
	ReturnTypes []string `json:"return_types"`
}

// NewFileParser initializes a new FileParser
func NewFileParser() *FileParser {
	return &FileParser{docMap: make(map[string]FileData)}
}

// ParseFile parses the Go file and extracts documentation comments and function details
func (fp *FileParser) ParseFile(filePath string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file: %v", err)
		return
	}

	fileData := FileData{
		Package: node.Name.Name, // Get the declared package name
	}

	// Extract imported packages
	for _, imp := range node.Imports {
		if imp.Path != nil {
			fileData.Imports = append(fileData.Imports, strings.Trim(imp.Path.Value, `"`)) // Remove quotes safely
		}
	}

	// Iterate through declarations in the file
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Handle function declarations
			funcDetail := FunctionDetail{
				Name: d.Name.Name,
			}

			// Extract documentation comments for the function
			if d.Doc != nil {
				funcDetail.Docs = sanitizeDoc(d.Doc.Text())
			}

			// Extract parameters
			if d.Type != nil && d.Type.Params != nil {
				for _, param := range d.Type.Params.List {
					// Add parameter names
					for _, name := range param.Names {
						funcDetail.Params = append(funcDetail.Params, name.Name)
					}
					// Add parameter types
					funcDetail.ParamTypes = append(funcDetail.ParamTypes, formatType(param.Type))
				}
			}

			// Extract return types
			if d.Type != nil && d.Type.Results != nil {
				for _, result := range d.Type.Results.List {
					funcDetail.ReturnTypes = append(funcDetail.ReturnTypes, formatType(result.Type))
				}
			}

			fileData.Functions = append(fileData.Functions, funcDetail)
		}
	}

	// Store the parsed data for the file
	fp.docMap[filePath] = fileData

	// Write the JSON to the reference.json file
	if err := fp.writeJSONToFile("reference.json"); err != nil {
		log.Printf("Error writing JSON to file: %v", err)
	}
}

// writeJSONToFile writes the documentation map to a JSON file with pretty formatting
func (fp *FileParser) writeJSONToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err // Return the error to handle it in ParseFile
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Set indentation for better readability

	if err := encoder.Encode(fp.docMap); err != nil {
		return err // Return the error to handle it in ParseFile
	}
	return nil
}

// sanitizeDoc removes newlines and trims spaces from documentation strings
func sanitizeDoc(doc string) string {
	return strings.TrimSpace(strings.ReplaceAll(doc, "\n", " "))
}

// formatType converts an ast.Expr to a string representation
func formatType(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.X.(*ast.Ident).Name + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt) // handle array type
	case *ast.MapType:
		return "map[" + formatType(t.Key) + "]" + formatType(t.Value) // handle map type
	case *ast.StarExpr:
		return "*" + formatType(t.X) // handle pointer type
	case *ast.FuncType:
		return "func" + formatFuncType(t) // handle function type
	default:
		return "<unknown type>" // Provide a fallback for unknown types
	}
}

// formatFuncType formats the function type
func formatFuncType(funcType *ast.FuncType) string {
	paramTypes := ""
	if funcType.Params != nil {
		paramTypes += "("
		for i, param := range funcType.Params.List {
			paramTypes += formatType(param.Type)
			if i < len(funcType.Params.List)-1 {
				paramTypes += ", "
			}
		}
		paramTypes += ")"
	}
	return paramTypes
}
