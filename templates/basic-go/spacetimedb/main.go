//go:generate go run github.com/clockworklabs/spacetimedb-go-server/cmd/stdbgen

package main

import (
	spacetimedb "github.com/clockworklabs/spacetimedb-go-server"
	"github.com/clockworklabs/spacetimedb-go/bsatn"
	"github.com/clockworklabs/spacetimedb-go/types"
	"github.com/clockworklabs/spacetimedb-go-server/sys"
)

// Person is a row in the Person table.
type Person struct {
	Name string
}

func encodePerson(w *bsatn.Writer, p Person) {
	w.WriteString(p.Name)
}

func decodePerson(r *bsatn.Reader) (Person, error) {
	name, err := r.ReadString()
	if err != nil {
		return Person{}, err
	}
	return Person{Name: name}, nil
}

var personTable = spacetimedb.NewTableHandle("Person", encodePerson, decodePerson)

func init() {
	// Register the Person table.
	spacetimedb.RegisterTableDef(spacetimedb.TableDef{
		Name: "Person",
		Columns: []spacetimedb.ColumnDef{
			{Name: "name", Type: types.AlgebraicString},
		},
		Access: spacetimedb.TableAccessPublic,
	})

	// Register the Add reducer (must match handler index below).
	spacetimedb.RegisterReducerDef(spacetimedb.ReducerDef{
		Name: "Add",
		Params: []spacetimedb.ColumnDef{
			{Name: "name", Type: types.AlgebraicString},
		},
		Visibility: spacetimedb.ReducerVisibilityClientCallable,
	})
	spacetimedb.RegisterReducerHandler(addReducer)

	// Register the SayHello reducer.
	spacetimedb.RegisterReducerDef(spacetimedb.ReducerDef{
		Name:       "SayHello",
		Params:     []spacetimedb.ColumnDef{},
		Visibility: spacetimedb.ReducerVisibilityClientCallable,
	})
	spacetimedb.RegisterReducerHandler(sayHelloReducer)
}

func addReducer(_ spacetimedb.ReducerContext, args sys.BytesSource) {
	data, err := sys.ReadBytesSource(args)
	if err != nil {
		spacetimedb.LogError("addReducer: failed to read args: " + err.Error())
		return
	}
	r := bsatn.NewReader(data)
	name, err := r.ReadString()
	if err != nil {
		spacetimedb.LogError("addReducer: failed to decode name: " + err.Error())
		return
	}
	_, err = personTable.Insert(Person{Name: name})
	if err != nil {
		spacetimedb.LogError("addReducer: insert failed: " + err.Error())
	}
}

func sayHelloReducer(_ spacetimedb.ReducerContext, _ sys.BytesSource) {
	for person, err := range personTable.Iter() {
		if err != nil {
			break
		}
		spacetimedb.LogInfo("Hello, " + person.Name + "!")
	}
	spacetimedb.LogInfo("Hello, World!")
}

// main is required by TinyGo but never called by the SpacetimeDB host.
func main() {}
