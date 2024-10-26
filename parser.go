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

type FileParser struct {
	docMap map[string]FileData
}

type FileData struct {
	Package   string           `json:"package"`
	Imports   []string         `json:"imports"`
	Functions []FunctionDetail `json:"functions"`
	Variables []VariableDetail `json:"variables"`
}

type FunctionDetail struct {
	Name        string   `json:"name"`
	Docs        []string `json:"docs"`
	Params      []string `json:"params"`
	ParamTypes  []string `json:"param_types"`
	ReturnTypes []string `json:"return_types"`
}

type VariableDetail struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Docs string `json:"docs"`
}

func NewFileParser() *FileParser {
	return &FileParser{docMap: make(map[string]FileData)}
}

func (fp *FileParser) ParseFile(filePath string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file '%s': %v", filePath, err)
		return
	}

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
		case *ast.GenDecl:
			if d.Tok == token.VAR || d.Tok == token.CONST {
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok && len(vs.Names) > 0 {
						varDetail := VariableDetail{
							Name: vs.Names[0].Name,
							Type: formatType(vs.Type),
						}
						if vs.Doc != nil {
							varDetail.Docs = sanitizeDoc(vs.Doc.Text())
						} else if vs.Comment != nil {
							varDetail.Docs = sanitizeDoc(vs.Comment.Text())
						}
						fileData.Variables = append(fileData.Variables, varDetail)
					}
				}
			}
		case *ast.FuncDecl:
			funcDetail := FunctionDetail{
				Name: d.Name.Name,
			}

			if d.Doc != nil {
				funcDetail.Docs = append(funcDetail.Docs, sanitizeDoc(d.Doc.Text()))
			}

			if d.Type.Params != nil {
				for _, param := range d.Type.Params.List {
					for _, name := range param.Names {
						funcDetail.Params = append(funcDetail.Params, name.Name)
					}
					funcDetail.ParamTypes = append(funcDetail.ParamTypes, formatType(param.Type))
				}
			}

			if d.Type.Results != nil {
				for _, result := range d.Type.Results.List {
					funcDetail.ReturnTypes = append(funcDetail.ReturnTypes, formatType(result.Type))
				}
			}

			fileData.Functions = append(fileData.Functions, funcDetail)
		}
	}

	fp.docMap[filePath] = fileData

	if err := fp.writeJSONToFile("reference.json"); err != nil {
		log.Printf("Error writing JSON to file: %v", err)
	}
}

func (fp *FileParser) writeJSONToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(fp.docMap); err != nil {
		return err
	}
	return nil
}

func sanitizeDoc(doc string) string {
	return strings.TrimSpace(strings.ReplaceAll(doc, "\n", " "))
}

func formatType(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	case *ast.MapType:
		return "map[" + formatType(t.Key) + "]" + formatType(t.Value)
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.FuncType:
		return "func" + formatFuncType(t)
	default:
		return "<unknown type>"
	}
}

func formatFuncType(funcType *ast.FuncType) string {
	if funcType == nil {
		return ""
	}

	paramTypes := "("
	for i, param := range funcType.Params.List {
		paramTypes += formatType(param.Type)
		if i < len(funcType.Params.List)-1 {
			paramTypes += ", "
		}
	}
	paramTypes += ")"
	return paramTypes
}
