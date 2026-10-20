package organize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

var (
	// Regex patterns for lightweight non-Go AST heuristics
	pyMainRegex     = regexp.MustCompile(`if\s+__name__\s*==\s*['"]__main__['"]\s*:`)
	pyAPIRegex      = regexp.MustCompile(`@(?:app|router|bp)\.(?:get|post|put|delete|patch|route)`)
	pyModelRegex    = regexp.MustCompile(`class\s+\w+\s*\([^)]*(?:BaseModel|db\.Model|Entity|dataclass)[^)]*\):`)
	pyImportRegex   = regexp.MustCompile(`(?:from\s+([\w\.]+)\s+import|import\s+([\w\.]+))`)

	jsMainRegex     = regexp.MustCompile(`^#!\/usr\/bin\/env\s+node|program\.parse|cli\.run`)
	jsAPIRegex      = regexp.MustCompile(`(?:app|router)\.(?:get|post|put|delete|use)\s*\(`)
	jsModelRegex    = regexp.MustCompile(`(?:export\s+)?(?:interface|type)\s+\w+\s*(?:\{|=|extends)|z\.object\(`)
	jsImportRegex   = regexp.MustCompile(`(?:import\s+.*?from\s+['"]([^'"]+)['"]|require\(['"]([^'"]+)['"]\))`)
)

// AnalyzeCodeNode performs deep semantic inspection on a source code file.
func AnalyzeCodeNode(node *FileNode) error {
	if node.Category != CategoryCode {
		return nil
	}

	content, err := os.ReadFile(node.AbsPath)
	if err != nil {
		return err
	}
	src := string(content)

	switch node.Language {
	case "go":
		return analyzeGoSource(node, src)
	case "python":
		return analyzePythonSource(node, src)
	case "javascript", "typescript":
		return analyzeJavaScriptSource(node, src)
	default:
		node.Role = RoleUnknown
		return nil
	}
}

// analyzeGoSource inspects Go source files using go/parser and go/ast.
func analyzeGoSource(node *FileNode, src string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, node.AbsPath, src, parser.ParseComments)
	if err != nil {
		// Fallback to lexical heuristics if syntax error occurs
		node.Role = RoleUnknown
		return nil
	}

	node.PackageName = f.Name.Name

	// Extract imported paths
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		node.Imports = append(node.Imports, path)
	}

	// Extract exported symbols and analyze declarations
	hasMainFunc := false
	hasHTTPHandler := false
	structCount := 0
	pureFunctionCount := 0

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			name := d.Name.Name
			if ast.IsExported(name) {
				node.ExportedSyms = append(node.ExportedSyms, name)
			}
			if name == "main" && node.PackageName == "main" {
				hasMainFunc = true
			}
			// Check for HTTP Handler signatures
			if d.Type.Params != nil {
				for _, param := range d.Type.Params.List {
					paramType := typesToString(param.Type)
					if strings.Contains(paramType, "ResponseWriter") || strings.Contains(paramType, "Request") {
						hasHTTPHandler = true
					}
				}
			}
			if d.Recv == nil {
				pureFunctionCount++
			}

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if ast.IsExported(typeSpec.Name.Name) {
						node.ExportedSyms = append(node.ExportedSyms, typeSpec.Name.Name)
					}
					if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
						structCount++
					}
					if _, isInterface := typeSpec.Type.(*ast.InterfaceType); isInterface {
						structCount++
					}
				}
			}
		}
	}

	// Determine Role
	if hasMainFunc {
		node.Role = RoleEntry
		return nil
	}
	if hasHTTPHandler || hasImportPrefix(node.Imports, "net/http") {
		node.Role = RoleAPI
		return nil
	}
	if structCount > 0 && len(node.Imports) <= 2 {
		node.Role = RoleModel
		return nil
	}
	if len(node.Imports) == 0 && pureFunctionCount > 0 {
		node.Role = RoleUtil
		return nil
	}
	if len(node.Imports) > 2 {
		node.Role = RoleService
		return nil
	}

	node.Role = RoleUnknown
	return nil
}

// typesToString converts AST expression types to string for signature checks.
func typesToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typesToString(t.X)
	case *ast.SelectorExpr:
		return typesToString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

func hasImportPrefix(imports []string, prefix string) bool {
	for _, imp := range imports {
		if strings.HasPrefix(imp, prefix) {
			return true
		}
	}
	return false
}

// analyzePythonSource inspects Python source files using regex heuristics.
func analyzePythonSource(node *FileNode, src string) error {
	matches := pyImportRegex.FindAllStringSubmatch(src, -1)
	for _, m := range matches {
		if m[1] != "" {
			node.Imports = append(node.Imports, m[1])
		} else if m[2] != "" {
			node.Imports = append(node.Imports, m[2])
		}
	}

	if pyMainRegex.MatchString(src) {
		node.Role = RoleEntry
		return nil
	}
	if pyAPIRegex.MatchString(src) || strings.Contains(src, "FastAPI") || strings.Contains(src, "Flask") {
		node.Role = RoleAPI
		return nil
	}
	if pyModelRegex.MatchString(src) || strings.Contains(src, "@dataclass") {
		node.Role = RoleModel
		return nil
	}
	if len(node.Imports) == 0 {
		node.Role = RoleUtil
		return nil
	}

	node.Role = RoleService
	return nil
}

// analyzeJavaScriptSource inspects JavaScript and TypeScript source files.
func analyzeJavaScriptSource(node *FileNode, src string) error {
	matches := jsImportRegex.FindAllStringSubmatch(src, -1)
	for _, m := range matches {
		if m[1] != "" {
			node.Imports = append(node.Imports, m[1])
		} else if m[2] != "" {
			node.Imports = append(node.Imports, m[2])
		}
	}

	if jsMainRegex.MatchString(src) {
		node.Role = RoleEntry
		return nil
	}
	if jsAPIRegex.MatchString(src) || strings.Contains(src, "express()") {
		node.Role = RoleAPI
		return nil
	}
	if jsModelRegex.MatchString(src) {
		node.Role = RoleModel
		return nil
	}
	if len(node.Imports) == 0 {
		node.Role = RoleUtil
		return nil
	}

	node.Role = RoleService
	return nil
}
