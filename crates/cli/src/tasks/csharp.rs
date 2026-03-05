use anyhow::Context;
use itertools::Itertools;
use std::ffi::OsString;
use std::fs;
use std::path::{Path, PathBuf};

fn parse_major_version(version: &str) -> Option<u8> {
    version.split('.').next()?.parse::<u8>().ok()
}

pub(crate) fn build_csharp(project_path: &Path, build_debug: bool) -> anyhow::Result<PathBuf> {
    // All `dotnet` commands must execute in the project directory, otherwise
    // global.json won't have any effect and wrong .NET SDK might be picked.
    macro_rules! dotnet {
        ($($arg:expr),*) => {
            duct::cmd!("dotnet", $($arg),*).dir(project_path)
        };
    }

    // Check .NET SDK version.
    match dotnet!("--version").read() {
        Ok(version) => {
            if parse_major_version(&version) != Some(10) {
                anyhow::bail!(
                    ".NET SDK 10.0 is required, but found {}.\n\
                     If you have multiple versions of .NET SDK installed, configure your project \
                     using https://learn.microsoft.com/en-us/dotnet/core/tools/global-json.",
                    version
                );
            }
        }
        Err(error) if error.kind() == std::io::ErrorKind::NotFound => {
            anyhow::bail!("dotnet not found in PATH. Please install .NET SDK 10.0.")
        }
        Err(error) => anyhow::bail!("{error}"),
    };

    let config_name = if build_debug { "Debug" } else { "Release" };

    // Ensure the project path exists.
    fs::metadata(project_path).with_context(|| {
        format!(
            "The provided project path '{}' does not exist.",
            project_path.to_str().unwrap()
        )
    })?;

    // Resolve WASI_SDK_PATH: use env var if set, otherwise default to ~/.wasi-sdk/wasi-sdk-25
    // (the same path that SpacetimeDB.Runtime.targets auto-downloads to).
    // Must have a trailing slash — .NET's WasiApp.targets concatenates paths without a separator.
    let mut wasi_sdk_path = std::env::var("WASI_SDK_PATH").ok().unwrap_or_else(|| {
        dirs::home_dir()
            .unwrap_or_else(|| PathBuf::from("/tmp"))
            .join(".wasi-sdk")
            .join("wasi-sdk-25")
            .to_string_lossy()
            .into_owned()
    });
    if !wasi_sdk_path.ends_with('/') {
        wasi_sdk_path.push('/');
    }

    // run dotnet publish, passing WASI_SDK_PATH both as env var and MSBuild property
    let wasi_prop = format!("-p:WASI_SDK_PATH={}", wasi_sdk_path);
    duct::cmd!("dotnet", "publish", "-c", config_name, "-v", "quiet", &wasi_prop)
        .dir(project_path)
        .env("WASI_SDK_PATH", &wasi_sdk_path)
        .run()?;

    // check if file exists
    let subdir = if std::env::var_os("EXPERIMENTAL_WASM_AOT").is_some_and(|v| v == "1") {
        "publish"
    } else {
        "AppBundle"
    };
    // TODO: This code looks for build outputs in both `bin` and `bin~` as output directories. @bfops feels like we shouldn't have to look for `bin~`, since the `~` suffix is just intended to cause Unity to ignore directories, and that shouldn't be relevant here. We do think we've seen `bin~` appear though, and it's not harmful to do the extra checks, so we're merging for now due to imminent code freeze. At some point, it would be good to figure out if we do actually see `bin~` in module directories, and where that's coming from (which could suggest a bug).
    // check for the old .NET 7 path for projects that haven't migrated yet
    let bad_output_paths = [
        project_path.join(format!("bin/{config_name}/net7.0/StdbModule.wasm")),
        // for some reason there is sometimes a tilde here?
        project_path.join(format!("bin~/{config_name}/net7.0/StdbModule.wasm")),
    ];
    if bad_output_paths.iter().any(|p| p.exists()) {
        anyhow::bail!(concat!(
            "Looks like your project is using the deprecated .NET 7.0 WebAssembly bindings.\n",
            "Please migrate your project to the new .NET 10.0 template and delete the folders: bin, bin~, obj, obj~"
        ));
    }
    let possible_output_paths = [
        project_path.join(format!("bin/{config_name}/net10.0/wasi-wasm/{subdir}/StdbModule.wasm")),
        project_path.join(format!("bin~/{config_name}/net10.0/wasi-wasm/{subdir}/StdbModule.wasm")),
    ];
    if possible_output_paths.iter().all(|p| p.exists()) {
        anyhow::bail!(concat!(
            "For some reason, your project has both a `bin` and a `bin~` folder.\n",
            "I don't know which to use, so please delete both and rerun this command so that we can see which is up-to-date."
        ));
    }
    for output_path in possible_output_paths {
        if output_path.exists() {
            return Ok(output_path);
        }
    }
    anyhow::bail!("Built project successfully but couldn't find the output file.");
}

pub(crate) fn dotnet_format(project_dir: &Path, files: impl IntoIterator<Item = PathBuf>) -> anyhow::Result<()> {
    let cwd = std::env::current_dir().expect("Failed to retrieve current directory");
    duct::cmd(
        "dotnet",
        itertools::chain(
            [
                "format",
                // We can't guarantee that the output lives inside a valid project or solution,
                // so to avoid crash we need to use the `dotnet whitespace --folder` mode instead
                // of a full style-aware formatter. Still better than nothing though.
                "whitespace",
                "--folder",
                project_dir.to_str().unwrap(),
                // Our files are marked with <auto-generated /> and will be skipped without this option.
                "--include-generated",
                "--include",
            ]
            .into_iter()
            .map_into::<OsString>(),
            // Resolve absolute paths for all of the files, because we receive them as relative paths to cwd, but
            // `dotnet format` will interpret those paths relative to `project_dir`.
            files
                .into_iter()
                .map(|f| {
                    let f = if f.is_absolute() { f } else { cwd.join(f) };
                    f.canonicalize().expect("Failed to canonicalize path: {f}")
                })
                .map_into(),
        ),
    )
    .run()?;
    Ok(())
}
