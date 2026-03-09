# Ti - System Specification Tool

ti (pronounced "tee"). Is a greek word which means "what"

A tool to display important system specifications in one command. Works on Linux, macOS, and Windows!

For me personally the pain point was that 
when someone asks "What are your system specs?", you currently have to run mutltiple commands due to it being scattered accross multiple commands.

Now I can just run `./ti`

## How to build
```
# Build for your current OS
go build -o ti

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o ti-macos-intel

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o ti-macos-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o ti.exe

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ti-amd64
```

# Usage
```
# Default table format
./ti

# JSON output (great for scripting)
./ti -format json

# Plain text output
./ti -format text
```