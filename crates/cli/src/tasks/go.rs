use crate::detect::find_executable;
use anyhow::Context;
use std::path::{Path, PathBuf};

mod embedded_go_sdk {
    include!(concat!(env!("OUT_DIR"), "/embedded_go_sdk.rs"));
}

pub(crate) fn build_go(project_path: &Path, _build_debug: bool) -> anyhow::Result<PathBuf> {
    // Verify that TinyGo is installed.
    let tinygo = if cfg!(windows) {
        find_executable("tinygo.exe")
    } else {
        find_executable("tinygo")
    };
    if tinygo.is_none() {
        anyhow::bail!(
            "tinygo not found in PATH.\n\
             Please install TinyGo from https://tinygo.org/getting-started/install/ \
             and ensure it is on your PATH."
        );
    }

    // Check if go.mod references spacetimedb-go-server and needs replace directives.
    let go_mod_path = project_path.join("go.mod");
    let go_mod_content = std::fs::read_to_string(&go_mod_path).context("Failed to read go.mod")?;

    let needs_server_sdk = go_mod_content.contains("spacetimedb-go-server")
        && !go_mod_content.contains("replace github.com/clockworklabs/spacetimedb-go-server");
    let needs_client_sdk = go_mod_content.contains("spacetimedb-go")
        && !go_mod_content.contains("replace github.com/clockworklabs/spacetimedb-go ");

    if needs_server_sdk || needs_client_sdk {
        // Extract embedded Go SDK files to a cache directory.
        let cache_dir = dirs::cache_dir()
            .unwrap_or_else(|| PathBuf::from("/tmp"))
            .join("spacetimedb")
            .join("go-sdk");

        extract_embedded_go_sdk(&cache_dir)?;

        // Add replace directives to go.mod.
        if needs_server_sdk {
            let server_sdk_path = cache_dir.join("bindings-go");
            duct::cmd!(
                "go",
                "mod",
                "edit",
                "-replace",
                format!(
                    "github.com/clockworklabs/spacetimedb-go-server={}",
                    server_sdk_path.display()
                )
            )
            .dir(project_path)
            .run()
            .context("Failed to add replace directive for spacetimedb-go-server")?;
        }

        if needs_client_sdk {
            let client_sdk_path = cache_dir.join("sdks-go");
            duct::cmd!(
                "go",
                "mod",
                "edit",
                "-replace",
                format!(
                    "github.com/clockworklabs/spacetimedb-go={}",
                    client_sdk_path.display()
                )
            )
            .dir(project_path)
            .run()
            .context("Failed to add replace directive for spacetimedb-go")?;
        }

        // Run go mod tidy to generate go.sum.
        duct::cmd!("go", "mod", "tidy")
            .dir(project_path)
            .env("GONOSUMCHECK", "*")
            .env("GONOSUMDB", "*")
            .run()
            .context("Failed to run go mod tidy")?;
    }

    let output_path = project_path.join("module.wasm");

    duct::cmd!(
        "tinygo",
        "build",
        "-target",
        "wasm",
        "-o",
        &output_path,
        "./"
    )
    .dir(project_path)
    .run()?;

    Ok(output_path)
}

fn extract_embedded_go_sdk(cache_dir: &Path) -> anyhow::Result<()> {
    let server_sdk_dir = cache_dir.join("bindings-go");
    let client_sdk_dir = cache_dir.join("sdks-go");

    // Extract server SDK (crates/bindings-go).
    for (relative_path, content) in embedded_go_sdk::server_sdk_files() {
        let dest = server_sdk_dir.join(relative_path);
        if let Some(parent) = dest.parent() {
            std::fs::create_dir_all(parent)?;
        }
        write_if_changed(&dest, content.as_bytes())?;
    }

    // Extract client SDK (sdks/go).
    for (relative_path, content) in embedded_go_sdk::client_sdk_files() {
        let dest = client_sdk_dir.join(relative_path);
        if let Some(parent) = dest.parent() {
            std::fs::create_dir_all(parent)?;
        }
        write_if_changed(&dest, content.as_bytes())?;
    }

    Ok(())
}

fn write_if_changed(path: &Path, contents: &[u8]) -> std::io::Result<()> {
    match std::fs::read(path) {
        Ok(existing) if existing == contents => Ok(()),
        _ => std::fs::write(path, contents),
    }
}
