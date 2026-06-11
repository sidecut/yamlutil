use std::fs;
use std::io::Write;

use anyhow::bail;
use serde_yaml::Value;

use crate::argsprocessor::{self, STDIN};

pub fn run(files: &[String], replace: bool, auto: bool) -> anyhow::Result<()> {
    if files.is_empty() && (auto || replace) {
        bail!("Can't use --auto or --replace with stdin");
    } else if !files.is_empty() && auto && replace {
        bail!("Can't use both --auto and --replace");
    }

    argsprocessor::process_args(files, |filename, contents| {
        if files.len() > 1 && filename != STDIN {
            println!("{filename}:");
        }

        let sorted = sort_value(argsprocessor::first_document(contents)?);
        let output = serde_yaml::to_string(&sorted)?;

        if filename == STDIN {
            print!("{output}");
        } else {
            let output_filename = get_output_filename(filename, replace, auto)?;
            if output_filename == STDIN {
                print!("{output}");
            } else {
                let mut file = fs::File::create(&output_filename)?;
                file.write_all(output.as_bytes())?;
            }
        }

        Ok(())
    })
}

/// Recursively sorts mapping keys alphabetically.
fn sort_value(value: Value) -> Value {
    match value {
        Value::Mapping(map) => {
            let mut entries: Vec<(Value, Value)> =
                map.into_iter().map(|(k, v)| (k, sort_value(v))).collect();
            entries.sort_by_key(|(k, _)| key_sort_string(k));
            Value::Mapping(entries.into_iter().collect())
        }
        Value::Sequence(seq) => Value::Sequence(seq.into_iter().map(sort_value).collect()),
        other => other,
    }
}

fn key_sort_string(key: &Value) -> String {
    match key {
        Value::Null => "null".to_string(),
        Value::Bool(b) => b.to_string(),
        Value::Number(n) => n.to_string(),
        Value::String(s) => s.clone(),
        other => serde_yaml::to_string(other)
            .unwrap_or_default()
            .trim_end()
            .to_string(),
    }
}

fn get_output_filename(filename: &str, replace: bool, auto: bool) -> anyhow::Result<String> {
    if replace {
        return Ok(filename.to_string());
    }
    if !auto {
        return Ok(STDIN.to_string());
    }

    const OUT_YAML: &str = ".sorted.yaml";
    if filename.is_empty() {
        bail!("Empty filename");
    }

    let parts: Vec<&str> = filename.split('.').collect();
    if parts.len() == 1 {
        return Ok(format!("{filename}{OUT_YAML}"));
    }

    let extension = parts[parts.len() - 1];
    if extension.eq_ignore_ascii_case("yaml") {
        Ok(format!("{}{OUT_YAML}", parts[..parts.len() - 1].join(".")))
    } else {
        Ok(format!("{filename}{OUT_YAML}"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn output_filename_replace() {
        assert_eq!(
            get_output_filename("foo.yaml", true, false).unwrap(),
            "foo.yaml"
        );
    }

    #[test]
    fn output_filename_stdout() {
        assert_eq!(get_output_filename("foo.yaml", false, false).unwrap(), "-");
    }

    #[test]
    fn output_filename_auto() {
        assert_eq!(
            get_output_filename("foo.yaml", false, true).unwrap(),
            "foo.sorted.yaml"
        );
        assert_eq!(
            get_output_filename("foo.YAML", false, true).unwrap(),
            "foo.sorted.yaml"
        );
        assert_eq!(
            get_output_filename("foo", false, true).unwrap(),
            "foo.sorted.yaml"
        );
        assert_eq!(
            get_output_filename("foo.txt", false, true).unwrap(),
            "foo.txt.sorted.yaml"
        );
        assert_eq!(
            get_output_filename("a.b.yaml", false, true).unwrap(),
            "a.b.sorted.yaml"
        );
        assert!(get_output_filename("", false, true).is_err());
    }

    #[test]
    fn sorts_keys_recursively() {
        let value: Value = serde_yaml::from_str("b:\n  z: 1\n  a: 2\na: 3\n").unwrap();
        let sorted = sort_value(value);
        let out = serde_yaml::to_string(&sorted).unwrap();
        assert_eq!(out, "a: 3\nb:\n  a: 2\n  z: 1\n");
    }
}
