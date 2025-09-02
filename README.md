# 🚀 outputbuddy

[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=for-the-badge)](LICENSE)
[![Version](https://img.shields.io/badge/version-2.0.0-green?style=for-the-badge)](https://github.com/zmunro/outputbuddy/releases)

> **Flexible output redirection with color preservation** - Never lose your terminal colors when logging to files again!

outputbuddy (or `ob` for short) is a powerful command-line tool that lets you redirect stdout and stderr to multiple destinations simultaneously while preserving ANSI color codes and handling terminal control sequences intelligently.

## ✨ Features

- 🎨 **Preserves Colors**: Maintains ANSI color codes when redirecting to terminals
- 📝 **Smart ANSI Stripping**: Automatically removes ANSI codes from file outputs (configurable)
- 🔀 **Multiple Destinations**: Route stdout/stderr to any combination of files and terminal
- 🔄 **PTY Support**: Full pseudo-terminal support for interactive applications
- 🧹 **Progress Bar Filtering**: Intelligently filters out progress bars and spinners from logs
- ⚡ **High Performance**: Efficient buffering and parallel processing
- 🎯 **Flexible Routing**: Use intuitive shorthand syntax (1=stdout, 2=stderr)

## 📦 Installation

### From Source

```bash
go install github.com/zmunro/outputbuddy@latest
```

### Using Homebrew (macOS/Linux)

```bash
brew tap zmunro/outputbuddy
brew install outputbuddy
```

### Pre-built Binaries

Download the latest release from the [releases page](https://github.com/zmunro/outputbuddy/releases).

## 🚀 Quick Start

```bash
# Redirect both stdout and stderr to a file AND show on terminal
outputbuddy 2+1=output.log 2+1 -- python script.py

# Or use the short alias
ob 2+1=output.log 2+1 -- python script.py

# Separate stdout and stderr to different files
ob 1=out.log 2=err.log -- make

# Only log errors, but still show them on screen
ob 2=errors.log 2 -- ./my-program
```

## 📖 Usage

```
outputbuddy [options] -- command [args...]
```

### Options

| Option | Description |
|--------|-------------|
| `1=file.log` or `stdout=file.log` | Redirect stdout to file |
| `2=file.log` or `stderr=file.log` | Redirect stderr to file |
| `2+1=file.log` or `stderr+stdout=file.log` | Redirect both to same file |
| `1` or `stdout` | Show stdout on terminal |
| `2` or `stderr` | Show stderr on terminal |
| `2+1` or `stderr+stdout` | Show both on terminal |
| `--no-pty` | Disable PTY mode (use pipes instead) |
| `--keep-ansi` | Keep ANSI codes in file outputs |
| `--version`, `-v` | Show version information |

### Examples

#### 🎯 Basic Logging
```bash
# Log everything to a file while watching output
ob 2+1=build.log 2+1 -- cargo build --release
```

#### 🔍 Debug Logging
```bash
# Separate streams for debugging
ob 1=output.log 2=debug.log 1 2 -- node app.js
```

#### 🎨 Preserve Colors in Files
```bash
# Keep ANSI codes for later viewing with 'less -R'
ob --keep-ansi 2+1=colored.log -- npm test
```

#### 🤫 Silent Logging
```bash
# Log everything but show nothing on terminal
ob 2+1=silent.log -- ./batch-process.sh
```

#### ⚡ Development Workflow
```bash
# Perfect for development - see and log everything
ob 2+1=dev.log 2+1 -- npm run dev
```

## 🎮 Advanced Features

### PTY Mode

By default, outputbuddy uses a pseudo-terminal (PTY) to capture output. This ensures:
- Interactive applications work correctly
- Colors are preserved in terminal output
- Terminal size changes are handled properly

Disable PTY mode with `--no-pty` if you need pure pipe behavior.

### Smart Filtering

outputbuddy automatically filters out:
- Progress bars and spinners
- Carriage return overwrites
- Braille pattern characters
- Empty lines from progress updates

This keeps your log files clean and readable while preserving important output.

### Multiple File Handling

You can redirect the same stream to multiple files:
```bash
# Log errors to both general log and error-specific log
ob 2+1=all.log 2=errors-only.log 2+1 -- ./app
```

## 🏗️ Building from Source

```bash
# Clone the repository
git clone https://github.com/zmunro/outputbuddy.git
cd outputbuddy

# Build
go build -o outputbuddy

# Install to $GOPATH/bin
go install
```

### Requirements

- Go 1.18 or higher
- POSIX-compliant system (Linux, macOS, BSD)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [creack/pty](https://github.com/creack/pty) for PTY handling
- Inspired by the need for better logging in CI/CD pipelines

## 💡 Tips & Tricks

### Alias Setup
Add to your shell configuration:
```bash
alias ob='outputbuddy'
```

### Viewing Colored Logs
If you used `--keep-ansi`, view colored logs with:
```bash
less -R colored.log
# or
cat colored.log  # if your terminal supports it
```

### CI/CD Integration
Perfect for CI/CD pipelines where you need both real-time output and complete logs:
```bash
ob 2+1=ci-build.log 2+1 -- make test
```

---

<p align="center">
  Made with ❤️ by developers, for developers
</p>

<p align="center">
  <a href="https://github.com/zmunro/outputbuddy/issues">Report Bug</a>
  ·
  <a href="https://github.com/zmunro/outputbuddy/issues">Request Feature</a>
</p>
