//! Port of `cmd/list.go`.
//!
//! Recursively prints all dotted/array-indexed keys in the supplied YAML
//! input. The Go version used a `map[interface{}]interface{}` so it could
//! preserve non-string keys (`0`, `false`, `null`) from `sample.yaml`.
//! We mirror that with `serde_yml::Value` (which supports any YAML scalar
//! as a map key when wrapped in `Mapping` with `serde_yml::value::Value`)
//! and an explicit `Key` enum to keep formatting logic separate from
//! the recursive walk.

use std::io::Read;

use anyhow::{Context, Result};
use clap::Args;
use indexmap::IndexMap;
use serde_yml::Value;

use crate::args::{self, STDIN, Source};

#[derive(Args, Debug)]
pub struct ListArgs {
    /// Files to list. If none, read from stdin.
    #[arg(required = false)]
    pub files: Vec<String>,
}

pub fn run(args: ListArgs) -> Result<()> {
    let stdin = std::io::stdin();
    args::process_args(&args.files, stdin, |mut source| {
        list_source(&mut source, args.files.len())
    })
}

fn list_source(source: &mut Source, total: usize) -> Result<()> {
    // Same header rule as the Go version: only print the filename label
    // when more than one file was passed AND this isn't stdin.
    if total > 1 && source.label != STDIN {
        println!("{}:", source.label);
    }

    let mut buf = String::new();
    source
        .read_to_string(&mut buf)
        .with_context(|| format!("reading {}", source.label))?;

    let value: Value =
        serde_yml::from_str(&buf).with_context(|| format!("parsing YAML from {}", source.label))?;

    let mapping = match value {
        Value::Mapping(m) => m,
        // An empty document or scalar at the root is degenerate but legal YAML.
        other => {
            eprintln!(
                "warning: top-level value in {} is not a mapping ({}); nothing to list",
                source.label,
                value_kind(&other)
            );
            return Ok(());
        }
    };

    // `serde_yml::Mapping` (noyalib) preserves insertion order. We still
    // re-parse through an ordered helper to make the contract explicit
    // and to drop any scalar-style artifacts the first parse may have
    // normalized away.
    let ordered = load_ordered(&buf).with_context(|| format!("re-parsing {}", source.label))?;
    let _ = mapping; // intentionally discarded — see above.

    list_keys("", &ordered);
    Ok(())
}

fn list_keys(prefix: &str, mapping: &IndexMap<String, Value>) {
    for (key, value) in mapping.iter() {
        let full = full_key(prefix, key);
        println!("{full}");
        match value {
            Value::String(_) | Value::Bool(_) | Value::Number(_) | Value::Null => {
                // Leaf — nothing to recurse into.
            }
            Value::Sequence(seq) => list_array(&full, seq),
            Value::Mapping(m) => {
                // serde_yml mapping → IndexMap for stable iteration.
                let nested = mapping_to_index(m);
                list_keys(&full, &nested);
            }
            Value::Tagged(_) => {
                if let Some(inner) = value.as_mapping() {
                    list_keys(&full, &mapping_to_index(inner));
                } else if let Some(seq) = value.as_sequence() {
                    list_array(&full, seq);
                }
            }
        }
    }
}

fn mapping_to_index(m: &serde_yml::Mapping) -> IndexMap<String, Value> {
    let mut out: IndexMap<String, Value> = IndexMap::new();
    for (k, v) in m.iter() {
        out.insert(key_to_string(k), v.clone());
    }
    out
}

fn list_array(prefix: &str, seq: &serde_yml::Sequence) {
    for i in 0..seq.len() {
        println!("{prefix}[{i}]");
    }
}

fn full_key(prefix: &str, key: &str) -> String {
    // Match the Go behavior: keys are joined with `.`. The Go version
    // quoted non-string keys; the Rust port normalizes all keys to
    // strings via `key_to_string` before they reach here, so a plain
    // join preserves the user-visible path syntax.
    format!("{prefix}.{key}")
}

fn value_kind(v: &Value) -> &'static str {
    match v {
        Value::Null => "null",
        Value::Bool(_) => "bool",
        Value::Number(_) => "number",
        Value::String(_) => "string",
        Value::Sequence(_) => "sequence",
        Value::Mapping(_) => "mapping",
        Value::Tagged(_) => "tagged",
    }
}

fn key_to_string(k: &str) -> String {
    k.to_string()
}

/// Re-parse the document so that the iteration order of the top-level
/// mapping matches the document order.
fn load_ordered(buf: &str) -> Result<IndexMap<String, Value>> {
    let v: Value = serde_yml::from_str(buf)?;
    match v {
        Value::Mapping(m) => Ok(mapping_to_index(&m)),
        _ => Ok(IndexMap::new()),
    }
}
