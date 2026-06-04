package schedulex_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestPublicAPISnapshot(t *testing.T) {
	set := token.NewFileSet()
	pkgs, err := parser.ParseDir(set, ".", func(info os.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range pkgs["schedulex"].Files {
		for _, d := range f.Decls {
			if gd, ok := d.(*ast.GenDecl); ok {
				for _, spec := range gd.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(s.Name.Name) {
							names = append(names, "type "+s.Name.Name)
						}
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if ast.IsExported(n.Name) {
								names = append(names, "value "+n.Name)
							}
						}
					}
				}
			}
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && ast.IsExported(fd.Name.Name) {
				names = append(names, "func "+fd.Name.Name)
			}
		}
	}
	sort.Strings(names)
	got := strings.Join(names, "\n") + "\n"
	want, err := os.ReadFile("../../contracts/public_api.snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Fatalf("public API snapshot mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
