package structkeys

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"testing"
)

const fileName = "test.go"

func TestValidatorDetection(t *testing.T) {

	t.Run("ignores non-struct composite literals", func(t *testing.T) {
		src := `
			package p
					
			var c = []string{"a", "b", "c"}
			var d = map[string]int{"a": 1, "b": 2}
		`
		failures := validate(t, src)
		if len(failures) != 0 {
			t.Errorf("got %d failures, want 0", len(failures))
		}
	})

	t.Run("empty struct initialization is allowed", func(t *testing.T) {
		src := `
			package p
			
			type A struct {
				Field1 string
				Field2 int
			}
			
			var a = A{}
		`
		failures := validate(t, src)
		if len(failures) != 0 {
			t.Errorf("got %d failures, want 0", len(failures))
		}
	})

	t.Run("struct initialization with keys is allowed", func(t *testing.T) {
		src := `
			package p
			
			type A struct {
				Field1 string
				Field2 int
			}
			
			var a = A{Field1: "hello", Field2: 1}
		`
		failures := validate(t, src)
		if len(failures) != 0 {
			t.Errorf("got %d failures, want 0", len(failures))
		}
	})

	t.Run("struct initialization without keys is not allowed", func(t *testing.T) {
		src := `
			package p
			
			type A struct {
				Field1 string
				Field2 int
			}
			
			var a = A{"hello", 1}
		`
		failures := validate(t, src)
		if len(failures) != 1 {
			t.Errorf("got %d failures, want 1", len(failures))
		}
	})

	t.Run("inline struct initialization without keys is not allowed", func(t *testing.T) {
		src := `
			package p
			
			var e = struct{ Field1 string }{"hello"}
		`
		failures := validate(t, src)
		if len(failures) != 1 {
			t.Errorf("got %d failures, want 1", len(failures))
		}
	})

	t.Run("inline struct initialization with keys is allowed", func(t *testing.T) {
		src := `
			package p
			
			var e = struct{ Field1 string }{Field1: "hello"}
		`
		failures := validate(t, src)
		if len(failures) != 0 {
			t.Errorf("got %d failures, want 0", len(failures))
		}
	})

}

func TestValidatorFailures(t *testing.T) {
	t.Run("reports failure position", func(t *testing.T) {
		src := `
			package p

			var a = struct{ Field1 string}{"hello"} // line starts with 4 tabs
		`

		failures := validate(t, src)
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}
		failure := failures[0]
		expectedMessage := "struct literals must use keys during initialization"
		if failure.Message != expectedMessage {
			t.Errorf("got %q, want %q", failure.Message, expectedMessage)
		}

		expectedLine := 4
		expectedPos := 12
		expectedInfo := fmt.Sprintf("%s:%d:%d: %s", fileName, expectedLine, expectedPos, expectedMessage)
		if failure.String() != expectedInfo {
			t.Errorf("got %q, want %q", failure.String(), expectedInfo)
		}

	})
}

func validate(t *testing.T, src string) []Failure {
	t.Helper()
	fs := token.NewFileSet()

	f, err := parser.ParseFile(fs, fileName, src, parser.SkipObjectResolution)
	if err != nil {
		log.Fatal(err)
	}
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}
	pkgPath := ""
	_, err = conf.Check(pkgPath, fs, []*ast.File{f}, info)

	failures := make([]Failure, 0)
	onFailure := func(failure Failure) {
		failures = append(failures, failure)
	}
	validator, err := NewValidator(fs, info, onFailure)
	if err != nil {
		log.Fatal(err)
	}

	ast.Walk(validator, f)

	return failures
}
