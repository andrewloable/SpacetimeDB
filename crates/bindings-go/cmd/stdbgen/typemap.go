package main

import (
	"fmt"
	"strings"
)

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

// customAlgebraicTypeOf builds the Go expression for a TypeExport's AlgebraicType.
func customAlgebraicTypeOf(te TypeExport) string {
	if len(te.Sum) > 0 {
		var variants []string
		for _, v := range te.Sum {
			nameExpr := fmt.Sprintf(`func() *string { s := %q; return &s }()`, v.Name)
			var innerType string
			if v.Type == "" {
				innerType = "types.ProductType{}" // unit variant
			} else {
				innerType = algebraicType(v.Type)
			}
			variants = append(variants, fmt.Sprintf(
				`{Name: %s, Type: %s}`, nameExpr, innerType))
		}
		return "types.SumType{Variants: []types.SumTypeVariant{" + strings.Join(variants, ", ") + "}}"
	}
	// Product type
	var elems []string
	for _, col := range te.Product {
		nameExpr := fmt.Sprintf(`func() *string { s := %q; return &s }()`, col.Name)
		elems = append(elems, fmt.Sprintf(`{Name: %s, Type: %s}`, nameExpr, algebraicType(col.Type)))
	}
	return "types.ProductType{Elements: []types.ProductTypeElement{" + strings.Join(elems, ", ") + "}}"
}
