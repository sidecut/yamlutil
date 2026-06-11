use serde_yaml::{Mapping, Value};

use crate::argsprocessor::{self, STDIN};

pub fn run(files: &[String]) -> anyhow::Result<()> {
    argsprocessor::process_args(files, |filename, contents| {
        if files.len() > 1 && filename != STDIN {
            println!("{filename}:");
        }

        let data = argsprocessor::first_document(contents)?;
        if let Value::Mapping(map) = &data {
            list_keys("", map);
        }

        Ok(())
    })
}

/// Recursively lists all the keys in a mapping using dot notation.
fn list_keys(prefix: &str, map: &Mapping) {
    for (key, value) in map {
        let key = full_key(prefix, key);
        println!("{key}");
        match value {
            Value::Sequence(seq) => list_array(&key, seq),
            Value::Mapping(map) => list_keys(&key, map),
            _ => {}
        }
    }
}

fn full_key(prefix: &str, key: &Value) -> String {
    match key {
        Value::String(s) => format!("{prefix}.{s}"),
        other => format!("{prefix}.\"{}\"", scalar_to_string(other)),
    }
}

fn scalar_to_string(value: &Value) -> String {
    match value {
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

fn list_array(key: &str, array: &[Value]) {
    for i in 0..array.len() {
        println!("{key}[{i}]");
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn string_keys_use_dot_notation() {
        assert_eq!(full_key("", &Value::String("key".into())), ".key");
        assert_eq!(
            full_key(".parent", &Value::String("child".into())),
            ".parent.child"
        );
    }

    #[test]
    fn non_string_keys_are_quoted() {
        assert_eq!(full_key("", &Value::Number(0.into())), ".\"0\"");
        assert_eq!(full_key("", &Value::Bool(false)), ".\"false\"");
        assert_eq!(full_key("", &Value::Null), ".\"null\"");
    }
}
