//go:build tinygo

package sys

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
