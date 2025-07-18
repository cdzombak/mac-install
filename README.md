# macOS Automated Setup System

A Go-based automated system for installing and configuring software on macOS environments. This tool ensures consistent, repeatable, and maintainable setup processes while reducing manual intervention.

## Features

- **Idempotent Operations**: Safe to run multiple times without side effects
- **Multi-Source Installation**: Supports Homebrew, Mac App Store, npm, gem, and custom scripts
- **Interactive Selection**: Prompts for optional software components with colored output
- **Required vs Optional Groups**: Some groups always install, others prompt the user
- **State Persistence**: Remembers user choices to avoid re-prompting
- **Automated Configuration**: Applies post-install configurations automatically
- **Manual Task Tracking**: Generates checklist for required manual steps
- **Internal Artifact Management**: Automatically installs Homebrew and dependencies when needed
- **Platform Verification**: Ensures macOS-only execution
- **Colored Output**: Tasteful, optional colored terminal output for better user experience

## Quick Start

1. **Build the program**:
   ```bash
   go build -o mac-install
   ```

2. **Create a configuration file** (see `install.example.yaml` for a complete example):
   ```yaml
   checklist: $HOME/SystemSetup.md
   install_groups:
     - group: Essential Tools
       software:
         - name: Git
           artifact: $BREW/bin/git
           install:
             - brew: git
   ```

   💡 **Pro tip**: Use the included YAML schema (`schema.yaml`) for autocompletion and validation in your editor. See [SCHEMA.md](SCHEMA.md) for setup instructions.

3. **Run the installer**:
   ```bash
   ./mac-install -config install.example.yaml
   ```
   
   To skip all optional sections:
   ```bash
   ./mac-install -config install.example.yaml -skip-optional
   ```
   
   To install only a single piece of software:
   ```bash
   ./mac-install -config install.example.yaml -only "Autodesk"
   ```

## Configuration Format

The configuration is defined in YAML format with the following structure:

### Root Level

- `checklist`: Path to the checklist file where manual setup steps are written
- `install_groups`: Array of software groups to install

### Install Groups

Each group contains:
- `group`: Human-readable group name (e.g., "Core Tools", "Development")
- `optional`: Boolean indicating whether to prompt for each software item (optional, defaults to true)
- `software`: Array of software definitions

### Software Definitions

Each software item must have:
- `artifact`: Path to the installed artifact (file/app that indicates successful installation)

Optional fields:
- `name`: Human-readable software name (defaults to artifact display name if not provided)
- `persist`: Boolean indicating whether to remember user's choice not to install (defaults to false)
- `install`: Array of installation steps
- `configure`: Array of configuration steps
- `checklist`: Array of manual post-installation steps

### Installation Methods

The `install` section supports these methods:

- `brew: package-name` - Install via Homebrew
- `cask: package-name` - Install GUI app via Homebrew Cask
- `mas: app-id` - Install from Mac App Store. Accepts either an app ID (e.g., `"1502933106"`) or an App Store URL (e.g., `"https://apps.apple.com/us/app/meshman-3d-viewer-pro/id1502933106?mt=12"`). **NOTE:** The value must be enclosed in quotes.
- `npm: package-name` - Install global npm package
- `gem: package-name` - Install Ruby gem
- `pipx: package-name` - Install Python package via pipx
- `dl: url` - Download file from URL and save directly to artifact path
- `run: command` - Execute shell command
- `script: /path/to/script.sh` - Run shell script
- `archive: url` + `file: filename` - Download and extract archive (.dmg, .zip, .tar.gz), then copy specified file to /Applications
- `archive: url` (without `file`) - Download and extract all files from archive to the directory containing the artifact

**Note:** Archive type is automatically detected from the URL (e.g., URLs containing `.dmg`, `.zip`, `.tar.gz`) or from the downloaded file extension.

### Configuration Methods

The `configure` section supports:

- `ignore_errors: true` - Ignore errors in subsequent configuration steps
- `run: command` - Execute shell command
- `script: /path/to/script.sh` - Run shell script

**Note:** When installing a `.app` application that has `run` or `script` configuration steps, the application will be automatically opened before configuration begins. This ensures apps that need to be running for configuration are launched.

### Variable Expansion

The following variables are automatically expanded:
- `$HOME`: User's home directory
- `$BREW`: Homebrew prefix (typically `/opt/homebrew` or `/usr/local`)
- `$ENV_VARIABLE_NAME`: Environment variables using the `$ENV_` prefix (e.g., `$ENV_ASDF_PY` expands to the value of the `ASDF_PY` environment variable)

**Note:** If an environment variable referenced with `$ENV_` is not set, the configuration loading will fail with an error message.

## Usage Examples

### Required vs Optional Groups

```yaml
# Required group - installs automatically
- group: Essential Tools
  optional: false
  software:
    - name: Homebrew
      artifact: /opt/homebrew/bin/brew
      # Homebrew is automatically installed as an internal artifact

# Optional group - prompts user
- group: Optional Development Tools
  optional: true  # This is the default
  software:
    - name: Visual Studio Code
      artifact: /Applications/Visual Studio Code.app
      install:
        - cask: visual-studio-code
```

### Basic Software Installation

```yaml
- name: Visual Studio Code
  artifact: /Applications/Visual Studio Code.app
  install:
    - cask: visual-studio-code
  configure:
    - run: code --install-extension ms-vscode.vscode-json
  checklist:
    - Sign in to Settings Sync
    - Configure preferred themes
```

### Multiple Installation Methods

```yaml
- name: Node.js
  artifact: $BREW/bin/node
  install:
    - brew: node
    - npm: npm@latest
  configure:
    - run: npm config set init-license "MIT"
```

### Error Handling in Configuration

```yaml
- name: Docker Desktop
  artifact: /Applications/Docker.app
  install:
    - cask: docker
  configure:
    - ignore_errors: "true"
    - run: docker --version  # May fail if Docker isn't running
```

### Manual Installation Only

```yaml
- name: Custom Software
  artifact: /Applications/Custom.app
  # No install steps - will prompt to add manual installation to checklist
  checklist:
    - Download from vendor website
    - Install manually
```

### Persist User Choices

```yaml
# Software that remembers user's choice not to install
- name: Optional Tool (Remembered)
  artifact: /Applications/OptionalTool.app
  persist: true  # Remember choice - won't ask again if user says no
  install:
    - cask: optional-tool

# Software that asks every time
- name: Optional Tool (Always Ask)
  artifact: /Applications/AlwaysAsk.app
  persist: false  # Default - will ask on every run
  install:
    - cask: always-ask-tool
```

### Archive Installation

```yaml
# Install specific file from a downloadable archive
- name: Custom Application
  artifact: /Applications/CustomApp.app
  install:
    - archive: https://example.com/releases/CustomApp.dmg
      file: CustomApp.app  # File/directory to copy from the archive
  checklist:
    - Launch CustomApp and complete setup
    - Enter license key if required

# Extract all files from archive to target directory
- name: Font Collection
  artifact: /Library/Fonts/CustomFont.ttf
  install:
    - archive: https://fonts.example.com/font-pack.zip
      # No 'file' parameter - extracts all files to /Library/Fonts/
  checklist:
    - Verify fonts appear in Font Book
    - Test fonts in applications

# Support for various archive formats
- name: Command Line Tool
  artifact: /usr/local/bin/tool
  install:
    - archive: https://github.com/vendor/tool/releases/download/v1.0/tool.tar.gz
      file: tool  # Binary file to copy
```

## Command Line Options

- `-config <file>`: Path to configuration YAML file (default: `./install.yaml`)
- `-skip-optional`: Skip all optional sections - no installation, configuration, or checklist actions are taken for items in optional groups
- `-only <name>`: Install only a single piece of software matching this name. Searches both user-chosen names and artifact basenames. If multiple matches are found, lists candidates and exits with error. Cannot be used with `-skip-optional`.

## Program Behavior

### Installation Workflow

1. **Internal Artifacts**: Automatically installs Homebrew and dependencies if any software requires them
2. **Group Processing**: Processes each software group in order
3. **Artifact Check**: Verifies if the target artifact already exists
4. **Skip or Install**: 
   - If exists: Reports "already installed", checks for missing checklist items, and skips to configuration
   - If missing: For optional groups, prompts user; for required groups, installs automatically
5. **Configuration**: Applies post-install configurations if artifact exists
6. **Checklist Update**: Adds manual steps to checklist for newly installed software or existing software with missing checklist items

### User Interaction

- For optional groups only: prompts "Install [software]? (y/N)" in colored cyan text
- Required groups (optional: false) install automatically without prompting
- User choices are persisted in `~/.config/dotfiles/software/` as flag files only when `persist: true`
- When `persist: false` (default), software will be prompted about on every run
- Subsequent runs respect previous choices and don't re-prompt for persisted software
- Colored output provides visual feedback (green=success, yellow=warning, red=error, blue=info)

### State Management

- Exclusion flags stored as files named `no-[normalized-software-name]` only when `persist: true`
- State directory: `~/.config/dotfiles/software/`
- Filename normalization: lowercase, spaces→hyphens, slashes→hyphens
- Software with `persist: false` (default) will not create state files and will be prompted about every run

### Checklist Generation

- Manual steps written to specified checklist file (typically `~/SystemSetup.md`)
- Uses Markdown format with checkboxes
- Headers based on artifact display names
- Idempotent: won't create duplicate entries
- Includes Homebrew caveats when applicable
- **Automatically creates checklist entries for already-installed software** if the header is missing

## Output and Interaction

### Colored Output

The program provides colored terminal output for better user experience:
- **Green**: Success messages and "already installed" status
- **Yellow**: Warnings and "no installation steps" messages
- **Red**: Error messages
- **Blue**: Info messages like "Installing..." and "Configuring..."
- **Cyan + Bold**: User prompts
- **Magenta + Bold**: Group headers
- **Bold**: Software names
- **Dim**: Skipped items and secondary information

Colors are automatically disabled when:
- `NO_COLOR` environment variable is set
- `TERM` is empty or set to "dumb"
- Output is redirected (non-terminal)

## Error Handling

- Program exits with failure if any installation or configuration step fails
- Idempotent design allows safe re-running to resolve errors
- Configuration steps can be set to ignore errors with `ignore_errors: true`

## Architecture

The program is structured with modular components:

- **Main Orchestrator** (`internal/orchestrator`): Coordinates the installation process
- **Config Manager** (`internal/config`): Loads and processes YAML configuration, manages embedded internal.yaml
- **Installer** (`internal/installer`): Handles various installation methods
- **Checklist Manager** (`internal/checklist`): Manages manual task tracking
- **State Store** (`internal/state`): Persists user choices
- **Colors** (`internal/colors`): Provides terminal color support with automatic detection

## Testing

Run the test suite:

```bash
go test ./...
```

Tests cover:
- Configuration loading and variable expansion
- Installation method execution
- State persistence and retrieval
- Checklist generation and idempotency
- Orchestrator workflow logic
- Internal artifact management
- Optional group handling
- Color output functionality

## Development

### Prerequisites

- Go 1.21 or later
- macOS (Darwin) for testing

### Building

```bash
go build -o mac-install
```

### Project Structure

```
mac-install/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── install.example.yaml    # Example configuration
├── schema.yaml             # YAML schema for editor support
├── SPEC.md                 # Detailed specification
├── SCHEMA.md               # Schema usage documentation
├── README.md               # This file
├── Makefile                # Build and development tasks
├── .vscode/                # VS Code configuration
│   └── settings.json       # YAML schema settings
└── internal/
    ├── config/             # Configuration management + embedded internal.yaml
    ├── orchestrator/       # Main coordination logic
    ├── installer/          # Installation methods
    ├── checklist/          # Checklist generation
    ├── state/              # User choice persistence
    └── colors/             # Terminal color support
```

## Contributing

1. Follow the existing code structure and patterns
2. Add comprehensive tests for new functionality
3. Update documentation for new features
4. Ensure idempotent behavior for all operations

## License

This project follows standard open-source practices. See license file for details.
