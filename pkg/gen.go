package pkg

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"sort"
	"strings"
	"text/template"
)

func (g *Generator) Generate() (string, error) {
	generated := strings.Builder{}
	generated.WriteString("package ")
	generated.WriteString(g.Package)
	generated.WriteString(fmt.Sprintf("\n\n// Code generated from %s by lazygen DO NOT EDIT.\n\n", g.SourcePath))

	if len(g.Imports) > 0 {
		generated.WriteString("import (\n")
		for _, imp := range g.Imports {
			generated.WriteString("\t")
			generated.WriteString(imp)
			generated.WriteString("\n")
		}
		generated.WriteString(")\n\n")
	}

	templateNames := make([]string, len(g.Instances))
	x := 0
	for templateName := range g.Instances {
		templateNames[x] = templateName
		x++
	}
	sort.Strings(templateNames)

	for _, templateName := range templateNames {
		templateData, ok := g.Templates[templateName]
		if !ok {
			return "", fmt.Errorf("found instance of template %s but template not defined", templateName)
		}

		t, err := template.New(templateName).Funcs(sprig.FuncMap()).Parse(templateData)

		for _, values := range g.Instances[templateName] {
			if err != nil {
				return "", err
			}

			buf := &bytes.Buffer{}
			err := t.Execute(buf, values)
			if err != nil {
				return "", err
			}
			generated.WriteString(buf.String())
			generated.WriteString("\n")
		}
	}
	return generated.String(), nil
}
