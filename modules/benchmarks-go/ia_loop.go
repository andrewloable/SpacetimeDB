package main

import (
	"fmt"

	spacetimedb "github.com/clockworklabs/spacetimedb-go-server"
	"github.com/clockworklabs/spacetimedb-go-server/sys"
	"github.com/clockworklabs/spacetimedb-go/bsatn"
)

func momentMilliseconds() uint64 {
	return 1
}

// calculateHash is a deterministic bit-mixing function used to update quad values
// in the enemy AI benchmark loop. The exact result doesn't need to match Rust's
// DefaultHasher — any consistent permutation works for the benchmark.
func calculateHash(v int64) int64 {
	u := uint64(v)
	u ^= u >> 30
	u *= 0xbf58476d1ce4e5b9
	u ^= u >> 27
	u *= 0x94d049bb133111eb
	u ^= u >> 31
	return int64(u)
}

const maxMoveTimestamps = 20

func moveAgent(agent *GameEnemyAiAgentState, currentTimeMs uint64) {
	entityId := agent.EntityId

	// Find + no-op update enemy state (benchmark workload).
	enemy, err := gameEnemyStateEntityIdIdx.Find(entityId)
	if err != nil || enemy == nil {
		spacetimedb.LogPanic("GameEnemyState Entity ID not found")
	}
	_, _ = gameEnemyStateEntityIdIdx.Update(*enemy)

	agent.NextActionTimestamp = currentTimeMs + 2000

	// Track movement history (cap at maxMoveTimestamps).
	agent.LastMoveTimestamps = append(agent.LastMoveTimestamps, currentTimeMs)
	if len(agent.LastMoveTimestamps) > maxMoveTimestamps {
		agent.LastMoveTimestamps = agent.LastMoveTimestamps[1:]
	}

	// Update targetable quad with a hash of the previous value.
	targetable, err := gameTargetableStateEntityIdIdx.Find(entityId)
	if err != nil || targetable == nil {
		spacetimedb.LogPanic("GameTargetableState Entity ID not found")
	}
	newHash := calculateHash(targetable.Quad)
	targetable.Quad = newHash
	_, _ = gameTargetableStateEntityIdIdx.Update(*targetable)

	// If the entity is alive, also update the live targetable state.
	liveTarget, err2 := gameLiveTargetableStateEntityIdIdx.Find(entityId)
	if err2 == nil && liveTarget != nil {
		_, _ = gameLiveTargetableStateEntityIdIdx.Delete(entityId)
		_, _ = gameLiveTargetableStateHandle.Insert(GameLiveTargetableState{
			EntityId: entityId,
			Quad:     newHash,
		})
	}

	// Update mobile entity location.
	mobile, err := gameMobileEntityStateEntityIdIdx.Find(entityId)
	if err != nil || mobile == nil {
		spacetimedb.LogPanic("GameMobileEntityState Entity ID not found")
	}
	newMobile := GameMobileEntityState{
		EntityId:  entityId,
		LocationX: mobile.LocationX + 1,
		LocationY: mobile.LocationY + 1,
		Timestamp: agent.NextActionTimestamp,
	}

	_, _ = gameEnemyAiAgentStateEntityIdIdx.Update(*agent)
	_, _ = gameMobileEntityStateEntityIdIdx.Update(newMobile)
}

func agentLoop(agent GameEnemyAiAgentState, currentTimeMs uint64) {
	entityId := agent.EntityId

	// Load coordinates as benchmark workload (not functionally used).
	gameMobileEntityStateEntityIdIdx.Find(entityId)

	agentEntity, err := gameEnemyStateEntityIdIdx.Find(entityId)
	if err != nil || agentEntity == nil {
		spacetimedb.LogPanic("GameEnemyState Entity ID not found")
	}

	// Load herd as benchmark workload.
	gameHerdCacheIdIdx.Find(agentEntity.HerdId)

	moveAgent(&agent, currentTimeMs)
}

func getTargetablesNearQuad(entityId, numPlayers uint64) []GameTargetableState {
	result := make([]GameTargetableState, 0, 4)
	for id := entityId; id < numPlayers; id++ {
		for t, err := range gameLiveTargetableStateQuadIdx.Filter(int64(id)) {
			if err != nil {
				break
			}
			targetable, err2 := gameTargetableStateEntityIdIdx.Find(t.EntityId)
			if err2 != nil || targetable == nil {
				spacetimedb.LogPanic("Identity not found")
			}
			result = append(result, *targetable)
		}
	}
	return result
}

func insertBulkPosition(count uint32) {
	for id := uint32(0); id < count; id++ {
		x := float32(id)
		y := float32(id + 5)
		z := float32(id * 5)
		p := Position{
			EntityId: id,
			X:        x,
			Y:        y,
			Z:        z,
			Vx:       x + 10.0,
			Vy:       y + 20.0,
			Vz:       z + 30.0,
		}
		if _, err := positionHandle.Insert(p); err != nil {
			spacetimedb.LogPanic("insert_bulk_position: " + err.Error())
		}
	}
	spacetimedb.LogInfo(fmt.Sprintf("INSERT POSITION: %d", count))
}

func insertBulkPositionReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	count, _ := r.ReadU32()
	insertBulkPosition(count)
}

func insertBulkVelocity(count uint32) {
	for id := uint32(0); id < count; id++ {
		v := Velocity{
			EntityId: id,
			X:        float32(id),
			Y:        float32(id + 5),
			Z:        float32(id * 5),
		}
		if _, err := velocityHandle.Insert(v); err != nil {
			spacetimedb.LogPanic("insert_bulk_velocity: " + err.Error())
		}
	}
	spacetimedb.LogInfo(fmt.Sprintf("INSERT VELOCITY: %d", count))
}

func insertBulkVelocityReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	count, _ := r.ReadU32()
	insertBulkVelocity(count)
}

func updatePositionAll(expected uint32) {
	count := 0
	for pos, err := range positionHandle.Iter() {
		if err != nil {
			break
		}
		pos.X += pos.Vx
		pos.Y += pos.Vy
		pos.Z += pos.Vz
		_, _ = positionEntityIdIdx.Update(pos)
		count++
	}
	spacetimedb.LogInfo(fmt.Sprintf("UPDATE POSITION ALL: %d, processed: %d", expected, count))
}

func updatePositionAllReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	expected, _ := r.ReadU32()
	updatePositionAll(expected)
}

func updatePositionWithVelocity(expected uint32) {
	count := 0
	for vel, err := range velocityHandle.Iter() {
		if err != nil {
			break
		}
		pos, err2 := positionEntityIdIdx.Find(vel.EntityId)
		if err2 != nil || pos == nil {
			continue
		}
		pos.X += vel.X
		pos.Y += vel.Y
		pos.Z += vel.Z
		_, _ = positionEntityIdIdx.Update(*pos)
		count++
	}
	spacetimedb.LogInfo(fmt.Sprintf("UPDATE POSITION BY VELOCITY: %d, processed: %d", expected, count))
}

func updatePositionWithVelocityReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	expected, _ := r.ReadU32()
	updatePositionWithVelocity(expected)
}

func insertWorld(players uint64) {
	for i := uint64(0); i < players; i++ {
		id := i
		var nextActionTimestamp uint64
		if i&2 == 2 {
			nextActionTimestamp = momentMilliseconds() + 2000
		} else {
			nextActionTimestamp = momentMilliseconds()
		}

		_, _ = gameEnemyAiAgentStateHandle.Insert(GameEnemyAiAgentState{
			EntityId:            id,
			LastMoveTimestamps:  []uint64{id, 0, id * 2},
			NextActionTimestamp: nextActionTimestamp,
			Action:              AgentActionIdle,
		})
		_, _ = gameLiveTargetableStateHandle.Insert(GameLiveTargetableState{
			EntityId: id,
			Quad:     int64(id),
		})
		_, _ = gameTargetableStateHandle.Insert(GameTargetableState{
			EntityId: id,
			Quad:     int64(id),
		})
		_, _ = gameMobileEntityStateHandle.Insert(GameMobileEntityState{
			EntityId:  id,
			LocationX: int32(id),
			LocationY: int32(id),
			Timestamp: nextActionTimestamp,
		})
		_, _ = gameEnemyStateHandle.Insert(GameEnemyState{
			EntityId: id,
			HerdId:   int32(id),
		})
		_, _ = gameHerdCacheHandle.Insert(GameHerdCache{
			Id:                int32(id),
			DimensionId:       uint32(id),
			CurrentPopulation: int32(id) * 2,
			MaxPopulation:     int32(id) * 4,
			SpawnEagerness:    float32(id),
			RoamingDistance:   int32(id),
			Location: SmallHexTile{
				X:         int32(id),
				Z:         int32(id),
				Dimension: uint32(id) * 2,
			},
		})
	}
	spacetimedb.LogInfo(fmt.Sprintf("INSERT WORLD PLAYERS: %d", players))
}

func insertWorldReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	players, _ := r.ReadU64()
	insertWorld(players)
}

func gameLoopEnemyIa(players uint64) {
	count := 0
	currentTimeMs := momentMilliseconds()

	for agent, err := range gameEnemyAiAgentStateHandle.Iter() {
		if err != nil {
			break
		}
		agentTargetable, err2 := gameTargetableStateEntityIdIdx.Find(agent.EntityId)
		if err2 != nil || agentTargetable == nil {
			spacetimedb.LogPanic("No TargetableState for AgentState entity")
		}

		// Compute nearby targets as benchmark workload (result not used further).
		_ = getTargetablesNearQuad(agentTargetable.EntityId, players)

		agent.Action = AgentActionFighting
		agentLoop(agent, currentTimeMs)
		count++
	}
	spacetimedb.LogInfo(fmt.Sprintf("ENEMY IA LOOP PLAYERS: %d, processed: %d", players, count))
}

func gameLoopEnemyIaReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	players, _ := r.ReadU64()
	gameLoopEnemyIa(players)
}

func initGameIaLoopReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	initialLoad, _ := r.ReadU32()
	l := newLoad(initialLoad)

	insertBulkPosition(l.biggestTable)
	insertBulkVelocity(l.bigTable)
	updatePositionAll(l.biggestTable)
	updatePositionWithVelocity(l.bigTable)
	insertWorld(uint64(l.numPlayers))
}

func runGameIaLoopReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	initialLoad, _ := r.ReadU32()
	l := newLoad(initialLoad)

	gameLoopEnemyIa(uint64(l.numPlayers))
}
