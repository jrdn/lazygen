package pkg

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

const (
	LazyGenTemplate = "lazygen:template"
	LazyGenInstance = "lazygen:instance"
	LazyGenImports  = "lazygen:imports"
)

type Generator struct {
	SourcePath string
	Package    string
	Imports    []string
	Templates  map[string]string
	Instances  map[string][]map[string]interface{}
}

func NewParser(path string) (*Generator, error) {
	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	generator := &Generator{
		SourcePath: path,
		Templates:  make(map[string]string),
		Instances:  make(map[string][]map[string]interface{}),
	}

	generator.Package = astFile.Name.Name

	for _, commentGroup := range astFile.Comments {
		for _, comment := range commentGroup.List {
			cleaned := generator.CleanComment(comment.Text)
			err := generator.ParseComment(cleaned)
			if err != nil {
				return nil, err
			}
		}
	}

	return generator, nil
}

func (g *Generator) Outfile() string {
	return strings.TrimSuffix(g.SourcePath, ".go") + "_gen.go"
}

func (g *Generator) ParseImports(comment string) error {
	lines := strings.Split(comment, "\n")
	if len(lines) <= 1 {
		return fmt.Errorf("got %s header but not body", LazyGenImports)
	}

	g.Imports = append(g.Imports, lines[1:]...)

	return nil
}

func (g *Generator) ParseTemplate(comment string) error {
	lines := strings.Split(comment, "\n")
	if len(lines) <= 1 {
		return fmt.Errorf("got %s header but not body", LazyGenTemplate)
	}

	header := lines[0]
	parts := strings.Split(strings.TrimSpace(header), " ")
	if len(parts) != 2 {
		return fmt.Errorf("%s has no name", LazyGenTemplate)
	}

	name := parts[1]

	templateContent := strings.Builder{}
	for _, l := range lines[1:] {
		templateContent.WriteString(l)
		templateContent.WriteString("\n")
	}

	g.Templates[name] = templateContent.String()
	return nil
}

func (g *Generator) ParseInstance(comment string) error {
	var parts []string
	for _, line := range strings.Split(comment, "\n") {
		parts = append(parts, strings.Split(line, " ")...)
	}

	if parts == nil {
		return fmt.Errorf("failed to parse lazygen:instance comment")
	}

	if len(parts) < 2 {
		return fmt.Errorf("failed to parse lazygen:instance comment, missing name")
	}

	name := parts[1]

	params := make(map[string]interface{})
	for _, part := range parts[2:] {
		paramParts := strings.Split(part, "=")

		params[strings.TrimSpace(paramParts[0])] = strings.TrimSpace(paramParts[1])
	}
	g.Instances[name] = append(g.Instances[name], params)

	return nil
}

func (g *Generator) ParseComment(comment string) error {
	switch {
	case strings.HasPrefix(comment, LazyGenTemplate):
		return g.ParseTemplate(comment)
	case strings.HasPrefix(comment, LazyGenInstance):
		return g.ParseInstance(comment)
	case strings.HasPrefix(comment, LazyGenImports):
		return g.ParseImports(comment)
	}
	return nil
}

func (g *Generator) CleanComment(comment string) string {
	comment = strings.TrimPrefix(comment, "/*")
	comment = strings.TrimSuffix(comment, "*/")
	comment = strings.TrimSpace(comment)

	buf := strings.Builder{}

	for _, line := range strings.Split(comment, "\n") {
		line = strings.TrimPrefix(line, "//")
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	return strings.TrimSpace(buf.String())
}
