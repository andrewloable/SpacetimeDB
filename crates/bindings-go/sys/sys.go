//go:build tinygo

// Package sys provides raw FFI bindings to the SpacetimeDB host ABI.
// This package is only for TinyGo WASM compilation targeting wasm32-unknown-unknown.
//
// Compile with:
//
//	tinygo build -target wasm -o module.wasm ./
//
// The SpacetimeDB host provides five import modules:
//   - spacetime_10.0  (stable core ABI + volatile schedule)
//   - spacetime_10.1  (bytes source length query)
//   - spacetime_10.2  (JWT credential lookup)
//   - spacetime_10.3  (procedure transactions + HTTP client)
//   - spacetime_10.4  (point index scan)
package sys

import "unsafe"

// ── Handle types ─────────────────────────────────────────────────────────────

// TableId identifies a table in the SpacetimeDB datastore.
type TableId = uint32

// IndexId identifies an index on a table.
type IndexId = uint32

// ColId identifies a column (u16 in Rust, passed as i32 on WASM boundary).
type ColId = uint32

// RowIter is a handle to an iterator over BSATN-encoded table rows.
type RowIter = uint32

// BytesSink is a write-only byte stream provided by the host (e.g. for __describe_module__).
type BytesSink = uint32

// BytesSource is a read-only byte stream provided by the host (e.g. reducer args).
type BytesSource = uint32

// InvalidRowIter is the sentinel value for an invalid RowIter handle.
const InvalidRowIter RowIter = 0xFFFFFFFF

// InvalidBytesSource is the sentinel value for an invalid BytesSource handle.
const InvalidBytesSource BytesSource = 0

// InvalidBytesSink is the sentinel value for an invalid BytesSink handle.
const InvalidBytesSink BytesSink = 0

// ── Error codes ───────────────────────────────────────────────────────────────

// Errno is a SpacetimeDB ABI error code (non-zero on failure).
type Errno uint32

// Error implements the error interface.
func (e Errno) Error() string {
	switch e {
	case ErrHostCallFailure:
		return "ABI called by host returned an error"
	case ErrNotInTransaction:
		return "ABI call can only be made while in a transaction"
	case ErrBsatnDecodeError:
		return "couldn't decode the BSATN to the expected type"
	case ErrNoSuchTable:
		return "no such table"
	case ErrNoSuchIndex:
		return "no such index"
	case ErrNoSuchIter:
		return "the provided row iterator is not valid"
	case ErrNoSuchConsoleTimer:
		return "the provided console timer does not exist"
	case ErrNoSuchBytes:
		return "the provided bytes source or sink is not valid"
	case ErrNoSpace:
		return "the provided sink has no more space left"
	case ErrWrongIndexAlgo:
		return "the index does not support range scans"
	case ErrBufferTooSmall:
		return "the provided buffer is not large enough to store the data"
	case ErrUniqueAlreadyExists:
		return "value with given unique identifier already exists"
	case ErrScheduleAtDelayTooLong:
		return "specified delay in scheduling row was too long"
	case ErrIndexNotUnique:
		return "the index was not unique"
	case ErrNoSuchRow:
		return "the row was not found"
	case ErrAutoIncOverflow:
		return "the auto-increment sequence overflowed"
	case ErrWouldBlockTransaction:
		return "attempted async or blocking op while holding open a transaction"
	case ErrTransactionNotAnonymous:
		return "not in an anonymous transaction"
	case ErrTransactionIsReadOnly:
		return "ABI call can only be made while within a mutable transaction"
	case ErrTransactionIsMut:
		return "ABI call can only be made while within a read-only transaction"
	case ErrHttpError:
		return "the HTTP request failed"
	default:
		return "unknown SpacetimeDB error"
	}
}

const (
	ErrHostCallFailure         Errno = 1
	ErrNotInTransaction        Errno = 2
	ErrBsatnDecodeError        Errno = 3
	ErrNoSuchTable             Errno = 4
	ErrNoSuchIndex             Errno = 5
	ErrNoSuchIter              Errno = 6
	ErrNoSuchConsoleTimer      Errno = 7
	ErrNoSuchBytes             Errno = 8
	ErrNoSpace                 Errno = 9
	ErrWrongIndexAlgo          Errno = 10
	ErrBufferTooSmall          Errno = 11
	ErrUniqueAlreadyExists     Errno = 12
	ErrScheduleAtDelayTooLong  Errno = 13
	ErrIndexNotUnique          Errno = 14
	ErrNoSuchRow               Errno = 15
	ErrAutoIncOverflow         Errno = 16
	ErrWouldBlockTransaction   Errno = 17
	ErrTransactionNotAnonymous Errno = 18
	ErrTransactionIsReadOnly   Errno = 19
	ErrTransactionIsMut        Errno = 20
	ErrHttpError               Errno = 21
)

// checkErr converts a u16 ABI return value to an error (nil if 0).
func checkErr(code uint32) error {
	if code == 0 {
		return nil
	}
	return Errno(code)
}

// ── spacetime_10.0 raw imports ────────────────────────────────────────────────

//go:wasmimport spacetime_10.0 table_id_from_name
//go:noescape
func rawTableIdFromName(namePtr unsafe.Pointer, nameLen uint32, out unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 index_id_from_name
//go:noescape
func rawIndexIdFromName(namePtr unsafe.Pointer, nameLen uint32, out unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 datastore_table_row_count
//go:noescape
func rawDatastoreTableRowCount(tableId TableId, out unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 datastore_table_scan_bsatn
//go:noescape
func rawDatastoreTableScanBsatn(tableId TableId, out unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 datastore_index_scan_range_bsatn
//go:noescape
func rawDatastoreIndexScanRangeBsatn(
	indexId IndexId,
	prefixPtr unsafe.Pointer, prefixLen uint32,
	prefixElems ColId,
	rstartPtr unsafe.Pointer, rstartLen uint32,
	rendPtr unsafe.Pointer, rendLen uint32,
	out unsafe.Pointer,
) uint32

//go:wasmimport spacetime_10.0 datastore_delete_by_index_scan_range_bsatn
//go:noescape
func rawDatastoreDeleteByIndexScanRangeBsatn(
	indexId IndexId,
	prefixPtr unsafe.Pointer, prefixLen uint32,
	prefixElems ColId,
	rstartPtr unsafe.Pointer, rstartLen uint32,
	rendPtr unsafe.Pointer, rendLen uint32,
	out unsafe.Pointer,
) uint32

//go:wasmimport spacetime_10.0 datastore_delete_all_by_eq_bsatn
//go:noescape
func rawDatastoreDeleteAllByEqBsatn(tableId TableId, relPtr unsafe.Pointer, relLen uint32, out unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 row_iter_bsatn_advance
//go:noescape
func rawRowIterBsatnAdvance(iter RowIter, bufPtr unsafe.Pointer, bufLenPtr unsafe.Pointer) int32

//go:wasmimport spacetime_10.0 row_iter_bsatn_close
//go:noescape
func rawRowIterBsatnClose(iter RowIter) uint32

//go:wasmimport spacetime_10.0 datastore_insert_bsatn
//go:noescape
func rawDatastoreInsertBsatn(tableId TableId, rowPtr unsafe.Pointer, rowLenPtr unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 datastore_update_bsatn
//go:noescape
func rawDatastoreUpdateBsatn(tableId TableId, indexId IndexId, rowPtr unsafe.Pointer, rowLenPtr unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 bytes_sink_write
//go:noescape
func rawBytesSinkWrite(sink BytesSink, bufPtr unsafe.Pointer, bufLenPtr unsafe.Pointer) uint32

//go:wasmimport spacetime_10.0 bytes_source_read
//go:noescape
func rawBytesSourceRead(source BytesSource, bufPtr unsafe.Pointer, bufLenPtr unsafe.Pointer) int32

//go:wasmimport spacetime_10.0 console_log
//go:noescape
func rawConsoleLog(
	level uint32,
	targetPtr unsafe.Pointer, targetLen uint32,
	filenamePtr unsafe.Pointer, filenameLen uint32,
	lineNumber uint32,
	messagePtr unsafe.Pointer, messageLen uint32,
)

//go:wasmimport spacetime_10.0 console_timer_start
//go:noescape
func rawConsoleTimerStart(namePtr unsafe.Pointer, nameLen uint32) uint32

//go:wasmimport spacetime_10.0 console_timer_end
//go:noescape
func rawConsoleTimerEnd(timerId uint32) uint32

//go:wasmimport spacetime_10.0 identity
//go:noescape
func rawIdentity(outPtr unsafe.Pointer)

//go:wasmimport spacetime_10.0 volatile_nonatomic_schedule_immediate
//go:noescape
func rawVolatileNonatomicScheduleImmediate(namePtr unsafe.Pointer, nameLen uint32, argsPtr unsafe.Pointer, argsLen uint32)

// ── spacetime_10.1 raw imports ────────────────────────────────────────────────

//go:wasmimport spacetime_10.1 bytes_source_remaining_length
//go:noescape
func rawBytesSourceRemainingLength(source BytesSource, out unsafe.Pointer) int32

// ── spacetime_10.2 raw imports ────────────────────────────────────────────────

//go:wasmimport spacetime_10.2 get_jwt
//go:noescape
func rawGetJwt(connectionIdPtr unsafe.Pointer, bytesSourceIdOut unsafe.Pointer) uint32

// ── spacetime_10.3 raw imports ────────────────────────────────────────────────

//go:wasmimport spacetime_10.3 procedure_start_mut_tx
//go:noescape
func rawProcedureStartMutTx(microsOut unsafe.Pointer) uint32

//go:wasmimport spacetime_10.3 procedure_commit_mut_tx
//go:noescape
func rawProcedureCommitMutTx() uint32

//go:wasmimport spacetime_10.3 procedure_abort_mut_tx
//go:noescape
func rawProcedureAbortMutTx() uint32

//go:wasmimport spacetime_10.3 procedure_http_request
//go:noescape
func rawProcedureHttpRequest(
	requestPtr unsafe.Pointer, requestLen uint32,
	bodyPtr unsafe.Pointer, bodyLen uint32,
	out unsafe.Pointer,
) uint32

// ── spacetime_10.4 raw imports ────────────────────────────────────────────────

//go:wasmimport spacetime_10.4 datastore_index_scan_point_bsatn
//go:noescape
func rawDatastoreIndexScanPointBsatn(indexId IndexId, pointPtr unsafe.Pointer, pointLen uint32, out unsafe.Pointer) uint32

//go:wasmimport spacetime_10.4 datastore_delete_by_index_scan_point_bsatn
//go:noescape
func rawDatastoreDeleteByIndexScanPointBsatn(indexId IndexId, pointPtr unsafe.Pointer, pointLen uint32, out unsafe.Pointer) uint32

// ── High-level wrappers ───────────────────────────────────────────────────────

// TableIdFromName resolves a table name to its numeric TableId.
func TableIdFromName(name string) (TableId, error) {
	b := []byte(name)
	var id TableId
	ret := rawTableIdFromName(unsafe.Pointer(&b[0]), uint32(len(b)), unsafe.Pointer(&id))
	return id, checkErr(ret)
}

// IndexIdFromName resolves an index name to its numeric IndexId.
func IndexIdFromName(name string) (IndexId, error) {
	b := []byte(name)
	var id IndexId
	ret := rawIndexIdFromName(unsafe.Pointer(&b[0]), uint32(len(b)), unsafe.Pointer(&id))
	return id, checkErr(ret)
}

// TableRowCount returns the number of rows in the given table.
func TableRowCount(tableId TableId) (uint64, error) {
	var count uint64
	ret := rawDatastoreTableRowCount(tableId, unsafe.Pointer(&count))
	return count, checkErr(ret)
}

// TableScanBsatn starts a full table scan and returns an iterator handle.
func TableScanBsatn(tableId TableId) (RowIter, error) {
	var iter RowIter
	ret := rawDatastoreTableScanBsatn(tableId, unsafe.Pointer(&iter))
	return iter, checkErr(ret)
}

// IndexScanRangeBsatn starts an index range scan and returns an iterator handle.
// prefix, rstart, rend are BSATN-encoded values (rstart and rend are Bound<AlgebraicValue>).
// Pass nil slices for unused bounds.
func IndexScanRangeBsatn(indexId IndexId, prefix []byte, prefixElems ColId, rstart, rend []byte) (RowIter, error) {
	var iter RowIter
	prefixPtr, prefixLen := slicePtr(prefix)
	rstartPtr, rstartLen := slicePtr(rstart)
	rendPtr, rendLen := slicePtr(rend)
	ret := rawDatastoreIndexScanRangeBsatn(
		indexId,
		prefixPtr, prefixLen, prefixElems,
		rstartPtr, rstartLen,
		rendPtr, rendLen,
		unsafe.Pointer(&iter),
	)
	return iter, checkErr(ret)
}

// DeleteByIndexScanRangeBsatn deletes rows matching the index scan range.
// Returns the number of rows deleted.
func DeleteByIndexScanRangeBsatn(indexId IndexId, prefix []byte, prefixElems ColId, rstart, rend []byte) (uint32, error) {
	var deleted uint32
	prefixPtr, prefixLen := slicePtr(prefix)
	rstartPtr, rstartLen := slicePtr(rstart)
	rendPtr, rendLen := slicePtr(rend)
	ret := rawDatastoreDeleteByIndexScanRangeBsatn(
		indexId,
		prefixPtr, prefixLen, prefixElems,
		rstartPtr, rstartLen,
		rendPtr, rendLen,
		unsafe.Pointer(&deleted),
	)
	return deleted, checkErr(ret)
}

// DeleteAllByEqBsatn deletes all rows equal to any row in rel (BSATN Vec<ProductValue>).
// Returns the number of rows deleted.
func DeleteAllByEqBsatn(tableId TableId, rel []byte) (uint32, error) {
	var deleted uint32
	relPtr, relLen := slicePtr(rel)
	ret := rawDatastoreDeleteAllByEqBsatn(tableId, relPtr, relLen, unsafe.Pointer(&deleted))
	return deleted, checkErr(ret)
}

// RowIterAdvance reads rows from the iterator into buf.
// Returns (bytesWritten, exhausted, error).
// When exhausted is true, the iterator handle is automatically closed by the host.
func RowIterAdvance(iter RowIter, buf []byte) (uint32, bool, error) {
	if len(buf) == 0 {
		return 0, false, nil
	}
	bufLen := uint32(len(buf))
	ret := rawRowIterBsatnAdvance(iter, unsafe.Pointer(&buf[0]), unsafe.Pointer(&bufLen))
	switch ret {
	case -1:
		return bufLen, true, nil
	case 0:
		return bufLen, false, nil
	default:
		return 0, false, Errno(uint32(ret))
	}
}

// RowIterClose destroys the iterator handle.
func RowIterClose(iter RowIter) error {
	return checkErr(rawRowIterBsatnClose(iter))
}

// CollectIter reads all BSATN bytes from iter into a single contiguous buffer.
// The returned bytes contain concatenated BSATN-encoded ProductValues.
func CollectIter(iter RowIter) ([]byte, error) {
	result := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, exhausted, err := RowIterAdvance(iter, buf)
		if err != nil {
			if err == ErrBufferTooSmall {
				// n holds the required buffer size for the next row.
				buf = make([]byte, n)
				continue
			}
			_ = RowIterClose(iter)
			return nil, err
		}
		result = append(result, buf[:n]...)
		if exhausted {
			return result, nil
		}
	}
}

// InsertBsatn inserts a BSATN-encoded row into the given table.
// On success, the returned slice may contain auto-generated column values.
func InsertBsatn(tableId TableId, row []byte) ([]byte, error) {
	if len(row) == 0 {
		return row, nil
	}
	buf := append([]byte(nil), row...) // copy so we can write back
	bufLen := uint32(len(buf))
	ret := rawDatastoreInsertBsatn(tableId, unsafe.Pointer(&buf[0]), unsafe.Pointer(&bufLen))
	if err := checkErr(ret); err != nil {
		return nil, err
	}
	return buf[:bufLen], nil
}

// UpdateBsatn updates a row identified by the unique index.
func UpdateBsatn(tableId TableId, indexId IndexId, row []byte) ([]byte, error) {
	if len(row) == 0 {
		return row, nil
	}
	buf := append([]byte(nil), row...)
	bufLen := uint32(len(buf))
	ret := rawDatastoreUpdateBsatn(tableId, indexId, unsafe.Pointer(&buf[0]), unsafe.Pointer(&bufLen))
	if err := checkErr(ret); err != nil {
		return nil, err
	}
	return buf[:bufLen], nil
}

// ReadBytesSource reads all bytes from a BytesSource into a []byte.
func ReadBytesSource(source BytesSource) ([]byte, error) {
	result := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		bufLen := uint32(len(buf))
		ret := rawBytesSourceRead(source, unsafe.Pointer(&buf[0]), unsafe.Pointer(&bufLen))
		result = append(result, buf[:bufLen]...)
		switch ret {
		case -1:
			return result, nil
		case 0:
			if bufLen == uint32(len(buf)) {
				buf = make([]byte, len(buf)*2)
			}
		default:
			return nil, Errno(uint32(ret))
		}
	}
}

// WriteBytesToSink writes all bytes to a BytesSink.
func WriteBytesToSink(sink BytesSink, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	offset := 0
	for offset < len(data) {
		chunk := data[offset:]
		n := uint32(len(chunk))
		ret := rawBytesSinkWrite(sink, unsafe.Pointer(&chunk[0]), unsafe.Pointer(&n))
		if err := checkErr(ret); err != nil {
			return err
		}
		offset += int(n)
	}
	return nil
}

// Log level constants matching SpacetimeDB's host logging levels.
const (
	LogError = uint32(0)
	LogWarn  = uint32(1)
	LogInfo  = uint32(2)
	LogDebug = uint32(3)
	LogTrace = uint32(4)
	LogPanic = uint32(101)
)

// ConsoleLog logs a message at the given level.
// target and filename may be empty strings (passed as NULL to host).
func ConsoleLog(level uint32, target, filename string, lineNumber uint32, message string) {
	msgBytes := []byte(message)
	var targetPtr, filenamePtr unsafe.Pointer
	var targetLen, filenameLen uint32
	if len(target) > 0 {
		tb := []byte(target)
		targetPtr = unsafe.Pointer(&tb[0])
		targetLen = uint32(len(tb))
	}
	if len(filename) > 0 {
		fb := []byte(filename)
		filenamePtr = unsafe.Pointer(&fb[0])
		filenameLen = uint32(len(fb))
	}
	rawConsoleLog(level, targetPtr, targetLen, filenamePtr, filenameLen, lineNumber, unsafe.Pointer(&msgBytes[0]), uint32(len(msgBytes)))
}

// ConsoleTimerStart begins a timing span and returns a timer handle.
func ConsoleTimerStart(name string) uint32 {
	b := []byte(name)
	return rawConsoleTimerStart(unsafe.Pointer(&b[0]), uint32(len(b)))
}

// ConsoleTimerEnd ends a timing span and logs its duration.
func ConsoleTimerEnd(timerId uint32) error {
	return checkErr(rawConsoleTimerEnd(timerId))
}

// Identity returns the 32-byte module identity.
func Identity() [32]byte {
	var out [32]byte
	rawIdentity(unsafe.Pointer(&out[0]))
	return out
}

// BytesSourceRemainingLength returns the remaining byte count in source (spacetime_10.1).
func BytesSourceRemainingLength(source BytesSource) (uint32, error) {
	var remaining uint32
	ret := rawBytesSourceRemainingLength(source, unsafe.Pointer(&remaining))
	if ret < 0 {
		return 0, Errno(uint32(-ret))
	}
	return remaining, nil
}

// GetJwt looks up the JWT payload for the given 16-byte ConnectionId (spacetime_10.2).
// Returns the BytesSource for the JWT payload, or InvalidBytesSource if none found.
func GetJwt(connectionId [16]byte) (BytesSource, error) {
	var src BytesSource
	ret := rawGetJwt(unsafe.Pointer(&connectionId[0]), unsafe.Pointer(&src))
	return src, checkErr(ret)
}

// VolatileNonatomicScheduleImmediate schedules a reducer call outside the current
// transaction. The reducer runs immediately after the current transaction commits.
// name is the reducer name; args is the BSATN-encoded argument payload.
func VolatileNonatomicScheduleImmediate(name string, args []byte) {
	nb := []byte(name)
	argsPtr, argsLen := slicePtr(args)
	rawVolatileNonatomicScheduleImmediate(unsafe.Pointer(&nb[0]), uint32(len(nb)), argsPtr, argsLen)
}

// ProcedureStartMutTx begins a mutable transaction within a procedure.
// Returns the transaction timestamp in microseconds since Unix epoch.
func ProcedureStartMutTx() (int64, error) {
	var micros int64
	ret := rawProcedureStartMutTx(unsafe.Pointer(&micros))
	return micros, checkErr(ret)
}

// ProcedureCommitMutTx commits the current mutable procedure transaction.
func ProcedureCommitMutTx() error {
	return checkErr(rawProcedureCommitMutTx())
}

// ProcedureAbortMutTx aborts the current mutable procedure transaction.
func ProcedureAbortMutTx() error {
	return checkErr(rawProcedureAbortMutTx())
}

// ProcedureHttpRequest makes an HTTP request from within a procedure.
// request is the BSATN-encoded HttpRequest struct; body is the optional body bytes.
// Returns two BytesSource handles: the first for the BSATN-encoded HttpResponse,
// the second for the response body bytes.
func ProcedureHttpRequest(request, body []byte) (BytesSource, BytesSource, error) {
	var pair [2]BytesSource
	reqPtr, reqLen := slicePtr(request)
	bodyPtr, bodyLen := slicePtr(body)
	ret := rawProcedureHttpRequest(reqPtr, reqLen, bodyPtr, bodyLen, unsafe.Pointer(&pair))
	return pair[0], pair[1], checkErr(ret)
}

// IndexScanPointBsatn looks up all rows matching a BSATN-encoded point value on an index.
// Returns an iterator handle over matching rows.
func IndexScanPointBsatn(indexId IndexId, point []byte) (RowIter, error) {
	var iter RowIter
	pointPtr, pointLen := slicePtr(point)
	ret := rawDatastoreIndexScanPointBsatn(indexId, pointPtr, pointLen, unsafe.Pointer(&iter))
	return iter, checkErr(ret)
}

// DeleteByIndexScanPointBsatn deletes all rows matching a BSATN-encoded point value on an index.
// Returns the number of rows deleted.
func DeleteByIndexScanPointBsatn(indexId IndexId, point []byte) (uint32, error) {
	var deleted uint32
	pointPtr, pointLen := slicePtr(point)
	ret := rawDatastoreDeleteByIndexScanPointBsatn(indexId, pointPtr, pointLen, unsafe.Pointer(&deleted))
	return deleted, checkErr(ret)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// slicePtr returns the unsafe.Pointer and length for a byte slice.
// Returns a non-nil sentinel pointer for empty slices to avoid NULL traps.
func slicePtr(b []byte) (unsafe.Pointer, uint32) {
	if len(b) == 0 {
		// Use a valid non-null pointer for zero-length slices.
		var sentinel [1]byte
		return unsafe.Pointer(&sentinel[0]), 0
	}
	return unsafe.Pointer(&b[0]), uint32(len(b))
}
