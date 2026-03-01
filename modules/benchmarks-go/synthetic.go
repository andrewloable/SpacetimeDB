package main

import (
	"fmt"

	spacetimedb "github.com/clockworklabs/spacetimedb-go-server"
	"github.com/clockworklabs/spacetimedb-go-server/sys"
	"github.com/clockworklabs/spacetimedb-go/bsatn"
)

func emptyReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {}

// ── Insert single row ─────────────────────────────────────────────────────────

func insertUnique0U32U64StrReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		spacetimedb.LogPanic("insert_unique_0_u32_u64_str: " + err.Error())
	}
	r := bsatn.NewReader(data)
	id, age, name, err := decodeU32U64Str(r)
	if err != nil {
		spacetimedb.LogPanic("insert_unique_0_u32_u64_str: decode: " + err.Error())
	}
	if _, err := unique0U32U64StrHandle.Insert(Unique0U32U64Str{Id: id, Age: age, Name: name}); err != nil {
		spacetimedb.LogPanic("insert_unique_0_u32_u64_str: insert: " + err.Error())
	}
}

func insertNoIndexU32U64StrReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		spacetimedb.LogPanic("insert_no_index_u32_u64_str: " + err.Error())
	}
	r := bsatn.NewReader(data)
	id, age, name, err := decodeU32U64Str(r)
	if err != nil {
		spacetimedb.LogPanic("insert_no_index_u32_u64_str: decode: " + err.Error())
	}
	if _, err := noIndexU32U64StrHandle.Insert(NoIndexU32U64Str{Id: id, Age: age, Name: name}); err != nil {
		spacetimedb.LogPanic("insert_no_index_u32_u64_str: insert: " + err.Error())
	}
}

func insertBtreeEachColumnU32U64StrReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		spacetimedb.LogPanic("insert_btree_each_column_u32_u64_str: " + err.Error())
	}
	r := bsatn.NewReader(data)
	id, age, name, err := decodeU32U64Str(r)
	if err != nil {
		spacetimedb.LogPanic("insert_btree_each_column_u32_u64_str: decode: " + err.Error())
	}
	if _, err := btreeEachColumnU32U64StrHandle.Insert(BtreeEachColumnU32U64Str{Id: id, Age: age, Name: name}); err != nil {
		spacetimedb.LogPanic("insert_btree_each_column_u32_u64_str: insert: " + err.Error())
	}
}

func insertUnique0U32U64U64Reducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		spacetimedb.LogPanic("insert_unique_0_u32_u64_u64: " + err.Error())
	}
	r := bsatn.NewReader(data)
	id, x, y, err := decodeU32U64U64(r)
	if err != nil {
		spacetimedb.LogPanic("insert_unique_0_u32_u64_u64: decode: " + err.Error())
	}
	if _, err := unique0U32U64U64Handle.Insert(Unique0U32U64U64{Id: id, X: x, Y: y}); err != nil {
		spacetimedb.LogPanic("insert_unique_0_u32_u64_u64: insert: " + err.Error())
	}
}

func insertNoIndexU32U64U64Reducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		spacetimedb.LogPanic("insert_no_index_u32_u64_u64: " + err.Error())
	}
	r := bsatn.NewReader(data)
	id, x, y, err := decodeU32U64U64(r)
	if err != nil {
		spacetimedb.LogPanic("insert_no_index_u32_u64_u64: decode: " + err.Error())
	}
	if _, err := noIndexU32U64U64Handle.Insert(NoIndexU32U64U64{Id: id, X: x, Y: y}); err != nil {
		spacetimedb.LogPanic("insert_no_index_u32_u64_u64: insert: " + err.Error())
	}
}

func insertBtreeEachColumnU32U64U64Reducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		spacetimedb.LogPanic("insert_btree_each_column_u32_u64_u64: " + err.Error())
	}
	r := bsatn.NewReader(data)
	id, x, y, err := decodeU32U64U64(r)
	if err != nil {
		spacetimedb.LogPanic("insert_btree_each_column_u32_u64_u64: decode: " + err.Error())
	}
	if _, err := btreeEachColumnU32U64U64Handle.Insert(BtreeEachColumnU32U64U64{Id: id, X: x, Y: y}); err != nil {
		spacetimedb.LogPanic("insert_btree_each_column_u32_u64_u64: insert: " + err.Error())
	}
}

// ── Insert bulk ───────────────────────────────────────────────────────────────

func insertBulkUnique0U32U64U64Reducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	n, _ := r.ReadArrayLen()
	for i := uint32(0); i < n; i++ {
		row, err := decodeUnique0U32U64U64(r)
		if err != nil {
			break
		}
		_, _ = unique0U32U64U64Handle.Insert(row)
	}
}

func insertBulkNoIndexU32U64U64Reducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	n, _ := r.ReadArrayLen()
	for i := uint32(0); i < n; i++ {
		row, err := decodeNoIndexU32U64U64(r)
		if err != nil {
			break
		}
		_, _ = noIndexU32U64U64Handle.Insert(row)
	}
}

func insertBulkBtreeEachColumnU32U64U64Reducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	n, _ := r.ReadArrayLen()
	for i := uint32(0); i < n; i++ {
		row, err := decodeBtreeEachColumnU32U64U64(r)
		if err != nil {
			break
		}
		_, _ = btreeEachColumnU32U64U64Handle.Insert(row)
	}
}

func insertBulkUnique0U32U64StrReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	n, _ := r.ReadArrayLen()
	for i := uint32(0); i < n; i++ {
		row, err := decodeUnique0U32U64Str(r)
		if err != nil {
			break
		}
		_, _ = unique0U32U64StrHandle.Insert(row)
	}
}

func insertBulkNoIndexU32U64StrReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	n, _ := r.ReadArrayLen()
	for i := uint32(0); i < n; i++ {
		row, err := decodeNoIndexU32U64Str(r)
		if err != nil {
			break
		}
		_, _ = noIndexU32U64StrHandle.Insert(row)
	}
}

func insertBulkBtreeEachColumnU32U64StrReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	n, _ := r.ReadArrayLen()
	for i := uint32(0); i < n; i++ {
		row, err := decodeBtreeEachColumnU32U64Str(r)
		if err != nil {
			break
		}
		_, _ = btreeEachColumnU32U64StrHandle.Insert(row)
	}
}

// ── Update bulk ───────────────────────────────────────────────────────────────

func updateBulkUnique0U32U64U64Reducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	rowCount, _ := r.ReadU32()

	var rows []Unique0U32U64U64
	for row, err := range unique0U32U64U64Handle.Iter() {
		if err != nil || uint32(len(rows)) >= rowCount {
			break
		}
		rows = append(rows, row)
	}
	if uint32(len(rows)) != rowCount {
		spacetimedb.LogPanic(fmt.Sprintf("update_bulk_unique_0_u32_u64_u64: expected %d rows, got %d", rowCount, len(rows)))
	}
	for _, row := range rows {
		_, _ = unique0U32U64U64IdIdx.Update(Unique0U32U64U64{Id: row.Id, X: row.X + 1, Y: row.Y})
	}
}

func updateBulkUnique0U32U64StrReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	rowCount, _ := r.ReadU32()

	var rows []Unique0U32U64Str
	for row, err := range unique0U32U64StrHandle.Iter() {
		if err != nil || uint32(len(rows)) >= rowCount {
			break
		}
		rows = append(rows, row)
	}
	if uint32(len(rows)) != rowCount {
		spacetimedb.LogPanic(fmt.Sprintf("update_bulk_unique_0_u32_u64_str: expected %d rows, got %d", rowCount, len(rows)))
	}
	for _, row := range rows {
		_, _ = unique0U32U64StrIdIdx.Update(Unique0U32U64Str{Id: row.Id, Age: row.Age + 1, Name: row.Name})
	}
}

// ── Iterate ───────────────────────────────────────────────────────────────────

func iterateUnique0U32U64StrReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	for _, err := range unique0U32U64StrHandle.Iter() {
		if err != nil {
			break
		}
	}
}

func iterateUnique0U32U64U64Reducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	for _, err := range unique0U32U64U64Handle.Iter() {
		if err != nil {
			break
		}
	}
}

// ── Filter ────────────────────────────────────────────────────────────────────

func filterUnique0U32U64StrByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	_, _ = unique0U32U64StrIdIdx.Find(id)
}

func filterNoIndexU32U64StrByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	for row, err := range noIndexU32U64StrHandle.Iter() {
		if err != nil {
			break
		}
		if row.Id == id {
			_ = row
			break
		}
	}
}

func filterBtreeEachColumnU32U64StrByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	for _, err := range btreeStrIdIdx.Filter(id) {
		if err != nil {
			break
		}
	}
}

func filterUnique0U32U64StrByNameReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	name, _ := r.ReadString()
	for row, err := range unique0U32U64StrHandle.Iter() {
		if err != nil {
			break
		}
		if row.Name == name {
			_ = row
		}
	}
}

func filterNoIndexU32U64StrByNameReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	name, _ := r.ReadString()
	for row, err := range noIndexU32U64StrHandle.Iter() {
		if err != nil {
			break
		}
		if row.Name == name {
			_ = row
		}
	}
}

func filterBtreeEachColumnU32U64StrByNameReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	name, _ := r.ReadString()
	for _, err := range btreeStrNameIdx.Filter(name) {
		if err != nil {
			break
		}
	}
}

func filterUnique0U32U64U64ByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	_, _ = unique0U32U64U64IdIdx.Find(id)
}

func filterNoIndexU32U64U64ByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	for row, err := range noIndexU32U64U64Handle.Iter() {
		if err != nil {
			break
		}
		if row.Id == id {
			_ = row
			break
		}
	}
}

func filterBtreeEachColumnU32U64U64ByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	for _, err := range btreeU64IdIdx.Filter(id) {
		if err != nil {
			break
		}
	}
}

func filterUnique0U32U64U64ByXReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	x, _ := r.ReadU64()
	for row, err := range unique0U32U64U64Handle.Iter() {
		if err != nil {
			break
		}
		if row.X == x {
			_ = row
		}
	}
}

func filterNoIndexU32U64U64ByXReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	x, _ := r.ReadU64()
	for row, err := range noIndexU32U64U64Handle.Iter() {
		if err != nil {
			break
		}
		if row.X == x {
			_ = row
		}
	}
}

func filterBtreeEachColumnU32U64U64ByXReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	x, _ := r.ReadU64()
	for _, err := range btreeU64XIdx.Filter(x) {
		if err != nil {
			break
		}
	}
}

func filterUnique0U32U64U64ByYReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	y, _ := r.ReadU64()
	for row, err := range unique0U32U64U64Handle.Iter() {
		if err != nil {
			break
		}
		if row.Y == y {
			_ = row
		}
	}
}

func filterNoIndexU32U64U64ByYReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	y, _ := r.ReadU64()
	for row, err := range noIndexU32U64U64Handle.Iter() {
		if err != nil {
			break
		}
		if row.Y == y {
			_ = row
		}
	}
}

func filterBtreeEachColumnU32U64U64ByYReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	y, _ := r.ReadU64()
	for _, err := range btreeU64YIdx.Filter(y) {
		if err != nil {
			break
		}
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func deleteUnique0U32U64StrByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	_, _ = unique0U32U64StrIdIdx.Delete(id)
}

func deleteUnique0U32U64U64ByIdReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	id, _ := r.ReadU32()
	_, _ = unique0U32U64U64IdIdx.Delete(id)
}

// ── Clear table (unimplemented) ───────────────────────────────────────────────

func clearTableUnimplementedReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	spacetimedb.LogPanic("Modules currently have no interface to clear a table")
}

// ── Count ─────────────────────────────────────────────────────────────────────

func countUnique0U32U64StrReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	count, _ := unique0U32U64StrHandle.Count()
	spacetimedb.LogInfo(fmt.Sprintf("COUNT: %d", count))
}

func countNoIndexU32U64StrReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	count, _ := noIndexU32U64StrHandle.Count()
	spacetimedb.LogInfo(fmt.Sprintf("COUNT: %d", count))
}

func countBtreeEachColumnU32U64StrReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	count, _ := btreeEachColumnU32U64StrHandle.Count()
	spacetimedb.LogInfo(fmt.Sprintf("COUNT: %d", count))
}

func countUnique0U32U64U64Reducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	count, _ := unique0U32U64U64Handle.Count()
	spacetimedb.LogInfo(fmt.Sprintf("COUNT: %d", count))
}

func countNoIndexU32U64U64Reducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	count, _ := noIndexU32U64U64Handle.Count()
	spacetimedb.LogInfo(fmt.Sprintf("COUNT: %d", count))
}

func countBtreeEachColumnU32U64U64Reducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	count, _ := btreeEachColumnU32U64U64Handle.Count()
	spacetimedb.LogInfo(fmt.Sprintf("COUNT: %d", count))
}

// ── Module-specific reducers ──────────────────────────────────────────────────

func fnWith1ArgsReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {}

func fnWith32ArgsReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {}

func printManyThingsReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	n, _ := r.ReadU32()
	for i := uint32(0); i < n; i++ {
		spacetimedb.LogInfo("hello again!")
	}
}
