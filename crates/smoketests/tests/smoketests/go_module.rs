#![allow(clippy::disallowed_macros)]
use spacetimedb_guard::ensure_binaries_built;
use spacetimedb_smoketests::{require_tinygo, workspace_root};
use std::process::Command;

/// Ensure that the CLI is able to create and compile a Go project using TinyGo.
/// This test does not depend on a running SpacetimeDB instance.
/// Skips if tinygo is not available in PATH.
#[test]
fn test_build_go_module() {
    require_tinygo!();

    let workspace = workspace_root();
    let cli_path = ensure_binaries_built();

    // Create temp directory for the project.
    let tmpdir = tempfile::tempdir().expect("Failed to create temp directory");

    // Initialize Go project via spacetime init.
    let output = Command::new(&cli_path)
        .args([
            "init",
            "--non-interactive",
            "--lang=go",
            "--project-path",
            tmpdir.path().to_str().unwrap(),
            "go-project",
        ])
        .output()
        .expect("Failed to run spacetime init");
    assert!(
        output.status.success(),
        "spacetime init --lang=go failed:\nstdout: {}\nstderr: {}",
        String::from_utf8_lossy(&output.stdout),
        String::from_utf8_lossy(&output.stderr)
    );

    let server_path = tmpdir.path().join("spacetimedb");
    assert!(server_path.exists(), "spacetimedb/ directory was not created");
    assert!(server_path.join("go.mod").exists(), "go.mod was not created");
    assert!(server_path.join("main.go").exists(), "main.go was not created");

    // Add a replace directive so TinyGo can find the local server SDK.
    // In a real user scenario the module would be fetched from the registry.
    let go_mod_path = server_path.join("go.mod");
    let go_mod = std::fs::read_to_string(&go_mod_path).expect("Failed to read go.mod");
    let sdk_path = workspace.join("crates/bindings-go");
    let client_sdk_path = workspace.join("sdks/go");
    let updated = format!(
        "{}\nreplace github.com/clockworklabs/spacetimedb-go-server => {}\nreplace github.com/clockworklabs/spacetimedb-go => {}\n",
        go_mod,
        sdk_path.display(),
        client_sdk_path.display()
    );
    std::fs::write(&go_mod_path, updated).expect("Failed to write go.mod");

    // Compile the module to WASM using TinyGo.
    let output = Command::new("tinygo")
        .args([
            "build",
            "-target",
            "wasm",
            "-o",
            server_path.join("module.wasm").to_str().unwrap(),
            "./",
        ])
        .current_dir(&server_path)
        .output()
        .expect("Failed to run tinygo build");
    assert!(
        output.status.success(),
        "tinygo build failed:\nstdout: {}\nstderr: {}",
        String::from_utf8_lossy(&output.stdout),
        String::from_utf8_lossy(&output.stderr)
    );

    let wasm_path = server_path.join("module.wasm");
    assert!(wasm_path.exists(), "module.wasm was not produced");
    assert!(
        wasm_path.metadata().unwrap().len() > 0,
        "module.wasm is empty"
    );

    eprintln!(
        "Go module compiled successfully: {} bytes",
        wasm_path.metadata().unwrap().len()
    );
}
