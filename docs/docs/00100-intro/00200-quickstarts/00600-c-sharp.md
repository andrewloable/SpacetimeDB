---
title: C# Quickstart
sidebar_label: C#
slug: /quickstarts/c-sharp
hide_table_of_contents: true
---

import { InstallCardLink } from "@site/src/components/InstallCardLink";
import { StepByStep, Step, StepText, StepCode } from "@site/src/components/Steps";


Get a SpacetimeDB C# app running in under 5 minutes.

## Prerequisites

- [.NET 10 SDK](https://dotnet.microsoft.com/download/dotnet/10.0) installed
- [SpacetimeDB CLI](https://spacetimedb.com/install) installed
- [wasi-sdk](https://github.com/WebAssembly/wasi-sdk/releases) v25+ — required by .NET 10 to compile C# modules to WebAssembly. SpacetimeDB auto-downloads it on first build to `~/.wasi-sdk/`, but if that fails, install it manually (see step 1).

<InstallCardLink />

---

<StepByStep>
  <Step title="Install wasi-sdk (if auto-download fails)">
    <StepText>
      .NET 10 uses wasi-sdk (replacing the old `wasi-experimental` workload from .NET 8) to compile C# modules to WebAssembly. SpacetimeDB auto-downloads it on first build, but if that fails you can install it manually:
    </StepText>
    <StepCode>
```bash
# macOS (ARM64)
curl -LO https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-25/wasi-sdk-25.0-arm64-macos.tar.gz
tar xf wasi-sdk-25.0-arm64-macos.tar.gz
mkdir -p ~/.wasi-sdk && mv wasi-sdk-25.0-arm64-macos ~/.wasi-sdk/wasi-sdk-25

# macOS (x64): use wasi-sdk-25.0-x86_64-macos.tar.gz
# Linux (x64): use wasi-sdk-25.0-x86_64-linux.tar.gz
# Windows (x64): use wasi-sdk-25.0-x86_64-windows.tar.gz

# Set the environment variable (add to your shell profile)
export WASI_SDK_PATH="$HOME/.wasi-sdk/wasi-sdk-25"
```
    </StepCode>
  </Step>

  <Step title="Create your project">
    <StepText>
      Run the `spacetime dev` command to create a new project with a C# SpacetimeDB module.

      This will start the local SpacetimeDB server, compile and publish your module, and generate C# client bindings.
    </StepText>
    <StepCode>
```bash
spacetime dev --template basic-cs
```
    </StepCode>
  </Step>

  <Step title="Explore the project structure">
    <StepText>
      Your project contains both server and client code.

      Edit `spacetimedb/Lib.cs` to add tables and reducers. Use the generated bindings in the client project.
    </StepText>
    <StepCode>
```
my-spacetime-app/
├── spacetimedb/             # Your SpacetimeDB module
│   ├── StdbModule.csproj
│   └── Lib.cs               # Server-side logic
├── client.csproj
├── Program.cs               # Client application
└── module_bindings/         # Auto-generated types
```
    </StepCode>
  </Step>

  <Step title="Understand tables and reducers">
    <StepText>
      Open `spacetimedb/Lib.cs` to see the module code. The template includes a `Person` table and two reducers: `Add` to insert a person, and `SayHello` to greet everyone.

      Tables store your data. Reducers are functions that modify data — they're the only way to write to the database.
    </StepText>
    <StepCode>
```csharp
using SpacetimeDB;

public static partial class Module
{
    [SpacetimeDB.Table(Accessor = "Person", Public = true)]
    public partial struct Person
    {
        public string Name;
    }

    [SpacetimeDB.Reducer]
    public static void Add(ReducerContext ctx, string name)
    {
        ctx.Db.Person.Insert(new Person { Name = name });
    }

    [SpacetimeDB.Reducer]
    public static void SayHello(ReducerContext ctx)
    {
        foreach (var person in ctx.Db.Person.Iter())
        {
            Log.Info($"Hello, {person.Name}!");
        }
        Log.Info("Hello, World!");
    }
}
```
    </StepCode>
  </Step>

  <Step title="Test with the CLI">
    <StepText>
      Open a new terminal and navigate to your project directory. Then use the SpacetimeDB CLI to call reducers and query your data directly.
    </StepText>
    <StepCode>
```bash
cd my-spacetime-app

# Call the add reducer to insert a person
spacetime call add Alice

# Query the person table
spacetime sql "SELECT * FROM Person"
 name
---------
 "Alice"

# Call say_hello to greet everyone
spacetime call say_hello

# View the module logs
spacetime logs
2025-01-13T12:00:00.000000Z  INFO: Hello, Alice!
2025-01-13T12:00:00.000000Z  INFO: Hello, World!
```
    </StepCode>
  </Step>
</StepByStep>

## Next steps

- See the [Chat App Tutorial](../00300-tutorials/00100-chat-app.md) for a complete example
- Read the [C# SDK Reference](../../00200-core-concepts/00600-clients/00600-csharp-reference.md) for detailed API docs
