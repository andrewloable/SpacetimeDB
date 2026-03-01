//go:build tinygo

package spacetimedb

import (
	"iter"

	"github.com/clockworklabs/spacetimedb-go/bsatn"
	"github.com/clockworklabs/spacetimedb-go-server/sys"
)

// TableHandle provides Insert, Iter, Count, and Delete operations for a table.
// Generated module bindings create one TableHandle per table and embed type-specific
// accessor methods (FindBy*, DeleteBy*, Filter*) as wrapper types.
type TableHandle[Row any] struct {
	name    string
	tableId sys.TableId
	ready   bool
	encode  func(*bsatn.Writer, Row)
	decode  func(*bsatn.Reader) (Row, error)
}

// NewTableHandle creates a TableHandle for the named table.
// The table ID is resolved lazily on first use.
func NewTableHandle[Row any](
	name string,
	encode func(*bsatn.Writer, Row),
	decode func(*bsatn.Reader) (Row, error),
) *TableHandle[Row] {
	return &TableHandle[Row]{
		name:   name,
		encode: encode,
		decode: decode,
	}
}

// id returns the cached TableId, resolving it from the host on first call.
func (h *TableHandle[Row]) id() (sys.TableId, error) {
	if !h.ready {
		tid, err := sys.TableIdFromName(h.name)
		if err != nil {
			return 0, err
		}
		h.tableId = tid
		h.ready = true
	}
	return h.tableId, nil
}

// Count returns the number of rows in the table.
func (h *TableHandle[Row]) Count() (uint64, error) {
	tid, err := h.id()
	if err != nil {
		return 0, err
	}
	return sys.TableRowCount(tid)
}

// Insert encodes row as BSATN and inserts it into the table.
// The returned row reflects any auto-increment or generated column values.
func (h *TableHandle[Row]) Insert(row Row) (Row, error) {
	tid, err := h.id()
	if err != nil {
		return row, err
	}
	w := bsatn.NewWriter()
	h.encode(w, row)
	out, err := sys.InsertBsatn(tid, w.Bytes())
	if err != nil {
		return row, err
	}
	r := bsatn.NewReader(out)
	return h.decode(r)
}

// Delete removes the row equal to row from the table.
// Returns true if the row was found and deleted.
func (h *TableHandle[Row]) Delete(row Row) (bool, error) {
	tid, err := h.id()
	if err != nil {
		return false, err
	}
	w := bsatn.NewWriter()
	h.encode(w, row)
	// Wrap in a 1-element Vec<ProductValue> as required by the ABI.
	payload := encodeRowVec(w.Bytes())
	deleted, err := sys.DeleteAllByEqBsatn(tid, payload)
	return deleted > 0, err
}

// Iter iterates over all rows in the table in an unspecified order.
// The iterator is safe to break early.
func (h *TableHandle[Row]) Iter() iter.Seq2[Row, error] {
	return func(yield func(Row, error) bool) {
		tid, err := h.id()
		if err != nil {
			var zero Row
			yield(zero, err)
			return
		}
		rowIter, err := sys.TableScanBsatn(tid)
		if err != nil {
			var zero Row
			yield(zero, err)
			return
		}
		iterRows(rowIter, h.decode, yield)
	}
}

// UniqueIndex provides Find, Delete, and Update operations via a unique column index.
type UniqueIndex[Row any, Col any] struct {
	indexName string
	indexId   sys.IndexId
	ready     bool
	tableId   sys.TableId
	tableName string
	tableReady bool
	encodeCol func(*bsatn.Writer, Col)
	encodeRow func(*bsatn.Writer, Row)
	decodeRow func(*bsatn.Reader) (Row, error)
}

// NewUniqueIndex creates a UniqueIndex for the named index.
func NewUniqueIndex[Row any, Col any](
	tableName, indexName string,
	encodeCol func(*bsatn.Writer, Col),
	encodeRow func(*bsatn.Writer, Row),
	decodeRow func(*bsatn.Reader) (Row, error),
) *UniqueIndex[Row, Col] {
	return &UniqueIndex[Row, Col]{
		indexName: indexName,
		tableName: tableName,
		encodeCol: encodeCol,
		encodeRow: encodeRow,
		decodeRow: decodeRow,
	}
}

func (idx *UniqueIndex[Row, Col]) iid() (sys.IndexId, error) {
	if !idx.ready {
		id, err := sys.IndexIdFromName(idx.indexName)
		if err != nil {
			return 0, err
		}
		idx.indexId = id
		idx.ready = true
	}
	return idx.indexId, nil
}

func (idx *UniqueIndex[Row, Col]) tid() (sys.TableId, error) {
	if !idx.tableReady {
		id, err := sys.TableIdFromName(idx.tableName)
		if err != nil {
			return 0, err
		}
		idx.tableId = id
		idx.tableReady = true
	}
	return idx.tableId, nil
}

// Find returns the row whose unique column equals col, or nil if not found.
func (idx *UniqueIndex[Row, Col]) Find(col Col) (*Row, error) {
	iid, err := idx.iid()
	if err != nil {
		return nil, err
	}
	w := bsatn.NewWriter()
	idx.encodeCol(w, col)
	prefix := w.Bytes()
	rowIter, err := sys.IndexScanRangeBsatn(iid, prefix, 1, nil, nil)
	if err != nil {
		return nil, err
	}
	var found *Row
	iterRows(rowIter, idx.decodeRow, func(row Row, e error) bool {
		if e != nil {
			err = e
			return false
		}
		r := row
		found = &r
		return false // stop after first match
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

// Delete removes the row whose unique column equals col.
// Returns true if a row was found and deleted.
func (idx *UniqueIndex[Row, Col]) Delete(col Col) (bool, error) {
	iid, err := idx.iid()
	if err != nil {
		return false, err
	}
	w := bsatn.NewWriter()
	idx.encodeCol(w, col)
	prefix := w.Bytes()
	deleted, err := sys.DeleteByIndexScanRangeBsatn(iid, prefix, 1, nil, nil)
	return deleted > 0, err
}

// Update replaces the row identified by the unique column value in row.
func (idx *UniqueIndex[Row, Col]) Update(row Row) (Row, error) {
	tid, err := idx.tid()
	if err != nil {
		return row, err
	}
	iid, err := idx.iid()
	if err != nil {
		return row, err
	}
	w := bsatn.NewWriter()
	idx.encodeRow(w, row)
	out, err := sys.UpdateBsatn(tid, iid, w.Bytes())
	if err != nil {
		return row, err
	}
	r := bsatn.NewReader(out)
	return idx.decodeRow(r)
}

// BTreeIndex provides Filter operations over a btree-indexed column.
type BTreeIndex[Row any, Col any] struct {
	indexName string
	indexId   sys.IndexId
	ready     bool
	encodeCol func(*bsatn.Writer, Col)
	decodeRow func(*bsatn.Reader) (Row, error)
}

// NewBTreeIndex creates a BTreeIndex for the named index.
func NewBTreeIndex[Row any, Col any](
	indexName string,
	encodeCol func(*bsatn.Writer, Col),
	decodeRow func(*bsatn.Reader) (Row, error),
) *BTreeIndex[Row, Col] {
	return &BTreeIndex[Row, Col]{
		indexName: indexName,
		encodeCol: encodeCol,
		decodeRow: decodeRow,
	}
}

func (idx *BTreeIndex[Row, Col]) iid() (sys.IndexId, error) {
	if !idx.ready {
		id, err := sys.IndexIdFromName(idx.indexName)
		if err != nil {
			return 0, err
		}
		idx.indexId = id
		idx.ready = true
	}
	return idx.indexId, nil
}

// Filter returns an iterator over all rows whose indexed column equals col.
func (idx *BTreeIndex[Row, Col]) Filter(col Col) iter.Seq2[Row, error] {
	return func(yield func(Row, error) bool) {
		iid, err := idx.iid()
		if err != nil {
			var zero Row
			yield(zero, err)
			return
		}
		w := bsatn.NewWriter()
		idx.encodeCol(w, col)
		prefix := w.Bytes()
		rowIter, err := sys.IndexScanRangeBsatn(iid, prefix, 1, nil, nil)
		if err != nil {
			var zero Row
			yield(zero, err)
			return
		}
		iterRows(rowIter, idx.decodeRow, yield)
	}
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// iterRows reads all rows from a RowIter, decoding each with decode and yielding
// via yield. Stops when yield returns false or the iterator is exhausted.
func iterRows[Row any](rowIter sys.RowIter, decode func(*bsatn.Reader) (Row, error), yield func(Row, error) bool) {
	data, err := sys.CollectIter(rowIter)
	if err != nil {
		var zero Row
		yield(zero, err)
		return
	}
	r := bsatn.NewReader(data)
	for r.Remaining() > 0 {
		row, err := decode(r)
		if !yield(row, err) || err != nil {
			return
		}
	}
}

// encodeRowVec wraps a single BSATN-encoded row in a Vec<ProductValue> envelope
// (4-byte LE count = 1, then the row bytes), as required by DeleteAllByEqBsatn.
func encodeRowVec(rowBytes []byte) []byte {
	w := bsatn.NewWriter()
	w.WriteArrayLen(1)
	w.WriteRaw(rowBytes)
	return w.Bytes()
}
