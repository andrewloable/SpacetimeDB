use crate::detect::find_executable;
use std::path::{Path, PathBuf};

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
