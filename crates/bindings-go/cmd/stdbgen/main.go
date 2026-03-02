// stdbgen generates SpacetimeDB module registration code from a YAML schema.
//
// Usage:
//
//	stdbgen [-schema stdb.yaml] [-out generated_stdb.go] [-pkg main]
//
// Place a schema file in your module directory (default: stdb.yaml) and run:
//
//	go run github.com/clockworklabs/spacetimedb-go-server/cmd/stdbgen
//
// Or add the following to your main.go:
//
//	//go:generate go run github.com/clockworklabs/spacetimedb-go-server/cmd/stdbgen
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// ── Schema types ──────────────────────────────────────────────────────────────

// Schema is the top-level schema definition loaded from stdb.yaml.
type Schema struct {
	Package    string      `yaml:"package"`
	Module     string      `yaml:"module"`
	Tables     []Table     `yaml:"tables"`
	Reducers   []Reducer   `yaml:"reducers"`
	Procedures []Procedure `yaml:"procedures"`
	Views      []View      `yaml:"views"`
	Scheduled  []Scheduled `yaml:"scheduled"`
	Lifecycle  Lifecycle   `yaml:"lifecycle"`
}

// Table describes a SpacetimeDB table.
type Table struct {
	Name          string        `yaml:"name"`
	Columns       []Column      `yaml:"columns"`
	PrimaryKey    []string      `yaml:"primary_key"`
	UniqueIndexes []UniqueIndex `yaml:"unique_indexes"`
	BTreeIndexes  []BTreeIndex  `yaml:"btree_indexes"`
	Access        string        `yaml:"access"` // "public" or "private" (default: "public")
	IsEvent       bool          `yaml:"is_event"`
}

// UniqueIndex describes a unique index on a table.
type UniqueIndex struct {
	Name    string   `yaml:"name"`
	Columns []string `yaml:"columns"`
}

// BTreeIndex describes a BTree index on a table.
type BTreeIndex struct {
	Name    string   `yaml:"name"`
	Columns []string `yaml:"columns"`
}

// Column describes a single column of a table (or reducer/procedure/view parameter).
type Column struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// Reducer describes a SpacetimeDB reducer.
type Reducer struct {
	Name       string   `yaml:"name"`
	Params     []Column `yaml:"params"`
	Visibility string   `yaml:"visibility"` // "public" or "private" (default: "public")
}

// Procedure describes a SpacetimeDB stored procedure (can span transactions, make HTTP calls).
type Procedure struct {
	Name       string   `yaml:"name"`
	Params     []Column `yaml:"params"`
	ReturnType string   `yaml:"return_type"` // type name; empty = void
	Visibility string   `yaml:"visibility"`  // "public" or "private" (default: "public")
}

// View describes a SpacetimeDB view (read-only query, no transaction).
type View struct {
	Name        string   `yaml:"name"`
	Params      []Column `yaml:"params"`
	ReturnType  string   `yaml:"return_type"` // type name of the returned row
	IsPublic    bool     `yaml:"is_public"`
	IsAnonymous bool     `yaml:"is_anonymous"`
}

// Scheduled describes a scheduled reducer.
type Scheduled struct {
	Name  string `yaml:"name"`  // name of the reducer to schedule
	Table string `yaml:"table"` // name of the schedule-tracking table
}

// Lifecycle maps lifecycle events to reducer names.
type Lifecycle struct {
	OnInit       string `yaml:"on_init"`
	OnConnect    string `yaml:"on_connect"`
	OnDisconnect string `yaml:"on_disconnect"`
}

// ── Type mapping ──────────────────────────────────────────────────────────────

var typeMap = map[string]string{
	"String":    "types.AlgebraicString",
	"Bool":      "types.AlgebraicBool",
	"I8":        "types.AlgebraicI8",
	"U8":        "types.AlgebraicU8",
	"I16":       "types.AlgebraicI16",
	"U16":       "types.AlgebraicU16",
	"I32":       "types.AlgebraicI32",
	"U32":       "types.AlgebraicU32",
	"I64":       "types.AlgebraicI64",
	"U64":       "types.AlgebraicU64",
	"I128":      "types.AlgebraicI128",
	"U128":      "types.AlgebraicU128",
	"F32":       "types.AlgebraicF32",
	"F64":       "types.AlgebraicF64",
	"Bytes":     "types.AlgebraicBytes",
	"Identity":  "algebraicIdentity",
	"Timestamp": "types.AlgebraicTimestamp",
}

var readMethod = map[string]string{
	"String":    "r.ReadString()",
	"Bool":      "r.ReadBool()",
	"I8":        "r.ReadI8()",
	"U8":        "r.ReadU8()",
	"I16":       "r.ReadI16()",
	"U16":       "r.ReadU16()",
	"I32":       "r.ReadI32()",
	"U32":       "r.ReadU32()",
	"I64":       "r.ReadI64()",
	"U64":       "r.ReadU64()",
	"F32":       "r.ReadF32()",
	"F64":       "r.ReadF64()",
	"Bytes":     "r.ReadBytes()",
	"Identity":  "types.ReadIdentity(r)",
	"Timestamp": "types.ReadTimestamp(r)",
}

var goType = map[string]string{
	"String":    "string",
	"Bool":      "bool",
	"I8":        "int8",
	"U8":        "uint8",
	"I16":       "int16",
	"U16":       "uint16",
	"I32":       "int32",
	"U32":       "uint32",
	"I64":       "int64",
	"U64":       "uint64",
	"F32":       "float32",
	"F64":       "float64",
	"Bytes":     "[]byte",
	"Identity":  "types.Identity",
	"Timestamp": "types.Timestamp",
}

var writeMethod = map[string]string{
	"String":    "w.WriteString",
	"Bool":      "w.WriteBool",
	"I8":        "w.WriteI8",
	"U8":        "w.WriteU8",
	"I16":       "w.WriteI16",
	"U16":       "w.WriteU16",
	"I32":       "w.WriteI32",
	"U32":       "w.WriteU32",
	"I64":       "w.WriteI64",
	"U64":       "w.WriteU64",
	"F32":       "w.WriteF32",
	"F64":       "w.WriteF64",
	"Bytes":     "w.WriteBytes",
}

// specialWrite returns the write call for types that need a method call on the value.
var specialWrite = map[string]string{
	"Identity":  ".WriteBsatn(w)",
	"Timestamp": ".WriteBsatn(w)",
}

// ── Code template ─────────────────────────────────────────────────────────────

const codeTmpl = `// Code generated by stdbgen. DO NOT EDIT.
// Source: {{.SchemaFile}}

package {{.Package}}

import (
	spacetimedb "github.com/clockworklabs/spacetimedb-go-server"
	"github.com/clockworklabs/spacetimedb-go/bsatn"
	"github.com/clockworklabs/spacetimedb-go/types"
	"github.com/clockworklabs/spacetimedb-go-server/sys"
)

// Ensure imports are used even if generated encode/decode functions don't reference all of them.
var _ = types.AlgebraicString
var _ = bsatn.NewWriter
var _ *sys.BytesSource

// algebraicIdentity is the SATS type for a SpacetimeDB Identity (U256 newtype).
var algebraicIdentity = types.ProductType{
	Elements: []types.ProductTypeElement{
		{Name: strPtr("__identity__"), Type: types.AlgebraicU256},
	},
}

func strPtr(s string) *string { return &s }

{{range .Tables}}{{$tname := .Name}}
// ── {{$tname}} table ──────────────────────────────────────────────────────────

// {{$tname}} is a row in the {{$tname}} table.
type {{$tname}} struct {
{{- range .Columns}}
	{{title .Name}} {{goType .Type}}
{{- end}}
}

func encode{{$tname}}(w *bsatn.Writer, row {{$tname}}) {
{{- range .Columns}}
	{{encodeCol .Type .Name}}
{{- end}}
}

func decode{{$tname}}(r *bsatn.Reader) ({{$tname}}, error) {
	var row {{$tname}}
	var err error
{{- range .Columns}}
	row.{{title .Name}}, err = {{readMethod .Type}}
	if err != nil {
		return {{$tname}}{}, err
	}
{{- end}}
	return row, nil
}

var {{lower $tname}}Table = spacetimedb.NewTableHandle("{{$tname}}", encode{{$tname}}, decode{{$tname}})
{{range .UniqueIndexes}}
// {{lower $tname}}{{title .Name}}Idx is a unique index on {{$tname}} columns: {{join .Columns ", "}}
var {{lower $tname}}{{title .Name}}Idx = spacetimedb.NewUniqueIndex[{{$tname}}, {{idxKeyGoType $tname .Columns}}](
	"{{$tname}}",
	"{{.Name}}",
	{{idxKeyWrite $tname .Columns}},
	encode{{$tname}},
	decode{{$tname}},
)
{{end}}
{{end}}

func init() {
{{range .Tables}}{{$tname := .Name}}
	// Register table: {{$tname}}
	spacetimedb.RegisterTableDef(spacetimedb.TableDef{
		Name: "{{$tname}}",
		Columns: []spacetimedb.ColumnDef{
{{- range .Columns}}
			{Name: "{{.Name}}", Type: {{algebraicType .Type}}},
{{- end}}
		},
{{- if .PrimaryKey}}
		PrimaryKey: []uint16{ {{pkColIDs $tname .PrimaryKey .Columns}} },
{{- end}}
{{- if .UniqueIndexes}}
		Indexes: []spacetimedb.IndexDef{
{{- range .UniqueIndexes}}
			{SourceName: strPtr("{{.Name}}"), Algorithm: spacetimedb.IndexAlgorithmBTree, Columns: []uint16{ {{colIDs $tname .Columns}} }},
{{- end}}
		},
		Constraints: []spacetimedb.ConstraintDef{
{{- range .UniqueIndexes}}
			{SourceName: strPtr("{{.Name}}_constraint"), Columns: []uint16{ {{colIDs $tname .Columns}} }},
{{- end}}
		},
{{- end}}
		Access: {{tableAccess .Access}},
		IsEvent: {{.IsEvent}},
	})
{{end}}

{{range .Reducers}}{{$rname := .Name}}
	// Register reducer: {{$rname}}
	spacetimedb.RegisterReducerDef(spacetimedb.ReducerDef{
		Name: "{{$rname}}",
		Params: []spacetimedb.ColumnDef{
{{- range .Params}}
			{Name: "{{.Name}}", Type: {{algebraicType .Type}}},
{{- end}}
		},
		Visibility: {{reducerVisibility .Visibility}},
	})
	spacetimedb.RegisterReducerHandler(handle{{$rname}})
{{end}}

{{if .Lifecycle.OnInit}}
	// Register lifecycle: Init ({{.Lifecycle.OnInit}})
	spacetimedb.RegisterReducerDef(spacetimedb.ReducerDef{
		Name:       "{{.Lifecycle.OnInit}}",
		Params:     []spacetimedb.ColumnDef{},
		Visibility: spacetimedb.ReducerVisibilityPrivate,
	})
	spacetimedb.RegisterReducerHandler(handle{{.Lifecycle.OnInit}})
	spacetimedb.RegisterLifecycleDef(spacetimedb.LifecycleDef{
		Kind:    spacetimedb.LifecycleInit,
		Reducer: "{{.Lifecycle.OnInit}}",
	})
{{end}}
{{if .Lifecycle.OnConnect}}
	// Register lifecycle: OnConnect ({{.Lifecycle.OnConnect}})
	spacetimedb.RegisterReducerDef(spacetimedb.ReducerDef{
		Name:       "{{.Lifecycle.OnConnect}}",
		Params:     []spacetimedb.ColumnDef{},
		Visibility: spacetimedb.ReducerVisibilityPrivate,
	})
	spacetimedb.RegisterReducerHandler(handle{{.Lifecycle.OnConnect}})
	spacetimedb.RegisterLifecycleDef(spacetimedb.LifecycleDef{
		Kind:    spacetimedb.LifecycleOnConnect,
		Reducer: "{{.Lifecycle.OnConnect}}",
	})
{{end}}
{{if .Lifecycle.OnDisconnect}}
	// Register lifecycle: OnDisconnect ({{.Lifecycle.OnDisconnect}})
	spacetimedb.RegisterReducerDef(spacetimedb.ReducerDef{
		Name:       "{{.Lifecycle.OnDisconnect}}",
		Params:     []spacetimedb.ColumnDef{},
		Visibility: spacetimedb.ReducerVisibilityPrivate,
	})
	spacetimedb.RegisterReducerHandler(handle{{.Lifecycle.OnDisconnect}})
	spacetimedb.RegisterLifecycleDef(spacetimedb.LifecycleDef{
		Kind:    spacetimedb.LifecycleOnDisconnect,
		Reducer: "{{.Lifecycle.OnDisconnect}}",
	})
{{end}}

{{range .Procedures}}{{$pname := .Name}}
	// Register procedure: {{$pname}}
	spacetimedb.RegisterProcedureDef(spacetimedb.ProcedureDef{
		Name: "{{$pname}}",
		Params: []spacetimedb.ColumnDef{
{{- range .Params}}
			{Name: "{{.Name}}", Type: {{algebraicType .Type}}},
{{- end}}
		},
		Visibility: {{reducerVisibility .Visibility}},
	})
	spacetimedb.RegisterProcedureHandler(handle{{$pname}}Procedure)
{{end}}

{{range .Views}}{{$vname := .Name}}
	// Register view: {{$vname}}
	spacetimedb.RegisterViewDef(spacetimedb.ViewDef{
		Name: "{{$vname}}",
		Params: []spacetimedb.ColumnDef{
{{- range .Params}}
			{Name: "{{.Name}}", Type: {{algebraicType .Type}}},
{{- end}}
		},
		IsPublic:    {{.IsPublic}},
		IsAnonymous: {{.IsAnonymous}},
	})
{{- if .IsAnonymous}}
	spacetimedb.RegisterViewAnonHandler(handle{{$vname}}ViewAnon)
{{- else}}
	spacetimedb.RegisterViewHandler(handle{{$vname}}View)
{{- end}}
{{end}}

}

{{range .Reducers}}{{$rname := .Name}}
// handle{{$rname}} is the generated reducer handler skeleton for {{$rname}}.
// Replace this with your own implementation.
func handle{{$rname}}(ctx spacetimedb.ReducerContext, args sys.BytesSource) {
{{- if .Params}}
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		panic("{{$rname}}: failed to read args: " + err.Error())
	}
	r := bsatn.NewReader(data)
{{- range .Params}}
	{{lower .Name}}, err := {{readMethod .Type}}
	if err != nil {
		panic("{{$rname}}: failed to decode {{.Name}}: " + err.Error())
	}
	_ = {{lower .Name}}
{{- end}}
{{- end}}
	// TODO: implement {{$rname}}
	_ = ctx
}
{{end}}

{{if .Lifecycle.OnInit}}
// handle{{.Lifecycle.OnInit}} is the generated Init lifecycle handler.
func handle{{.Lifecycle.OnInit}}(ctx spacetimedb.ReducerContext, _ sys.BytesSource) {
	// TODO: implement module initialization
	_ = ctx
}
{{end}}
{{if .Lifecycle.OnConnect}}
// handle{{.Lifecycle.OnConnect}} is the generated OnConnect lifecycle handler.
func handle{{.Lifecycle.OnConnect}}(ctx spacetimedb.ReducerContext, _ sys.BytesSource) {
	// TODO: handle client connect for ctx.Sender
	_ = ctx
}
{{end}}
{{if .Lifecycle.OnDisconnect}}
// handle{{.Lifecycle.OnDisconnect}} is the generated OnDisconnect lifecycle handler.
func handle{{.Lifecycle.OnDisconnect}}(ctx spacetimedb.ReducerContext, _ sys.BytesSource) {
	// TODO: handle client disconnect for ctx.Sender
	_ = ctx
}
{{end}}

{{range .Procedures}}{{$pname := .Name}}
// handle{{$pname}}Procedure is the generated procedure handler skeleton for {{$pname}}.
func handle{{$pname}}Procedure(ctx spacetimedb.ProcedureContext, args sys.BytesSource, result sys.BytesSink) {
{{- if .Params}}
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		panic("{{$pname}}: failed to read args: " + err.Error())
	}
	r := bsatn.NewReader(data)
{{- range .Params}}
	{{lower .Name}}, err := {{readMethod .Type}}
	if err != nil {
		panic("{{$pname}}: failed to decode {{.Name}}: " + err.Error())
	}
	_ = {{lower .Name}}
{{- end}}
{{- end}}
	// TODO: implement {{$pname}}
	_ = ctx
	_ = result
}
{{end}}

{{range .Views}}{{$vname := .Name}}
{{- if .IsAnonymous}}
// handle{{$vname}}ViewAnon is the generated anonymous view handler skeleton for {{$vname}}.
func handle{{$vname}}ViewAnon(args sys.BytesSource, rows sys.BytesSink) {
	// TODO: implement {{$vname}} view; write BSATN-encoded rows to rows
	_ = args
	_ = rows
}
{{- else}}
// handle{{$vname}}View is the generated authenticated view handler skeleton for {{$vname}}.
func handle{{$vname}}View(sender types.Identity, connectionId *types.ConnectionId, args sys.BytesSource, rows sys.BytesSink) {
	// TODO: implement {{$vname}} view; write BSATN-encoded rows to rows
	_ = sender
	_ = connectionId
	_ = args
	_ = rows
}
{{- end}}
{{end}}
`

// ── Template helpers ──────────────────────────────────────────────────────────

func titleCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func lowerCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func tableAccess(access string) string {
	if strings.ToLower(access) == "private" {
		return "spacetimedb.TableAccessPrivate"
	}
	return "spacetimedb.TableAccessPublic"
}

func reducerVisibility(vis string) string {
	if strings.ToLower(vis) == "private" {
		return "spacetimedb.ReducerVisibilityPrivate"
	}
	return "spacetimedb.ReducerVisibilityClientCallable"
}

func algebraicType(t string) string {
	if v, ok := typeMap[t]; ok {
		return v
	}
	// Option<T> handling
	if strings.HasPrefix(t, "Option<") && strings.HasSuffix(t, ">") {
		inner := t[7 : len(t)-1]
		innerAlg := algebraicType(inner)
		return `types.SumType{Variants: []types.SumTypeVariant{{Name: func() *string { s := "some"; return &s }(), Type: ` + innerAlg + `}, {Name: func() *string { s := "none"; return &s }(), Type: types.ProductType{}}}}`
	}
	return "types.AlgebraicString" // fallback
}

func goTypeOf(t string) string {
	if v, ok := goType[t]; ok {
		return v
	}
	if strings.HasPrefix(t, "Option<") && strings.HasSuffix(t, ">") {
		inner := t[7 : len(t)-1]
		return "*" + goTypeOf(inner)
	}
	return "string" // fallback
}

func readMethodOf(t string) string {
	if v, ok := readMethod[t]; ok {
		return v
	}
	if strings.HasPrefix(t, "Option<") && strings.HasSuffix(t, ">") {
		inner := t[7 : len(t)-1]
		innerRead := readMethodOf(inner)
		return `bsatn.ReadOption(r, func(r *bsatn.Reader) (` + goTypeOf(inner) + `, error) { return ` + innerRead + ` })`
	}
	return "r.ReadString()"
}

func writeMethodOf(t string) string {
	if v, ok := writeMethod[t]; ok {
		return v
	}
	return "w.WriteString"
}

func specialWriteOf(t string) string {
	if v, ok := specialWrite[t]; ok {
		return v
	}
	return ""
}

// encodeExprOf returns the Go statement that writes a column value to w.
func encodeExprOf(t string, colName string) string {
	// field access expression, e.g. "row.Identity"
	field := "row." + titleCase(colName)
	if strings.HasPrefix(t, "Option<") && strings.HasSuffix(t, ">") {
		inner := t[7 : len(t)-1]
		gt := goTypeOf(inner)
		var innerExpr string
		if sw := specialWriteOf(inner); sw != "" {
			innerExpr = "v" + sw
		} else {
			innerExpr = writeMethodOf(inner) + "(v)"
		}
		return fmt.Sprintf("bsatn.WriteOption(w, %s, func(w *bsatn.Writer, v %s) { %s })", field, gt, innerExpr)
	}
	if sw := specialWriteOf(t); sw != "" {
		return field + sw
	}
	return writeMethodOf(t) + "(" + field + ")"
}

// idxKeyGoTypeOf returns the Go type for the first column of an index.
// tableColumns must be provided as context.
func idxKeyGoTypeOf(tableColumns map[string][]Column, tableName string, indexCols []string) string {
	if len(indexCols) == 0 {
		return "string"
	}
	cols := tableColumns[tableName]
	for _, col := range cols {
		if strings.EqualFold(col.Name, indexCols[0]) {
			return goTypeOf(col.Type)
		}
	}
	return "string"
}

// idxKeyWriteOf returns the encodeCol closure literal for the first column of an index.
func idxKeyWriteOf(tableColumns map[string][]Column, tableName string, indexCols []string) string {
	if len(indexCols) == 0 {
		return "func(w *bsatn.Writer, v string) { w.WriteString(v) }"
	}
	cols := tableColumns[tableName]
	for _, col := range cols {
		if strings.EqualFold(col.Name, indexCols[0]) {
			gt := goTypeOf(col.Type)
			if sw := specialWriteOf(col.Type); sw != "" {
				return fmt.Sprintf("func(w *bsatn.Writer, v %s) { v%s }", gt, sw)
			}
			wm := writeMethodOf(col.Type)
			return fmt.Sprintf("func(w *bsatn.Writer, v %s) { %s(v) }", gt, wm)
		}
	}
	return "func(w *bsatn.Writer, v string) { w.WriteString(v) }"
}

// colIndex returns the 0-based index of colName in the columns slice, or -1.
func colIndex(columns []Column, colName string) int {
	for i, c := range columns {
		if strings.EqualFold(c.Name, colName) {
			return i
		}
	}
	return -1
}

// pkColIDsOf builds the uint16 col-ID list for a primary_key spec.
func pkColIDsOf(tableName string, pks []string, columns []Column) string {
	var ids []string
	for _, pk := range pks {
		idx := colIndex(columns, pk)
		if idx < 0 {
			fmt.Fprintf(os.Stderr, "stdbgen: table %s: primary_key column %q not found\n", tableName, pk)
			idx = 0
		}
		ids = append(ids, fmt.Sprintf("%d", idx))
	}
	return strings.Join(ids, ", ")
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	schemaFile := flag.String("schema", "stdb.yaml", "path to the schema YAML file")
	outFile := flag.String("out", "generated_stdb.go", "path to the output Go file")
	pkg := flag.String("pkg", "", "package name (defaults to value from schema or 'main')")
	flag.Parse()

	data, err := os.ReadFile(*schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stdbgen: cannot read schema %s: %v\n", *schemaFile, err)
		os.Exit(1)
	}

	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		fmt.Fprintf(os.Stderr, "stdbgen: cannot parse schema: %v\n", err)
		os.Exit(1)
	}

	if *pkg != "" {
		schema.Package = *pkg
	}
	if schema.Package == "" {
		schema.Package = "main"
	}

	// Build a table-name → columns map so helpers can resolve column IDs.
	tableColumns := make(map[string][]Column)
	for _, t := range schema.Tables {
		tableColumns[t.Name] = t.Columns
	}

	funcMap := template.FuncMap{
		"title":             titleCase,
		"lower":             lowerCase,
		"tableAccess":       tableAccess,
		"reducerVisibility": reducerVisibility,
		"algebraicType":     algebraicType,
		"goType":            goTypeOf,
		"readMethod":        readMethodOf,
		"writeMethod":       writeMethodOf,
		"specialWrite":      specialWriteOf,
		"join":      strings.Join,
		"encodeCol": encodeExprOf,
		"pkColIDs": func(tableName string, pks []string, columns []Column) string {
			return pkColIDsOf(tableName, pks, columns)
		},
		"colIDs": func(tableName string, indexCols []string) string {
			cols := tableColumns[tableName]
			var ids []string
			for _, col := range indexCols {
				idx := colIndex(cols, col)
				if idx < 0 {
					fmt.Fprintf(os.Stderr, "stdbgen: table %s: index column %q not found\n", tableName, col)
					idx = 0
				}
				ids = append(ids, fmt.Sprintf("%d", idx))
			}
			return strings.Join(ids, ", ")
		},
		"idxKeyGoType": func(tableName string, indexCols []string) string {
			return idxKeyGoTypeOf(tableColumns, tableName, indexCols)
		},
		"idxKeyWrite": func(tableName string, indexCols []string) string {
			return idxKeyWriteOf(tableColumns, tableName, indexCols)
		},
	}

	tmpl, err := template.New("stdb").Funcs(funcMap).Parse(codeTmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stdbgen: template parse error: %v\n", err)
		os.Exit(1)
	}

	type tmplData struct {
		Schema
		SchemaFile string
	}
	d := tmplData{Schema: schema, SchemaFile: *schemaFile}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		fmt.Fprintf(os.Stderr, "stdbgen: template execute error: %v\n", err)
		os.Exit(1)
	}

	// Format the generated code.
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Write unformatted so the user can see what went wrong.
		_ = os.WriteFile(*outFile, buf.Bytes(), 0644)
		fmt.Fprintf(os.Stderr, "stdbgen: gofmt error (raw output written): %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outFile, formatted, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "stdbgen: cannot write %s: %v\n", *outFile, err)
		os.Exit(1)
	}

	fmt.Printf("stdbgen: wrote %s\n", *outFile)
}
