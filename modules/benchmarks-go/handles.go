package main

import (
	spacetimedb "github.com/clockworklabs/spacetimedb-go-server"
	"github.com/clockworklabs/spacetimedb-go/bsatn"
)

// ── Synthetic table handles ───────────────────────────────────────────────────

var unique0U32U64StrHandle = spacetimedb.NewTableHandle(
	"unique_0_u32_u64_str", encodeUnique0U32U64Str, decodeUnique0U32U64Str)

var noIndexU32U64StrHandle = spacetimedb.NewTableHandle(
	"no_index_u32_u64_str", encodeNoIndexU32U64Str, decodeNoIndexU32U64Str)

var btreeEachColumnU32U64StrHandle = spacetimedb.NewTableHandle(
	"btree_each_column_u32_u64_str", encodeBtreeEachColumnU32U64Str, decodeBtreeEachColumnU32U64Str)

var unique0U32U64U64Handle = spacetimedb.NewTableHandle(
	"unique_0_u32_u64_u64", encodeUnique0U32U64U64, decodeUnique0U32U64U64)

var noIndexU32U64U64Handle = spacetimedb.NewTableHandle(
	"no_index_u32_u64_u64", encodeNoIndexU32U64U64, decodeNoIndexU32U64U64)

var btreeEachColumnU32U64U64Handle = spacetimedb.NewTableHandle(
	"btree_each_column_u32_u64_u64", encodeBtreeEachColumnU32U64U64, decodeBtreeEachColumnU32U64U64)

// ── Synthetic unique indexes ──────────────────────────────────────────────────

// unique0U32U64StrIdIdx — unique index on unique_0_u32_u64_str.id (the primary key).
var unique0U32U64StrIdIdx = spacetimedb.NewUniqueIndex[Unique0U32U64Str, uint32](
	"unique_0_u32_u64_str", "id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	encodeUnique0U32U64Str, decodeUnique0U32U64Str,
)

// unique0U32U64U64IdIdx — unique index on unique_0_u32_u64_u64.id (the primary key).
var unique0U32U64U64IdIdx = spacetimedb.NewUniqueIndex[Unique0U32U64U64, uint32](
	"unique_0_u32_u64_u64", "id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	encodeUnique0U32U64U64, decodeUnique0U32U64U64,
)

// ── Synthetic BTree indexes ───────────────────────────────────────────────────

// btreeStrIdIdx — BTree on btree_each_column_u32_u64_str.id
var btreeStrIdIdx = spacetimedb.NewBTreeIndex[BtreeEachColumnU32U64Str, uint32](
	"id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	decodeBtreeEachColumnU32U64Str,
)

// btreeStrAgeIdx — BTree on btree_each_column_u32_u64_str.age
var btreeStrAgeIdx = spacetimedb.NewBTreeIndex[BtreeEachColumnU32U64Str, uint64](
	"age",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	decodeBtreeEachColumnU32U64Str,
)

// btreeStrNameIdx — BTree on btree_each_column_u32_u64_str.name
var btreeStrNameIdx = spacetimedb.NewBTreeIndex[BtreeEachColumnU32U64Str, string](
	"name",
	func(w *bsatn.Writer, v string) { w.WriteString(v) },
	decodeBtreeEachColumnU32U64Str,
)

// btreeU64IdIdx — BTree on btree_each_column_u32_u64_u64.id
var btreeU64IdIdx = spacetimedb.NewBTreeIndex[BtreeEachColumnU32U64U64, uint32](
	"id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	decodeBtreeEachColumnU32U64U64,
)

// btreeU64XIdx — BTree on btree_each_column_u32_u64_u64.x
var btreeU64XIdx = spacetimedb.NewBTreeIndex[BtreeEachColumnU32U64U64, uint64](
	"x",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	decodeBtreeEachColumnU32U64U64,
)

// btreeU64YIdx — BTree on btree_each_column_u32_u64_u64.y
var btreeU64YIdx = spacetimedb.NewBTreeIndex[BtreeEachColumnU32U64U64, uint64](
	"y",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	decodeBtreeEachColumnU32U64U64,
)

// ── Circles table handles ─────────────────────────────────────────────────────

var entityHandle = spacetimedb.NewTableHandle("Entity", encodeEntity, decodeEntity)
var circleHandle = spacetimedb.NewTableHandle("Circle", encodeCircle, decodeCircle)
var foodHandle = spacetimedb.NewTableHandle("Food", encodeFood, decodeFood)

// entityIdIdx — unique index on Entity.id (primary key).
var entityIdIdx = spacetimedb.NewUniqueIndex[Entity, uint32](
	"Entity", "id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	encodeEntity, decodeEntity,
)

// circleEntityIdIdx — unique index on Circle.entity_id (primary key).
var circleEntityIdIdx = spacetimedb.NewUniqueIndex[Circle, uint32](
	"Circle", "entity_id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	encodeCircle, decodeCircle,
)

// circlePlayerIdIdx — BTree index on Circle.player_id.
var circlePlayerIdIdx = spacetimedb.NewBTreeIndex[Circle, uint32](
	"player_id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	decodeCircle,
)

// foodEntityIdIdx — unique index on Food.entity_id (primary key).
var foodEntityIdIdx = spacetimedb.NewUniqueIndex[Food, uint32](
	"Food", "entity_id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	encodeFood, decodeFood,
)

// ── IA Loop table handles ─────────────────────────────────────────────────────

var velocityHandle = spacetimedb.NewTableHandle("Velocity", encodeVelocity, decodeVelocity)
var positionHandle = spacetimedb.NewTableHandle("Position", encodePosition, decodePosition)
var gameEnemyAiAgentStateHandle = spacetimedb.NewTableHandle(
	"GameEnemyAiAgentState", encodeGameEnemyAiAgentState, decodeGameEnemyAiAgentState)
var gameTargetableStateHandle = spacetimedb.NewTableHandle(
	"GameTargetableState", encodeGameTargetableState, decodeGameTargetableState)
var gameLiveTargetableStateHandle = spacetimedb.NewTableHandle(
	"GameLiveTargetableState", encodeGameLiveTargetableState, decodeGameLiveTargetableState)
var gameMobileEntityStateHandle = spacetimedb.NewTableHandle(
	"GameMobileEntityState", encodeGameMobileEntityState, decodeGameMobileEntityState)
var gameEnemyStateHandle = spacetimedb.NewTableHandle(
	"GameEnemyState", encodeGameEnemyState, decodeGameEnemyState)
var gameHerdCacheHandle = spacetimedb.NewTableHandle(
	"GameHerdCache", encodeGameHerdCache, decodeGameHerdCache)

// ── IA Loop unique indexes ────────────────────────────────────────────────────

var velocityEntityIdIdx = spacetimedb.NewUniqueIndex[Velocity, uint32](
	"Velocity", "entity_id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	encodeVelocity, decodeVelocity,
)

var positionEntityIdIdx = spacetimedb.NewUniqueIndex[Position, uint32](
	"Position", "entity_id",
	func(w *bsatn.Writer, v uint32) { w.WriteU32(v) },
	encodePosition, decodePosition,
)

var gameEnemyAiAgentStateEntityIdIdx = spacetimedb.NewUniqueIndex[GameEnemyAiAgentState, uint64](
	"GameEnemyAiAgentState", "entity_id",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	encodeGameEnemyAiAgentState, decodeGameEnemyAiAgentState,
)

var gameTargetableStateEntityIdIdx = spacetimedb.NewUniqueIndex[GameTargetableState, uint64](
	"GameTargetableState", "entity_id",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	encodeGameTargetableState, decodeGameTargetableState,
)

var gameLiveTargetableStateEntityIdIdx = spacetimedb.NewUniqueIndex[GameLiveTargetableState, uint64](
	"GameLiveTargetableState", "entity_id",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	encodeGameLiveTargetableState, decodeGameLiveTargetableState,
)

var gameMobileEntityStateEntityIdIdx = spacetimedb.NewUniqueIndex[GameMobileEntityState, uint64](
	"GameMobileEntityState", "entity_id",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	encodeGameMobileEntityState, decodeGameMobileEntityState,
)

var gameEnemyStateEntityIdIdx = spacetimedb.NewUniqueIndex[GameEnemyState, uint64](
	"GameEnemyState", "entity_id",
	func(w *bsatn.Writer, v uint64) { w.WriteU64(v) },
	encodeGameEnemyState, decodeGameEnemyState,
)

var gameHerdCacheIdIdx = spacetimedb.NewUniqueIndex[GameHerdCache, int32](
	"GameHerdCache", "id",
	func(w *bsatn.Writer, v int32) { w.WriteI32(v) },
	encodeGameHerdCache, decodeGameHerdCache,
)

// ── IA Loop BTree indexes ─────────────────────────────────────────────────────

var gameLiveTargetableStateQuadIdx = spacetimedb.NewBTreeIndex[GameLiveTargetableState, int64](
	"quad",
	func(w *bsatn.Writer, v int64) { w.WriteI64(v) },
	decodeGameLiveTargetableState,
)

var gameMobileEntityStateLocationXIdx = spacetimedb.NewBTreeIndex[GameMobileEntityState, int32](
	"location_x",
	func(w *bsatn.Writer, v int32) { w.WriteI32(v) },
	decodeGameMobileEntityState,
)
