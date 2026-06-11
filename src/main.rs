mod argsprocessor;
mod list;
mod sort;

use clap::{CommandFactory, Parser, Subcommand};
use clap_complete::Shell;

#[derive(Parser)]
#[command(name = "yamlutil", about = "YAML utility")]
struct Cli {
    /// config file (default is $HOME/.yamlutil.yaml)
    #[arg(long, global = true, value_name = "FILE")]
    config: Option<String>,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// List all keys in the file
    List {
        /// YAML files to list; reads stdin if none given
        files: Vec<String>,
    },
    /// Sort YAML keys
    #[command(long_about = "Sorts YAML keys

Non-option arguments are names of files to sort.
If no filenames, use stdin.

-r --replace -- do an in-place sort
-a --auto -- automatic *.sorted.yaml filename")]
    Sort {
        /// YAML files to sort; reads stdin if none given
        files: Vec<String>,

        /// Do an in-place sort, replacing the file(s).
        #[arg(short, long)]
        replace: bool,

        /// Automatic *.sorted.yaml filename.
        #[arg(short, long)]
        auto: bool,
    },
    /// Generate shell completion scripts
    Completion {
        /// The shell to generate completions for
        shell: Shell,
    },
}

fn main() {
    let cli = Cli::parse();

    let result = match cli.command {
        Commands::List { files } => list::run(&files),
        Commands::Sort {
            files,
            replace,
            auto,
        } => sort::run(&files, replace, auto),
        Commands::Completion { shell } => {
            let mut cmd = Cli::command();
            clap_complete::generate(shell, &mut cmd, "yamlutil", &mut std::io::stdout());
            Ok(())
        }
    };

    if let Err(err) = result {
        eprintln!("Error: {err}");
        std::process::exit(1);
    }
}
