package main

import (
	"fmt"
	"math"

	spacetimedb "github.com/clockworklabs/spacetimedb-go-server"
	"github.com/clockworklabs/spacetimedb-go-server/sys"
	"github.com/clockworklabs/spacetimedb-go/bsatn"
)

func massToRadius(mass uint32) float32 {
	return float32(math.Sqrt(float64(mass)))
}

func isOverlapping(e1, e2 Entity) bool {
	r1 := massToRadius(e1.Mass)
	r2 := massToRadius(e2.Mass)
	dx := e1.Position.X - e2.Position.X
	dy := e1.Position.Y - e2.Position.Y
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	maxR := r1
	if r2 > maxR {
		maxR = r2
	}
	return dist < maxR
}

func insertBulkEntity(count uint32) {
	for id := uint32(0); id < count; id++ {
		e := Entity{
			Id:       0, // auto_inc
			Position: Vector2{X: float32(id), Y: float32(id + 5)},
			Mass:     id * 5,
		}
		if _, err := entityHandle.Insert(e); err != nil {
			spacetimedb.LogPanic("insert_bulk_entity: " + err.Error())
		}
	}
	spacetimedb.LogInfo(fmt.Sprintf("INSERT ENTITY: %d", count))
}

func insertBulkEntityReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	count, _ := r.ReadU32()
	insertBulkEntity(count)
}

func insertBulkCircle(ctx spacetimedb.ReducerContext, count uint32) {
	for id := uint32(0); id < count; id++ {
		c := Circle{
			EntityId:      id,
			PlayerId:      id,
			Direction:     Vector2{X: float32(id), Y: float32(id + 5)},
			Magnitude:     float32(id * 5),
			LastSplitTime: ctx.Timestamp,
		}
		if _, err := circleHandle.Insert(c); err != nil {
			spacetimedb.LogPanic("insert_bulk_circle: " + err.Error())
		}
	}
	spacetimedb.LogInfo(fmt.Sprintf("INSERT CIRCLE: %d", count))
}

func insertBulkCircleReducer(ctx spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	count, _ := r.ReadU32()
	insertBulkCircle(ctx, count)
}

func insertBulkFood(count uint32) {
	for id := uint32(1); id <= count; id++ {
		if _, err := foodHandle.Insert(Food{EntityId: id}); err != nil {
			spacetimedb.LogPanic("insert_bulk_food: " + err.Error())
		}
	}
	spacetimedb.LogInfo(fmt.Sprintf("INSERT FOOD: %d", count))
}

func insertBulkFoodReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	count, _ := r.ReadU32()
	insertBulkFood(count)
}

func crossJoinAll(expected uint32) {
	count := 0
	for _, err1 := range circleHandle.Iter() {
		if err1 != nil {
			break
		}
		for _, err2 := range entityHandle.Iter() {
			if err2 != nil {
				break
			}
			for _, err3 := range foodHandle.Iter() {
				if err3 != nil {
					break
				}
				count++
			}
		}
	}
	spacetimedb.LogInfo(fmt.Sprintf("CROSS JOIN ALL: %d, processed: %d", expected, count))
}

func crossJoinAllReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	expected, _ := r.ReadU32()
	crossJoinAll(expected)
}

func crossJoinCircleFood(expected uint32) {
	count := 0
	for circle, err := range circleHandle.Iter() {
		if err != nil {
			break
		}
		circleEntity, err2 := entityIdIdx.Find(circle.EntityId)
		if err2 != nil || circleEntity == nil {
			continue
		}
		for food, err3 := range foodHandle.Iter() {
			if err3 != nil {
				break
			}
			foodEntity, err4 := entityIdIdx.Find(food.EntityId)
			if err4 != nil || foodEntity == nil {
				spacetimedb.LogPanic(fmt.Sprintf("Entity not found: %d", food.EntityId))
			}
			count++
			_ = isOverlapping(*circleEntity, *foodEntity)
		}
	}
	spacetimedb.LogInfo(fmt.Sprintf("CROSS JOIN CIRCLE FOOD: %d, processed: %d", expected, count))
}

func crossJoinCircleFoodReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	expected, _ := r.ReadU32()
	crossJoinCircleFood(expected)
}

func initGameCirclesReducer(ctx spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	initialLoad, _ := r.ReadU32()
	l := newLoad(initialLoad)

	insertBulkFood(l.initialLoad)
	insertBulkEntity(l.initialLoad)
	insertBulkCircle(ctx, l.smallTable)
}

func runGameCirclesReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, _ := sys.ReadBytesSource(args)
	r := bsatn.NewReader(data)
	initialLoad, _ := r.ReadU32()
	l := newLoad(initialLoad)

	crossJoinCircleFood(l.initialLoad * l.smallTable)
	crossJoinAll(l.initialLoad * l.initialLoad * l.smallTable)
}
