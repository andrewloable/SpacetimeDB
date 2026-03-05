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
		"title":      camelTitle,
		"camelTitle": camelTitle,
		"camelCols":  camelCols,
		"lower":      lowerCase,
		"tableAccess":       tableAccess,
		"reducerVisibility": reducerVisibility,
		"algebraicType":     algebraicType,
		"goType":            goTypeOf,
		"readMethod":        readMethodOf,
		"writeMethod":       writeMethodOf,
		"specialWrite":      specialWriteOf,
		"join":      strings.Join,
		"encodeCol": encodeExprOf,
		"add":       func(a, b int) int { return a + b },
		"customAlgebraicType": customAlgebraicTypeOf,
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

	// Pre-compute multi-column BTree prefix filter specs.
	var prefixFilters []PrefixFilterFunc
	for _, t := range schema.Tables {
		cols := t.Columns
		for _, idx := range t.BTreeIndexes {
			if len(idx.Columns) < 2 {
				continue // single-column already handled by Filter/FilterRange
			}
			idxCamel := camelTitle(idx.Name)
			varName := lowerCase(t.Name) + idxCamel + "BTreeIdx"
			// For each prefix length k from 1 to N-1, generate one filter function.
			for k := 1; k < len(idx.Columns); k++ {
				var prefixCols []PrefixColInfo
				for _, colName := range idx.Columns[:k] {
					gt := idxKeyGoTypeOf(tableColumns, t.Name, []string{colName})
					var encStmt string
					if sw := specialWriteOf(colTypeOf(cols, colName)); sw != "" {
						encStmt = lowerCase(colName) + sw
					} else {
						encStmt = writeMethodOf(colTypeOf(cols, colName)) + "(" + lowerCase(colName) + ")"
					}
					prefixCols = append(prefixCols, PrefixColInfo{
						Name:       colName,
						GoType:     gt,
						EncodeStmt: encStmt,
					})
				}
				trailingColName := idx.Columns[k]
				trailingType := idxKeyGoTypeOf(tableColumns, t.Name, []string{trailingColName})
				trailingEncExpr := idxKeyWriteOf(tableColumns, t.Name, []string{trailingColName})
				// Build function name: FilterTableBy{col1}...And{colK}Range
				var nameParts []string
				for _, c := range prefixCols {
					nameParts = append(nameParts, camelTitle(c.Name))
				}
				funcName := "Filter" + t.Name + "By" + strings.Join(nameParts, "") + "And" + camelTitle(trailingColName) + "Range"
				prefixFilters = append(prefixFilters, PrefixFilterFunc{
					FuncName:           funcName,
					TableName:          t.Name,
					IdxVarName:         varName,
					PrefixCols:         prefixCols,
					TrailingType:       trailingType,
					TrailingEncodeExpr: trailingEncExpr,
					NumPrefix:          uint32(k),
				})
			}
		}
	}

	type tmplData struct {
		Schema
		SchemaFile    string
		PrefixFilters []PrefixFilterFunc
	}
	d := tmplData{Schema: schema, SchemaFile: *schemaFile, PrefixFilters: prefixFilters}

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
