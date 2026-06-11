//! File-or-stdin iteration, the Rust port of `argsprocessor.ProcessArgs`.
//!
//! The Go version used a sentinel constant `STDIN = "-"` and a callback
//! (`eachFile`) invoked once per file (or once for stdin when no args were
//! supplied). We model the same behavior with a `Source` enum and a
//! [`process_args`] function that takes a closure with the same signature.

use std::fs::File;
use std::io::{self, Read};

/// Sentinel value that mirrors the Go `argsprocessor.STDIN` constant. The
/// `list` and `sort` commands both use this to detect when they are reading
/// from stdin so they can skip the per-file `filename:` header.
pub const STDIN: &str = "-";

/// A single source of YAML input, opened and ready to read.
pub struct Source {
    /// `"-"` for stdin, otherwise the path on disk.
    pub label: String,
    reader: InputReader,
}

enum InputReader {
    Stdin(io::Stdin),
    File(File),
}

impl Read for Source {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        match &mut self.reader {
            InputReader::Stdin(s) => s.read(buf),
            InputReader::File(f) => f.read(buf),
        }
    }
}

/// Port of `argsprocessor.ProcessArgs`.
///
/// If `args` is empty, yields a single stdin source labeled [`STDIN`].
/// Otherwise opens each file in `args` and yields it in order. The first
/// open error is returned; the callback is short-circuited at that point.
pub fn process_args<F>(args: &[String], stdin: io::Stdin, mut each: F) -> anyhow::Result<()>
where
    F: FnMut(Source) -> anyhow::Result<()>,
{
    if args.is_empty() {
        each(Source {
            label: STDIN.to_string(),
            reader: InputReader::Stdin(stdin),
        })?;
        return Ok(());
    }

    for filename in args {
        let f = File::open(filename).map_err(|e| anyhow::anyhow!("opening {filename}: {e}"))?;
        each(Source {
            label: filename.clone(),
            reader: InputReader::File(f),
        })?;
    }
    Ok(())
}
