use crate::detect::find_executable;
use std::fs;
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

    // Build using wasm-unknown target with c-shared buildmode (reactor model).
    // This produces a WASM module that:
    //   - exports _initialize instead of _start (no proc_exit call)
    //   - has zero non-SpacetimeDB imports
    duct::cmd!(
        "tinygo",
        "build",
        "-target",
        "wasm-unknown",
        "-buildmode",
        "c-shared",
        "-o",
        &output_path,
        "./"
    )
    .dir(project_path)
    .run()?;

    // Post-process: add a __preinit__10_go_init export that aliases _initialize.
    // SpacetimeDB calls all __preinit__* exports (sorted) before __describe_module__.
    // This ensures Go's init() functions run and populate the module registries
    // before the host reads the schema.
    let wasm_bytes = fs::read(&output_path)?;
    let patched = add_preinit_export(&wasm_bytes, "__preinit__10_go_init")
        .ok_or_else(|| anyhow::anyhow!("Failed to add __preinit__10_go_init export to WASM: _initialize not found"))?;
    fs::write(&output_path, patched)?;

    Ok(output_path)
}

/// Adds a `__preinit__10_go_init` export to the WASM binary that aliases the
/// `_initialize` function. Returns None if `_initialize` is not found.
fn add_preinit_export(wasm: &[u8], preinit_name: &str) -> Option<Vec<u8>> {
    if wasm.len() < 8 {
        return None;
    }

    // Find _initialize's function index in the export section.
    let initialize_func_idx = find_export_func_idx(wasm, "_initialize")?;

    // Encode the new export entry: name_len + name_bytes + kind(0=func) + func_idx_leb
    let name_bytes = preinit_name.as_bytes();
    let mut new_export = Vec::new();
    leb_encode(&mut new_export, name_bytes.len() as u64);
    new_export.extend_from_slice(name_bytes);
    new_export.push(0); // kind = func
    leb_encode(&mut new_export, initialize_func_idx as u64);

    // Inject the new export into the export section.
    inject_export(wasm, new_export)
}

/// Find the function index of an exported function by name.
fn find_export_func_idx(wasm: &[u8], name: &str) -> Option<u32> {
    let mut pos = 8usize;
    while pos < wasm.len() {
        let sid = wasm[pos];
        pos += 1;
        let (sz, after) = leb_decode(wasm, pos)?;
        pos = after;
        let section_end = pos + sz as usize;

        if sid == 7 {
            // Export section
            let (count, mut epos) = leb_decode(wasm, pos)?;
            for _ in 0..count {
                let (nlen, after_nlen) = leb_decode(wasm, epos)?;
                epos = after_nlen;
                let export_name = std::str::from_utf8(&wasm[epos..epos + nlen as usize]).ok()?;
                epos += nlen as usize;
                let kind = wasm[epos];
                epos += 1;
                let (idx, after_idx) = leb_decode(wasm, epos)?;
                epos = after_idx;
                if export_name == name && kind == 0 {
                    return Some(idx as u32);
                }
            }
        }

        pos = section_end;
    }
    None
}

/// Inject a pre-encoded export entry into the export section, incrementing the count.
fn inject_export(wasm: &[u8], new_export: Vec<u8>) -> Option<Vec<u8>> {
    let mut pos = 8usize;
    while pos < wasm.len() {
        let sid = wasm[pos];
        let sid_pos = pos;
        pos += 1;
        let (sz, after) = leb_decode(wasm, pos)?;
        pos = after;
        let section_end = pos + sz as usize;

        if sid == 7 {
            // Export section found — read current count
            let (count, after_count) = leb_decode(wasm, pos)?;
            let new_count = count + 1;

            let new_count_encoded = {
                let mut v = Vec::new();
                leb_encode(&mut v, new_count as u64);
                v
            };

            // Build new section content: new_count + rest_of_exports + new_export
            let rest = &wasm[after_count..section_end];
            let mut new_content = new_count_encoded;
            new_content.extend_from_slice(rest);
            new_content.extend_from_slice(&new_export);

            let mut new_size_encoded = Vec::new();
            leb_encode(&mut new_size_encoded, new_content.len() as u64);

            // Assemble: [before section] [sid] [new_size] [new_content] [after section]
            let mut result = Vec::with_capacity(wasm.len() + new_export.len() + 4);
            result.extend_from_slice(&wasm[..sid_pos]);
            result.push(7);
            result.extend_from_slice(&new_size_encoded);
            result.extend_from_slice(&new_content);
            result.extend_from_slice(&wasm[section_end..]);
            return Some(result);
        }

        pos = section_end;
    }
    None
}

fn leb_decode(data: &[u8], mut pos: usize) -> Option<(u64, usize)> {
    let mut v = 0u64;
    let mut s = 0u32;
    loop {
        let b = *data.get(pos)? as u64;
        pos += 1;
        v |= (b & 0x7f) << s;
        s += 7;
        if (b & 0x80) == 0 {
            break;
        }
    }
    Some((v, pos))
}

fn leb_encode(out: &mut Vec<u8>, mut n: u64) {
    loop {
        let b = (n & 0x7f) as u8;
        n >>= 7;
        if n != 0 {
            out.push(b | 0x80);
        } else {
            out.push(b);
            break;
        }
    }
}

