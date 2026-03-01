//go:build tinygo

package spacetimedb

import (
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/clockworklabs/spacetimedb-go/bsatn"
	"github.com/clockworklabs/spacetimedb-go/types"
	"github.com/clockworklabs/spacetimedb-go-server/sys"
)

// ── Procedure types ───────────────────────────────────────────────────────────

// ProcedureContext is passed to every procedure function.
// Unlike ReducerContext, procedures can manage their own transaction lifecycle
// and make HTTP requests to external services.
type ProcedureContext struct {
	// Sender is the identity of the client that called this procedure.
	Sender types.Identity

	// ConnectionId is the connection ID of the caller, or nil for internal calls.
	ConnectionId *types.ConnectionId

	// Timestamp is the time at which this procedure was invoked.
	Timestamp types.Timestamp

	// Rng is a deterministic pseudo-random generator seeded from the call timestamp.
	Rng *rand.Rand

	// Auth provides access to the JWT claims for the current call.
	Auth AuthCtx
}

// StartMutTx begins a mutable transaction within this procedure.
// Returns the transaction timestamp in microseconds since Unix epoch.
// Must be balanced with CommitMutTx or AbortMutTx.
func (p *ProcedureContext) StartMutTx() (int64, error) {
	return sys.ProcedureStartMutTx()
}

// CommitMutTx commits the current mutable transaction.
func (p *ProcedureContext) CommitMutTx() error {
	return sys.ProcedureCommitMutTx()
}

// AbortMutTx aborts the current mutable transaction.
func (p *ProcedureContext) AbortMutTx() error {
	return sys.ProcedureAbortMutTx()
}

// WithTx runs fn inside a mutable transaction with automatic commit and retry.
// If fn returns an error, the transaction is aborted and the error is returned.
// If the commit fails, the transaction is retried once (matching the C# SDK behavior).
// The callback receives the transaction timestamp in microseconds since Unix epoch.
func (p *ProcedureContext) WithTx(fn func(txTimestamp int64) error) error {
	runOnce := func() (int64, error) {
		micros, err := p.StartMutTx()
		if err != nil {
			return 0, fmt.Errorf("start transaction: %w", err)
		}
		if err := fn(micros); err != nil {
			// fn failed — abort the transaction
			_ = p.AbortMutTx()
			return 0, err
		}
		return micros, nil
	}

	// First attempt
	_, err := runOnce()
	if err != nil {
		return err
	}

	// Try to commit
	if err := p.CommitMutTx(); err != nil {
		// Commit failed — retry once
		_, err2 := runOnce()
		if err2 != nil {
			return err2
		}
		if err2 := p.CommitMutTx(); err2 != nil {
			_ = p.AbortMutTx()
			return fmt.Errorf("commit after retry: %w", err2)
		}
	}
	return nil
}

// HttpRequestRaw makes an HTTP request from within a procedure using raw BSATN bytes.
// request is the BSATN-encoded HttpRequest; body is the optional body bytes.
// Returns the raw BSATN-encoded response and the response body bytes.
func (p *ProcedureContext) HttpRequestRaw(request, body []byte) ([]byte, []byte, error) {
	respSrc, bodySrc, err := sys.ProcedureHttpRequest(request, body)
	if err != nil {
		return nil, nil, err
	}
	respBytes, err := sys.ReadBytesSource(respSrc)
	if err != nil {
		return nil, nil, err
	}
	bodyBytes, err := sys.ReadBytesSource(bodySrc)
	if err != nil {
		return nil, nil, err
	}
	return respBytes, bodyBytes, nil
}

// Http makes a typed HTTP request from within a procedure.
// Unlike HttpRequestRaw which takes raw BSATN bytes, this method accepts an
// HttpRequest struct and returns a typed HttpResponse.
// The body parameter is the optional request body bytes (nil for no body).
func (p *ProcedureContext) Http(req HttpRequest, body []byte) (HttpResponse, []byte, error) {
	w := bsatn.NewWriter()
	WriteHttpRequest(w, req)
	rawResp, respBody, err := p.HttpRequestRaw(w.Bytes(), body)
	if err != nil {
		return HttpResponse{}, nil, err
	}
	r := bsatn.NewReader(rawResp)
	resp, err := ReadHttpResponse(r)
	if err != nil {
		return HttpResponse{}, nil, err
	}
	return resp, respBody, nil
}

// ── Procedure handler and registry ───────────────────────────────────────────

// ProcedureHandler implements a procedure body.
// It receives the caller context, a BytesSource for input args, and a BytesSink for output.
// Write the BSATN-encoded result to resultSink.
type ProcedureHandler func(ctx ProcedureContext, args sys.BytesSource, resultSink sys.BytesSink)

// ProcedureDef describes a procedure to be registered in the module.
type ProcedureDef struct {
	// Name is the procedure name as exposed to clients.
	Name string
	// Params describes the input parameter types and names.
	Params []ColumnDef
	// ReturnType is the AlgebraicType of the return value. Use types.ProductType{} for void.
	ReturnType interface{} // types.AlgebraicType
	// Visibility controls who can invoke this procedure.
	Visibility ReducerVisibility
}

var (
	procedureRegistry []ProcedureDef
	procedureHandlers []ProcedureHandler
)

// RegisterProcedureDef adds a procedure descriptor to the module registry.
func RegisterProcedureDef(def ProcedureDef) {
	procedureRegistry = append(procedureRegistry, def)
}

// RegisterProcedureHandler appends a procedure handler to the dispatch table.
// Must be called in the same order as RegisterProcedureDef.
func RegisterProcedureHandler(fn ProcedureHandler) {
	procedureHandlers = append(procedureHandlers, fn)
}

// ── WASM exports ──────────────────────────────────────────────────────────────

// __call_procedure__ is invoked by the SpacetimeDB host to execute a stored procedure.
//
//export __call_procedure__
func callProcedure(
	id uint32,
	sender0, sender1, sender2, sender3 uint64,
	connID0, connID1 uint64,
	timestamp uint64,
	args sys.BytesSource,
	resultSink sys.BytesSink,
) int16 {
	if int(id) >= len(procedureHandlers) {
		return -1
	}

	var senderBytes [32]byte
	binary.LittleEndian.PutUint64(senderBytes[0:8], sender0)
	binary.LittleEndian.PutUint64(senderBytes[8:16], sender1)
	binary.LittleEndian.PutUint64(senderBytes[16:24], sender2)
	binary.LittleEndian.PutUint64(senderBytes[24:32], sender3)
	sender := types.Identity(senderBytes)

	var connID *types.ConnectionId
	if connID0 != 0 || connID1 != 0 {
		var connBytes [16]byte
		binary.LittleEndian.PutUint64(connBytes[0:8], connID0)
		binary.LittleEndian.PutUint64(connBytes[8:16], connID1)
		c := types.ConnectionId(connBytes)
		connID = &c
	}

	ctx := ProcedureContext{
		Sender:       sender,
		ConnectionId: connID,
		Timestamp:    types.Timestamp{Microseconds: int64(timestamp)},
		Rng:          rand.New(rand.NewSource(int64(timestamp))),
		Auth:         newAuthCtxFromConnection(connID, sender),
	}

	procedureHandlers[id](ctx, args, resultSink)
	return 0
}

// ── Procedure section in module def ──────────────────────────────────────────

// writeProcedureDef serializes a RawProcedureDefV10 value.
func writeProcedureDef(w *bsatn.Writer, p ProcedureDef) {
	// source_name: RawIdentifier
	w.WriteString(p.Name)
	// params: ProductType (inline — not registered in typespace)
	w.WriteArrayLen(uint32(len(p.Params)))
	for _, param := range p.Params {
		name := param.Name
		writeOptString(w, &name)
		types.WriteAlgebraicType(w, param.Type)
	}
	// return_type: AlgebraicType — void (empty ProductType) when nil
	if at, ok := p.ReturnType.(types.AlgebraicType); ok {
		types.WriteAlgebraicType(w, at)
	} else {
		types.WriteAlgebraicType(w, types.ProductType{})
	}
	// visibility: FunctionVisibility
	w.WriteVariantTag(uint8(p.Visibility))
}
