package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/clockworklabs/spacetimedb-go/client"
	"spacetimedb-client/module_bindings"
)

func main() {
	host := "localhost:3000"
	dbName := "my-db"
	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	if len(os.Args) > 2 {
		dbName = os.Args[2]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := client.NewDbConnectionBuilder().
		WithUri(host).
		WithModuleName(dbName).
		Build(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	tables := module_bindings.NewRemoteTables()
	reducers := module_bindings.NewRemoteReducers(conn)
	module_bindings.RegisterTables(conn, tables)

	// Register a callback for when rows are inserted into the person table.
	tables.Person.OnInsert(func(person module_bindings.Person, _ client.EventContext) {
		fmt.Printf("New person: %s\n", person.Name)
	})

	errCh := conn.RunAsync(ctx)

	// Subscribe to the person table.
	if _, err := conn.Subscribe([]string{"SELECT * FROM Person"}); err != nil {
		fmt.Fprintf(os.Stderr, "subscribe: %v\n", err)
		os.Exit(1)
	}

	// Allow subscription to be applied.
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("Persons in cache: %d\n", tables.Person.Count())

	conn.Disconnect()

	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			fmt.Fprintf(os.Stderr, "connection error: %v\n", err)
			os.Exit(1)
		}
	case <-time.After(2 * time.Second):
	}

	fmt.Println("Done.")
}
