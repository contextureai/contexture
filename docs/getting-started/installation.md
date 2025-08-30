# Installation

This guide provides instructions for installing `contexture`.

## System Requirements

- **Git**: Required for fetching remote rules from repositories.
- **Terminal/Command Line**: Basic familiarity with a command-line interface.

## Installation Methods

### Option 1: Build from Source (Recommended)

1.  **Install Go**

    Go 1.21 or later is required. It can be downloaded from [golang.org](https://golang.org/dl/).

2.  **Clone the Repository**
    ```bash
    git clone https://github.com/contextureai/contexture.git
    cd contexture
    ```

3.  **Build the Binary**
    ```bash
    make build
    ```
    This creates the binary at `bin/contexture`.

4.  **Install System-wide (Optional)**
    ```bash
    make install
    # This installs to your GOBIN directory.
    ```
    Alternatively, manually copy the binary to a directory in your `PATH`:
    ```bash
    sudo cp bin/contexture /usr/local/bin/
    ```

### Option 2: Go Install

With a working Go installation, `contexture` can be installed with:

```bash
go install github.com/contextureai/contexture/cmd/contexture@latest
```
This command installs the binary to your `$GOPATH/bin` directory.

## Verify Installation

After installation, verify that `contexture` is working correctly:

```bash
# Check the version
contexture --version

# Show help output
contexture --help
```

The `version` command should produce output similar to:
```
contexture version 1.0.0
```

## Next Steps

After installation, proceed to the [Quick Start Tutorial](quick-start.md).

## Updating

To update to a newer version:

1.  **Check the current version**: `contexture --version`
2.  **Pull the latest source and rebuild**:
    ```bash
    cd /path/to/contexture
    git pull origin main
    make build
    make install  # or copy bin/contexture to your PATH
    ```
3.  **Verify the update**: `contexture --version`

If you used `go install`:
```bash
go install github.com/contextureai/contexture/cmd/contexture@latest
```