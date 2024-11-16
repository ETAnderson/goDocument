package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	delete(fp.docMap, filePath) // Reset for fresh parse
	err := fp.tryParseFile(filePath, 3)
	if err != nil {
		log.Printf("Error parsing file: %v", err)
	}
}

// tryParseFile attempts to parse the file up to retryCount times to ensure stability
func (fp *FileParser) tryParseFile(filePath string, retryCount int) error {
	var lastErr error
	for i := 0; i < retryCount; i++ {
		fset := token.NewFileSet()
		if fp.fileExistsAndReadable(filePath) {
			node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
			if err == nil {
				fp.extractFileData(filePath, node)
				return nil
			}
			lastErr = err
		}
		time.Sleep(50 * time.Millisecond)
	}
	return lastErr
}

// fileExistsAndReadable checks if the file exists and is non-empty
func (fp *FileParser) fileExistsAndReadable(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir() && info.Size() > 0
}

// extractFileData processes the parsed node and stores it in docMap
func (fp *FileParser) extractFileData(filePath string, node *ast.File) {
	fileData := FileData{
		Package: node.Name.Name,
	}

	for _, imp := range node.Imports {
		if imp.Path != nil {
			fileData.Imports = append(fileData.Imports, strings.Trim(imp.Path.Value, `"`))
		}
	}

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			funcDetail := FunctionDetail{Name: d.Name.Name}
			if d.Doc != nil {
				funcDetail.Docs = sanitizeDoc(d.Doc.Text())
			}

			if d.Type != nil && d.Type.Params != nil {
				for _, param := range d.Type.Params.List {
					for _, name := range param.Names {
						funcDetail.Params = append(funcDetail.Params, name.Name)
					}
					funcDetail.ParamTypes = append(funcDetail.ParamTypes, formatType(param.Type))
				}
			}

			if d.Type != nil && d.Type.Results != nil {
				for _, result := range d.Type.Results.List {
					funcDetail.ReturnTypes = append(funcDetail.ReturnTypes, formatType(result.Type))
				}
			}
			fileData.Functions = append(fileData.Functions, funcDetail)
		}
	}

	fp.docMap[filePath] = fileData

	if err := fp.writeJSONToFile(filePath); err != nil {
		log.Printf("Error writing JSON to file: %v", err)
	}
}

// writeJSONToFile writes the documentation map to a JSON file in the references directory
func (fp *FileParser) writeJSONToFile(jsonFilePath string) error {
	// Create the references directory if it doesn't exist
	dir := filepath.Dir(jsonFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Create the JSON file
	file, err := os.Create(jsonFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(fp.docMap[jsonFilePath])
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
		return "[]" + formatType(t.Elt)
	case *ast.MapType:
		return "map[" + formatType(t.Key) + "]" + formatType(t.Value)
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.FuncType:
		return "func" + formatFuncType(t)
	case *ast.ChanType:
		return "chan " + formatType(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		if strType, ok := expr.(*ast.Ident); ok && strType.Name == "interface" {
			return "interface{}"
		}
		if strType, ok := expr.(*ast.MapType); ok {
			return "map[" + formatType(strType.Key) + "]" + formatType(strType.Value)
		}
		if strType, ok := expr.(*ast.ArrayType); ok {
			return "[]" + formatType(strType.Elt)
		}
		return "<unknown type>"
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
