package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTemplateAndInstances(t *testing.T) {
	path := writeSource(t, `package sample

/* lazygen:template Greeting
func {{.name}}() { println("{{.message}}") }
*/
// lazygen:instance Greeting name=Hello message=world
`)
	g, err := NewParser(path)
	if err != nil {
		t.Fatal(err)
	}
	if g.Package != "sample" {
		t.Fatalf("package = %q, want sample", g.Package)
	}
	got, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "func Hello()") || !strings.Contains(got, `println("world")`) {
		t.Fatalf("generated output missing rendered instance: %s", got)
	}
}

func TestParseImports(t *testing.T) {
	path := writeSource(t, "package sample\n\n/* lazygen:imports\n\"fmt\"\n*/\n")
	g, err := NewParser(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "import \"fmt\"") {
		t.Fatalf("generated output missing declared import: %s", got)
	}
}

func TestParseMalformedInstanceParameter(t *testing.T) {
	for _, directive := range []string{
		"// lazygen:instance Greeting name",
		"// lazygen:instance Greeting =value",
	} {
		t.Run(directive, func(t *testing.T) {
			path := writeSource(t, "package sample\n\n"+directive+"\n")
			if _, err := NewParser(path); err == nil {
				t.Fatal("NewParser succeeded for malformed instance parameter")
			}
		})
	}
}

func TestGenerateAddsStandardLibraryImportsUsedByTemplate(t *testing.T) {
	g := &Generator{
		Package:    "sample",
		SourcePath: "example.go",
		Templates:  map[string]string{"Greeting": `func {{.name}}() { fmt.Println("{{.message}}") }`},
		Instances:  map[string][]map[string]interface{}{"Greeting": {map[string]interface{}{"name": "Hello", "message": "world"}}},
	}
	got, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `import "fmt"`) {
		t.Fatalf("generated output missing fmt import: %s", got)
	}
}

func TestGenerateMissingTemplate(t *testing.T) {
	path := writeSource(t, "package sample\n\n// lazygen:instance Missing\n")
	g, err := NewParser(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Generate(); err == nil || !strings.Contains(err.Error(), "template Missing") {
		t.Fatalf("Generate error = %v, want missing-template error", err)
	}
}

func TestGenerateMalformedTemplate(t *testing.T) {
	path := writeSource(t, "package sample\n\n/* lazygen:template Broken\n{{if}\n*/\n// lazygen:instance Broken\n")
	g, err := NewParser(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Generate(); err == nil {
		t.Fatal("Generate succeeded with malformed template")
	}
}

func writeSource(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.go")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
