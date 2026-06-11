use std::fs;
use std::io::Read;

pub const STDIN: &str = "-";

/// Reads either stdin or a series of files named in `args`, invoking
/// `each_file` with the filename and file contents.
pub fn process_args<F>(args: &[String], mut each_file: F) -> anyhow::Result<()>
where
    F: FnMut(&str, &str) -> anyhow::Result<()>,
{
    if args.is_empty() {
        let mut contents = String::new();
        std::io::stdin().read_to_string(&mut contents)?;
        each_file(STDIN, &contents)?;
    } else {
        for filename in args {
            let contents = fs::read_to_string(filename)?;
            each_file(filename, &contents)?;
        }
    }

    Ok(())
}

/// Parses the first YAML document in `contents`, mirroring the behavior of
/// Go's yaml.Unmarshal on multi-document input.
pub fn first_document(contents: &str) -> anyhow::Result<serde_yaml::Value> {
    use serde::Deserialize;

    if let Some(doc) = serde_yaml::Deserializer::from_str(contents).next() {
        Ok(serde_yaml::Value::deserialize(doc)?)
    } else {
        Ok(serde_yaml::Value::Null)
    }
}
