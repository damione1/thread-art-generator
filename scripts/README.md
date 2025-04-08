# Scripts Directory

This directory contains various scripts used in the Thread Art Generator project.

## Directory Structure

- `cli` - Command-line interface wrapper script
- `local_setup.sh` - Main setup script for local development

### Subdirectories

- `dev/` - Scripts for local development environment

  - `tilt-minio-setup.sh` - Sets up MinIO for local development

- `build/` - Scripts for building and compiling
  - `build-protos.sh` - Generates code from Protocol Buffer definitions

## Usage

### Initial Setup

Run the main setup script to configure your local development environment:

```bash
./scripts/local_setup.sh
```

### CLI Usage

The CLI wrapper script automatically builds and configures the CLI tool:

```bash
./scripts/cli <command>
```

### Development Scripts

Development scripts are typically run automatically by Tilt but can be executed manually if needed:

```bash
./scripts/dev/tilt-minio-setup.sh
```

### Build Scripts

Build scripts are used for generating code and compiling:

```bash
./scripts/build/build-protos.sh
```
