package pkg

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

func (g *Generator) Generate() (string, error) {
	var body strings.Builder
	templateNames := make([]string, 0, len(g.Instances))
	for name := range g.Instances {
		templateNames = append(templateNames, name)
	}
	sort.Strings(templateNames)
	for _, name := range templateNames {
		templateBody, ok := g.Templates[name]
		if !ok {
			return "", fmt.Errorf("found instance of template %s but template not defined", name)
		}
		t, err := template.New(name).Funcs(sprig.FuncMap()).Parse(templateBody)
		if err != nil {
			return "", err
		}
		for _, values := range g.Instances[name] {
			buf := &bytes.Buffer{}
			if err := t.Execute(buf, values); err != nil {
				return "", err
			}
			body.WriteString(buf.String())
			body.WriteString("\n")
		}
	}
	imports := append([]string(nil), g.Imports...)
	if len(imports) == 0 {
		var err error
		imports, err = inferStandardLibraryImports(body.String())
		if err != nil {
			return "", fmt.Errorf("inspect generated templates: %w", err)
		}
	}

	var generated strings.Builder
	generated.WriteString("package ")
	generated.WriteString(g.Package)
	generated.WriteString(fmt.Sprintf("\n\n// Code generated from %s by lazygen DO NOT EDIT.\n\n", g.SourcePath))
	if len(imports) == 1 {
		generated.WriteString("import " + imports[0] + "\n\n")
	} else if len(imports) > 1 {
		generated.WriteString("import (\n")
		for _, imp := range imports {
			generated.WriteString("\t" + imp + "\n")
		}
		generated.WriteString(")\n\n")
	}
	generated.WriteString(body.String())
	formatted, err := format.Source([]byte(generated.String()))
	if err != nil {
		return "", fmt.Errorf("format generated Go: %w", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", formatted, parser.AllErrors); err != nil {
		return "", fmt.Errorf("generated output is not valid Go: %w", err)
	}
	return string(formatted), nil
}

func inferStandardLibraryImports(source string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "template.go", "package generated\n"+source, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	used := make(map[string]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		sel, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if ok && standardPackages[pkg.Name] {
			used[pkg.Name] = true
		}
		return true
	})
	imports := make([]string, 0, len(used))
	for name := range used {
		imports = append(imports, fmt.Sprintf("%q", name))
	}
	sort.Strings(imports)
	return imports, nil
}

var standardPackages = map[string]bool{
	"fmt": true, "strings": true, "strconv": true, "bytes": true,
	"io": true, "os": true, "time": true, "math": true,
	"sort": true, "errors": true, "context": true, "sync": true,
}
