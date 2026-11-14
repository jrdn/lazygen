package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLazyGenPrintDoesNotWriteOutputFile(t *testing.T) {
	src := writeMainSource(t, `package example

/* lazygen:template Hello
func Hello() {}
*/
// lazygen:instance Hello
`)
	out := strings.TrimSuffix(src, ".go") + "_gen.go"
	output, err := captureStdout(t, func() error {
		return lazyGen(LazyGenOptions{SrcFile: src, Print: true})
	})
	if err != nil {
		t.Fatalf("lazyGen print mode: %v", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("print mode should not create output file %q (stat err %v)", out, err)
	}
	if !strings.Contains(output, "func Hello()") {
		t.Fatalf("printed output missing generated function: %q", output)
	}
}

func TestLazyGenWritesGeneratedFile(t *testing.T) {
	src := writeMainSource(t, `package example

/* lazygen:template Hello
func Hello() {}
*/
// lazygen:instance Hello
`)
	if err := lazyGen(LazyGenOptions{SrcFile: src}); err != nil {
		t.Fatalf("lazyGen: %v", err)
	}
	if _, err := os.Stat(strings.TrimSuffix(src, ".go") + "_gen.go"); err != nil {
		t.Fatalf("generated file not written: %v", err)
	}
}

func TestMainReturnsNonzeroOnGenerationFailure(t *testing.T) {
	if os.Getenv("LAZYGEN_FAILURE_HELPER") == "1" {
		os.Args = []string{"lazygen", "-printOutput", filepath.Join(t.TempDir(), "missing.go")}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainReturnsNonzeroOnGenerationFailure")
	cmd.Env = append(os.Environ(), "LAZYGEN_FAILURE_HELPER=1")
	if err := cmd.Run(); err == nil {
		t.Fatal("CLI exited 0 for a missing source file")
	}
}

func writeMainSource(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.go")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	result := make(chan []byte, 1)
	readDone := make(chan error, 1)
	go func() {
		data, readErr := io.ReadAll(r)
		result <- data
		readDone <- readErr
	}()
	errFn := fn()
	if closeErr := w.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	os.Stdout = old
	out := <-result
	if readErr := <-readDone; readErr != nil {
		t.Fatal(readErr)
	}
	_ = r.Close()
	return string(out), errFn
}
