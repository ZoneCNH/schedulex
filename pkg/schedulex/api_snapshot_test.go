package schedulex_test

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestPublicAPISnapshot(t *testing.T) {
	got := publicAPISnapshot(t)
	want, err := os.ReadFile("../../contracts/public_api.snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Fatalf("public API snapshot mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func publicAPISnapshot(t *testing.T) string {
	t.Helper()

	set := token.NewFileSet()
	files := parsePackageFiles(t, set, ".", "schedulex")
	formatter := newAPIFormatter(set, files)

	var entries []apiEntry
	for _, f := range files {
		for _, d := range f.Decls {
			if gd, ok := d.(*ast.GenDecl); ok {
				for _, spec := range gd.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(s.Name.Name) {
							entries = append(entries, typeEntry(formatter, s))
						}
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if ast.IsExported(n.Name) {
								entries = append(entries, valueEntry(formatter, gd.Tok, s, n))
							}
						}
					}
				}
			}
			if fd, ok := d.(*ast.FuncDecl); ok {
				if fd.Recv == nil && ast.IsExported(fd.Name.Name) {
					entries = append(entries, funcEntry(formatter, fd))
				}
				if fd.Recv != nil && ast.IsExported(fd.Name.Name) {
					if recv, ok := exportedReceiver(formatter, fd.Recv); ok {
						entries = append(entries, methodEntry(formatter, recv, fd))
					}
				}
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })

	var lines []string
	for _, entry := range entries {
		lines = append(lines, entry.lines...)
	}
	return strings.Join(lines, "\n") + "\n"
}

func parsePackageFiles(t *testing.T, set *token.FileSet, dir, packageName string) map[string]*ast.File {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	files := make(map[string]*ast.File)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		if file.Name.Name == packageName {
			files[path] = file
		}
	}
	if len(files) == 0 {
		t.Fatalf("%s package not found", packageName)
	}
	return files
}

type apiEntry struct {
	key   string
	lines []string
}

type apiFormatter struct {
	set          *token.FileSet
	packageTypes map[string]struct{}
}

func newAPIFormatter(set *token.FileSet, files map[string]*ast.File) *apiFormatter {
	formatter := &apiFormatter{
		set:          set,
		packageTypes: make(map[string]struct{}),
	}
	for _, file := range files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if ok {
					formatter.packageTypes[typeSpec.Name.Name] = struct{}{}
				}
			}
		}
	}
	return formatter
}

func (f *apiFormatter) typeText(expr ast.Expr) string {
	switch typ := expr.(type) {
	case *ast.Ident:
		if _, ok := f.packageTypes[typ.Name]; ok && !ast.IsExported(typ.Name) {
			return "<unexported>"
		}
		return typ.Name
	case *ast.StarExpr:
		return "*" + f.typeText(typ.X)
	case *ast.Ellipsis:
		return "..." + f.typeText(typ.Elt)
	case *ast.ArrayType:
		if typ.Len == nil {
			return "[]" + f.typeText(typ.Elt)
		}
		return "[" + nodeString(f.set, typ.Len) + "]" + f.typeText(typ.Elt)
	case *ast.MapType:
		return "map[" + f.typeText(typ.Key) + "]" + f.typeText(typ.Value)
	case *ast.ChanType:
		switch typ.Dir {
		case ast.RECV:
			return "<-" + "chan " + f.typeText(typ.Value)
		case ast.SEND:
			return "chan<- " + f.typeText(typ.Value)
		default:
			return "chan " + f.typeText(typ.Value)
		}
	case *ast.SelectorExpr:
		return f.typeText(typ.X) + "." + typ.Sel.Name
	case *ast.ParenExpr:
		return "(" + f.typeText(typ.X) + ")"
	case *ast.IndexExpr:
		return f.typeText(typ.X) + "[" + f.typeText(typ.Index) + "]"
	case *ast.IndexListExpr:
		indices := make([]string, 0, len(typ.Indices))
		for _, index := range typ.Indices {
			indices = append(indices, f.typeText(index))
		}
		return f.typeText(typ.X) + "[" + strings.Join(indices, ", ") + "]"
	case *ast.FuncType:
		return f.funcType(typ)
	default:
		return nodeString(f.set, typ)
	}
}

func (f *apiFormatter) funcType(fn *ast.FuncType) string {
	signature := "func(" + strings.Join(f.fieldTypes(fn.Params), ", ") + ")"
	results := f.fieldTypes(fn.Results)
	switch len(results) {
	case 0:
		return signature
	case 1:
		return signature + " " + results[0]
	default:
		return signature + " (" + strings.Join(results, ", ") + ")"
	}
}

func (f *apiFormatter) fieldTypes(list *ast.FieldList) []string {
	if list == nil {
		return nil
	}
	var types []string
	for _, field := range list.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		typeText := f.typeText(field.Type)
		for i := 0; i < count; i++ {
			types = append(types, typeText)
		}
	}
	return types
}

func typeEntry(formatter *apiFormatter, spec *ast.TypeSpec) apiEntry {
	name := spec.Name.Name
	prefix := "type " + name
	if spec.Assign.IsValid() {
		prefix += " ="
	}

	switch typ := spec.Type.(type) {
	case *ast.StructType:
		lines := []string{prefix + " struct {"}
		for _, field := range typ.Fields.List {
			lines = append(lines, exportedFieldLines(formatter, field)...)
		}
		if len(lines) == 1 {
			lines[0] = prefix + " struct {}"
		} else {
			lines = append(lines, "}")
		}
		return apiEntry{key: "type " + name, lines: lines}
	case *ast.InterfaceType:
		lines := []string{prefix + " interface {"}
		for _, field := range typ.Methods.List {
			lines = append(lines, interfaceMethodLines(formatter, field)...)
		}
		if len(lines) == 1 {
			lines[0] = prefix + " interface {}"
		} else {
			lines = append(lines, "}")
		}
		return apiEntry{key: "type " + name, lines: lines}
	default:
		return apiEntry{key: "type " + name, lines: []string{prefix + " " + formatter.typeText(typ)}}
	}
}

func valueEntry(formatter *apiFormatter, tok token.Token, spec *ast.ValueSpec, name *ast.Ident) apiEntry {
	line := strings.ToLower(tok.String()) + " " + name.Name
	if spec.Type != nil {
		line += " " + formatter.typeText(spec.Type)
	}
	if value := valueForName(spec, name); value != nil {
		line += " = " + nodeString(formatter.set, value)
	}
	return apiEntry{key: line, lines: []string{line}}
}

func valueForName(spec *ast.ValueSpec, name *ast.Ident) ast.Expr {
	if len(spec.Values) == 0 {
		return nil
	}
	for i, candidate := range spec.Names {
		if candidate == name && i < len(spec.Values) {
			return spec.Values[i]
		}
	}
	return nil
}

func funcEntry(formatter *apiFormatter, fd *ast.FuncDecl) apiEntry {
	signature := strings.TrimPrefix(formatter.funcType(fd.Type), "func")
	line := "func " + fd.Name.Name + signature
	return apiEntry{key: line, lines: []string{line}}
}

func methodEntry(formatter *apiFormatter, receiver string, fd *ast.FuncDecl) apiEntry {
	signature := strings.TrimPrefix(formatter.funcType(fd.Type), "func")
	line := "method (" + receiver + ")." + fd.Name.Name + signature
	return apiEntry{key: line, lines: []string{line}}
}

func exportedReceiver(formatter *apiFormatter, recv *ast.FieldList) (string, bool) {
	if recv == nil || len(recv.List) == 0 {
		return "", false
	}
	base := receiverBaseName(recv.List[0].Type)
	if !ast.IsExported(base) {
		return "", false
	}
	return formatter.typeText(recv.List[0].Type), true
}

func receiverBaseName(expr ast.Expr) string {
	switch typ := expr.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.ParenExpr:
		return receiverBaseName(typ.X)
	case *ast.StarExpr:
		return receiverBaseName(typ.X)
	case *ast.IndexExpr:
		return receiverBaseName(typ.X)
	case *ast.IndexListExpr:
		return receiverBaseName(typ.X)
	default:
		return ""
	}
}

func exportedFieldLines(formatter *apiFormatter, field *ast.Field) []string {
	typeText := formatter.typeText(field.Type)
	tag := ""
	if field.Tag != nil {
		tag = " " + field.Tag.Value
	}
	if len(field.Names) == 0 {
		if !ast.IsExported(embeddedFieldName(field.Type)) {
			return nil
		}
		return []string{"\t" + typeText + tag}
	}

	var lines []string
	for _, name := range field.Names {
		if ast.IsExported(name.Name) {
			lines = append(lines, "\t"+name.Name+" "+typeText+tag)
		}
	}
	return lines
}

func embeddedFieldName(expr ast.Expr) string {
	switch typ := expr.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.ParenExpr:
		return embeddedFieldName(typ.X)
	case *ast.StarExpr:
		return embeddedFieldName(typ.X)
	case *ast.SelectorExpr:
		return typ.Sel.Name
	case *ast.IndexExpr:
		return embeddedFieldName(typ.X)
	case *ast.IndexListExpr:
		return embeddedFieldName(typ.X)
	default:
		return ""
	}
}

func interfaceMethodLines(formatter *apiFormatter, field *ast.Field) []string {
	if len(field.Names) == 0 {
		return []string{"\t" + formatter.typeText(field.Type)}
	}

	var lines []string
	for _, name := range field.Names {
		if fn, ok := field.Type.(*ast.FuncType); ok {
			signature := strings.TrimPrefix(formatter.funcType(fn), "func")
			lines = append(lines, "\t"+name.Name+signature)
			continue
		}
		lines = append(lines, "\t"+name.Name+" "+formatter.typeText(field.Type))
	}
	return lines
}

func nodeString(set *token.FileSet, node any) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, set, node); err != nil {
		panic(err)
	}
	return buf.String()
}
