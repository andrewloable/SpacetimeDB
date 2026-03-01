//go:build tinygo

// Package spacetimedb provides the SpacetimeDB server-side Go SDK.
// It is compiled to WebAssembly using TinyGo and linked against the SpacetimeDB host ABI.
//
// Compile a module with:
//
//	tinygo build -target wasm -o module.wasm ./
package spacetimedb

import (
	"math/rand"

	"github.com/clockworklabs/spacetimedb-go-server/sys"
	"github.com/clockworklabs/spacetimedb-go/types"
)

// ReducerContext is passed to every reducer function.
// It provides the caller's identity, optional connection ID, the call timestamp,
// and a deterministic random number generator seeded from the timestamp.
// Table accessors are provided by the generated module bindings via package-level variables.
type ReducerContext struct {
	// Sender is the identity of the client or scheduled-reducer that called this reducer.
	Sender types.Identity

	// ConnectionId is the connection ID of the caller, or nil for scheduled reducers.
	ConnectionId *types.ConnectionId

	// Timestamp is the time at which this reducer was invoked.
	Timestamp types.Timestamp

	// Rng is a deterministic pseudo-random generator seeded from the call timestamp.
	// All reducers with the same timestamp produce the same random sequence,
	// ensuring determinism across replicas.
	Rng *rand.Rand

	// Auth provides access to the JWT claims for the current call.
	// Use Auth.GetJwt() to load claims lazily. IsInternal is true for scheduled reducers.
	Auth AuthCtx
}

// LogLevel controls the severity of a log message.
type LogLevel uint32

const (
	LogLevelError LogLevel = 0
	LogLevelWarn  LogLevel = 1
	LogLevelInfo  LogLevel = 2
	LogLevelDebug LogLevel = 3
	LogLevelTrace LogLevel = 4
	LogLevelPanic LogLevel = 101
)

// Log writes msg at the given level to the SpacetimeDB host log.
func Log(level LogLevel, msg string) {
	sys.ConsoleLog(uint32(level), "", "", 0, msg)
}

// LogError writes msg at Error level.
func LogError(msg string) { Log(LogLevelError, msg) }

// LogWarn writes msg at Warn level.
func LogWarn(msg string) { Log(LogLevelWarn, msg) }

// LogInfo writes msg at Info level.
func LogInfo(msg string) { Log(LogLevelInfo, msg) }

// LogDebug writes msg at Debug level.
func LogDebug(msg string) { Log(LogLevelDebug, msg) }

// LogTrace writes msg at Trace level.
func LogTrace(msg string) { Log(LogLevelTrace, msg) }

// LogPanic writes msg at Panic level then panics.
// The panic will cause the host to roll back the current transaction.
func LogPanic(msg string) {
	Log(LogLevelPanic, msg)
	panic(msg)
}
