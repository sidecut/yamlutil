//! Root command and CLI entry point.
//!
//! Spirit of the Viper wiring in the original `cmd/root.go` is preserved:
//! a `--config` flag overrides the path, otherwise we look for
//! `$HOME/.yamlutil.yaml` and read it if present. We do not use Viper or
//! any equivalent here because the config file is informational only —
//! no subcommand reads from it today — so a hand-rolled YAML probe is
//! the minimum that preserves the documented behavior.

use std::path::PathBuf;

use clap::{Parser, Subcommand};

mod args;
mod list;
mod sort;

#[derive(Parser, Debug)]
#[command(
    name = "yamlutil",
    version,
    about = "YAML utility",
    long_about = None
)]
struct Cli {
    /// Path to a config file. Defaults to `$HOME/.yamlutil.yaml`.
    #[arg(long, global = true)]
    config: Option<PathBuf>,

    #[command(subcommand)]
    command: Command,
}

#[derive(Subcommand, Debug)]
enum Command {
    /// List all keys in the file (recursively, with dotted/indexed paths).
    List(list::ListArgs),
    /// Sort YAML keys.
    Sort(sort::SortArgs),
}

fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();

    // Best-effort config probe — same behavior as the Go init(): print
    // "Using config file: …" if one is found, otherwise silently continue.
    if let Some(path) = resolve_config_path(cli.config.as_deref())
        && path.exists()
    {
        println!("Using config file: {}", path.display());
    }

    match cli.command {
        Command::List(args) => list::run(args),
        Command::Sort(args) => sort::run(args),
    }
}

fn resolve_config_path(flag: Option<&std::path::Path>) -> Option<PathBuf> {
    if let Some(p) = flag {
        return Some(p.to_path_buf());
    }
    dirs::home_dir().map(|h| h.join(".yamlutil.yaml"))
}
