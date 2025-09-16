# `mac-install`

A program for installing and configuring a software suite on macOS. This tool ensures consistent, repeatable, and maintainable setup processes while reducing manual intervention.

## Features

- **Idempotent Operations**: Safe to run multiple times without side effects
- **Multi-Source Installation**: Supports Homebrew, Mac App Store, Go modules, npm, gem, and custom scripts
- **Required vs Optional Groups**: Some groups always install, others prompt the user
- **State Persistence**: Remembers user choices to avoid re-prompting
- **Automated Configuration**: Applies post-install configurations automatically
- **Manual Task Tracking**: Generates checklist for required manual steps
- **Internal Artifact Management**: Automatically installs Homebrew and dependencies when needed

## Installation

### Homebrew

```shell
brew install cdzombak/oss/mac-install
```

### Manual from build artifacts

Pre-built binaries for macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/mac-install/releases).

## Quick Start

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

   💡 **Pro tip**: Use the included YAML schema (`schema.yaml`) for autocompletion and validation in your editor. The project includes VS Code settings for automatic schema detection. See [SCHEMA.md](SCHEMA.md) for detailed setup instructions for various editors.

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

The configuration is defined in YAML format with the following structure. See [`@cdzombak/dotfiles/mac/install.yaml`](https://github.com/cdzombak/dotfiles/blob/master/mac/install.yaml) for a real-world example.

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
- `note`: Optional note displayed to the user when prompting for installation (useful for warnings, size information, etc.)
- `persist`: Boolean indicating whether to remember user's choice not to install (defaults to false)
- `install`: Array of installation steps
- `configure`: Array of configuration steps
- `checklist`: Array of manual post-installation steps

**Note:** Artifact paths support asterisk (`*`) wildcards for version-agnostic matching. See [Wildcard Support](#wildcard-support) section for details.

### Installation Methods

The `install` section supports these methods:

- `brew: package-name` - Install via Homebrew
- `cask: package-name` - Install GUI app via Homebrew Cask
- `mas: app-id` - Install from Mac App Store. Accepts either an app ID (e.g., `"1502933106"`) or an App Store URL (e.g., `"https://apps.apple.com/us/app/meshman-3d-viewer-pro/id1502933106?mt=12"`). **NOTE:** The value must be enclosed in quotes.
- `npm: package-name` - Install global npm package
- `gem: package-name` - Install Ruby gem
- `gomod: package-name` - Install Go module via Homebrew
- `pipx: package-name` - Install Python package via pipx
- `dl: url` - Download file from URL and save directly to artifact path
- `run: command` - Execute shell command
- `script: /path/to/script.sh` - Run shell script
- `archive: url` + `file: filename` - Download and extract archive (.dmg, .zip, .tar.gz), then copy specified file to /Applications
- `archive: url` (without `file`) - Download and extract all files from archive to the directory containing the artifact

**Note:** Archive type is automatically detected from the URL (e.g., URLs containing `.dmg`, `.zip`, `.tar.gz`), HTTP Content-Type headers, or from the downloaded file extension. Supported formats include DMG (disk images), ZIP archives, and TAR.GZ compressed archives.

### Configuration Methods

The `configure` section supports:

- `ignore_errors: true` - Ignore errors in subsequent configuration steps
- `run: command` - Execute shell command
- `script: /path/to/script.sh` - Run shell script

### Automatic Application Launch

When installing a `.app` application that has `run` or `script` configuration steps, the application will be automatically opened before configuration begins. This ensures apps that need to be running for configuration are launched automatically.

**Conditions for automatic launch:**
1. Software was just installed (not already present)
2. Artifact path ends with `.app`
3. Configuration steps include at least one `run` or `script` command

The system waits 2 seconds after opening the application before proceeding with configuration to allow the app to start up properly.

**Note:** If the application cannot be opened, a warning is displayed but the installation process continues.

### Variable Expansion

The following variables are automatically expanded:
- `$HOME`: User's home directory
- `$BREW`: Homebrew prefix (typically `/opt/homebrew` or `/usr/local`)
- `$ENV_VARIABLE_NAME`: Environment variables using the `$ENV_` prefix (e.g., `$ENV_ASDF_PY` expands to the value of the `ASDF_PY` environment variable)

**Note:** If an environment variable referenced with `$ENV_` is not set, the configuration loading will fail with an error message.

### Wildcard Support

Artifact paths support asterisk (`*`) wildcards for version-agnostic matching. This is useful when application names include version numbers that may change over time.

**Examples:**
- `/Applications/OpenSCAD*.app` matches both `OpenSCAD.app` and `OpenSCAD-2021.01.app`
- `$HOME/Library/Application Support/MyApp*/config.json` matches version-specific directories
- `$BREW/bin/tool-*` matches versioned command-line tools

Wildcards use Go's `filepath.Glob` pattern matching and will return true if at least one matching file or directory is found.

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

### Software with User Notes

```yaml
- name: Xcode
  artifact: /Applications/Xcode.app
  note: This is a large download and may take 30+ minutes
  install:
    - mas: "497799835"
  checklist:
    - Accept Xcode license agreement
    - Install additional components when prompted
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
# Install specific file from a DMG archive
- name: Custom Application
  artifact: /Applications/CustomApp.app
  install:
    - archive: https://example.com/releases/CustomApp.dmg
      file: CustomApp.app  # File/directory to copy from the archive
  checklist:
    - Launch CustomApp and complete setup
    - Enter license key if required

# Extract all files from ZIP archive to target directory
- name: Font Collection
  artifact: /Library/Fonts/CustomFont.ttf
  install:
    - archive: https://fonts.example.com/font-pack.zip
      # No 'file' parameter - extracts all files to /Library/Fonts/
  checklist:
    - Verify fonts appear in Font Book
    - Test fonts in applications

# TAR.GZ archive with specific binary extraction
- name: Command Line Tool
  artifact: /usr/local/bin/tool
  install:
    - archive: https://github.com/vendor/tool/releases/download/v1.0/tool.tar.gz
      file: tool  # Binary file to copy
```

### Wildcard Artifact Paths

```yaml
# Version-agnostic application matching
- name: OpenSCAD
  artifact: /Applications/OpenSCAD*.app  # Matches OpenSCAD.app, OpenSCAD-2021.01.app, etc.
  install:
    - cask: openscad
  checklist:
    - Configure 3D rendering preferences
    - Import custom libraries

# Versioned configuration files
- name: Development Config
  artifact: $HOME/.config/myapp-*/settings.json
  install:
    - dl: https://example.com/config/settings.json
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
- Filename normalization: lowercase, spaces→hyphens, slashes→hyphens, `.app` suffix removed
- Software with `persist: false` (default) will not create state files and will be prompted about every run

**Examples of state file names:**
- "Visual Studio Code" → `no-visual-studio-code`
- "1Password 7 - Password Manager.app" → `no-1password-7---password-manager`
- "My App/Tool" → `no-my-app-tool`

### Checklist Generation

- Manual steps written to specified checklist file (typically `~/SystemSetup.md`)
- Uses Markdown format with checkboxes
- Headers based on artifact display names
- Idempotent: won't create duplicate entries
- Includes Homebrew caveats when applicable
- **Automatically creates checklist entries for already-installed software** if the header is missing

## Error Handling

- Program exits with failure if any installation or configuration step fails
- Idempotent design allows safe re-running to resolve errors
- Configuration steps can be set to ignore errors with `ignore_errors: true`
