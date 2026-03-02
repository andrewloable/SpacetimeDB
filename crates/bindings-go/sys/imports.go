//go:build tinygo

package sys

import "unsafe"

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
