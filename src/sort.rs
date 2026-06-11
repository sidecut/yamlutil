//! Port of `cmd/sort.go`.
//!
//! Sorts the **top-level** keys of a YAML document alphabetically and
//! writes the result. The Go implementation relied on `yaml.v2`'s map
//! emission which is non-deterministic; the Rust port uses `IndexMap`
//! to guarantee lexical order, which is the user-visible contract
//! callers depend on.

use std::fs;
use std::io::{self, Read, Write};
use std::path::PathBuf;

use anyhow::{Context, Result, bail};
use clap::Args;
use indexmap::IndexMap;
use serde_yml::Value;

use crate::args::{self, STDIN};

#[derive(Args, Debug)]
pub struct SortArgs {
    /// Files to sort. If none, read from stdin.
    #[arg(required = false)]
    pub files: Vec<String>,

    /// In-place sort: overwrite the input file(s).
    #[arg(short, long)]
    pub replace: bool,

    /// Write to `*.sorted.yaml` alongside the input. If neither
    /// `--replace` nor `--auto` is set, the sorted YAML is written
    /// to stdout.
    #[arg(short, long)]
    pub auto: bool,
}

pub fn run(args: SortArgs) -> Result<()> {
    // Mirror the Go-side argument validation: --auto and --replace
    // require at least one filename, and the two are mutually
    // exclusive.
    if args.files.is_empty() && (args.replace || args.auto) {
        bail!("--auto/--replace cannot be used with stdin");
    }
    if !args.files.is_empty() && args.replace && args.auto {
        bail!("--auto and --replace are mutually exclusive");
    }

    let stdin = io::stdin();
    let mut stdout = io::stdout().lock();

    args::process_args(&args.files, stdin, |source| {
        // Per-file header: only emit the `filename:` prefix when more
        // than one file was supplied AND this isn't stdin. This matches
        // the behavior of `sort.go` and the `list` subcommand.
        let total = args.files.len();
        let label = source.label.clone();
        let is_stdin = label == STDIN;
        let mut source = source;
        if total > 1 && !is_stdin {
            writeln!(stdout, "{label}:")?;
        }

        if is_stdin {
            do_sort_stdin(&mut source, &mut stdout)?;
        } else {
            let output = get_output_filename(&label, args.replace, args.auto)?;
            do_sort_file(&mut source, &label, &output, &mut stdout)?;
        }
        Ok(())
    })
}

fn do_sort_stdin<R: Read, W: Write>(input: &mut R, output: &mut W) -> Result<()> {
    let mut buf = String::new();
    input.read_to_string(&mut buf)?;
    let sorted = sort_top_level(&buf)?;
    output.write_all(sorted.as_bytes())?;
    Ok(())
}

fn do_sort_file<R: Read, W: Write>(
    input: &mut R,
    input_path: &str,
    output_target: &str,
    stdout: &mut W,
) -> Result<()> {
    let mut buf = String::new();
    input
        .read_to_string(&mut buf)
        .with_context(|| format!("reading {input_path}"))?;
    let sorted = sort_top_level(&buf)?;

    if output_target == STDIN {
        stdout.write_all(sorted.as_bytes())?;
    } else {
        let path = PathBuf::from(output_target);
        fs::write(&path, sorted.as_bytes())
            .with_context(|| format!("writing {}", path.display()))?;
    }
    Ok(())
}

fn get_output_filename(input: &str, replace: bool, auto: bool) -> Result<String> {
    if replace {
        return Ok(input.to_string());
    }
    if !auto {
        return Ok(STDIN.to_string());
    }

    const OUT_SUFFIX: &str = ".sorted.yaml";
    let path = PathBuf::from(input);
    let file_name = path
        .file_name()
        .and_then(|f| f.to_str())
        .ok_or_else(|| anyhow::anyhow!("empty filename"))?;

    if let Some(dot) = file_name.rfind('.') {
        let ext = &file_name[dot + 1..];
        if ext.eq_ignore_ascii_case("yaml") {
            return Ok(format!("{}{}", &file_name[..dot], OUT_SUFFIX));
        }
    }
    Ok(format!("{file_name}{OUT_SUFFIX}"))
}

/// Parses a YAML document, sorts the top-level keys, and re-serializes
/// preserving the original scalar types for those keys. Nested mappings
/// are intentionally left unsorted — this matches the Go behavior.
fn sort_top_level(buf: &str) -> Result<String> {
    // First, parse into a generic `Value` so we preserve the original
    // scalar styles (e.g. `Yes` vs `true`, quoted vs unquoted strings).
    let value: Value = serde_yml::from_str(buf)?;

    let mapping = match value {
        Value::Mapping(m) => m,
        // Pass non-mapping documents through unchanged.
        _ => return Ok(serde_yml::to_string(&value)?),
    };

    // Collect the top-level entries into an IndexMap to guarantee
    // lexical iteration order when we re-serialize. Keys are normalized
    // to strings since noyalib's Mapping.insert only accepts String keys.
    let mut sorted: IndexMap<String, Value> = IndexMap::new();
    for (k, v) in mapping.iter() {
        sorted.insert(k.clone(), v.clone());
    }
    sorted.sort_keys();

    // Rebuild a Mapping now that the keys are guaranteed String.
    let mut out_mapping = serde_yml::Mapping::new();
    for (k, v) in sorted {
        out_mapping.insert(k, v);
    }
    let out = serde_yml::to_string(&Value::Mapping(out_mapping))?;
    Ok(out)
}

#[allow(dead_code)]
fn value_key_to_string(k: &Value) -> String {
    match k {
        Value::String(s) => s.clone(),
        Value::Number(n) => n.to_string(),
        Value::Bool(b) => b.to_string(),
        Value::Null => "null".to_string(),
        _ => format!("{k:?}"),
    }
}
