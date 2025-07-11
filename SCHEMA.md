# YAML Schema Usage

This project includes a comprehensive YAML schema (`schema.yaml`) that provides autocompletion, validation, and documentation for configuration files.

## Editor Setup

### VS Code

The project includes a `.vscode/settings.json` file that automatically configures VS Code to use the schema for YAML files matching the pattern:
- `install.example.yaml`
- `*.install.yaml` 
- `**/install*.yaml`

**Manual Setup (if needed):**

1. Install the "YAML" extension by Red Hat
2. Add to your VS Code settings.json:

```json
{
  "yaml.schemas": {
    "./schema.yaml": [
      "install.example.yaml",
      "*.install.yaml", 
      "**/install*.yaml"
    ]
  }
}
```

### Other Editors

#### Vim/Neovim with coc.nvim

Add to your `coc-settings.json`:

```json
{
  "yaml.schemas": {
    "./schema.yaml": [
      "install.yaml",
      "*.install.yaml",
      "**/install*.yaml"
    ]
  }
}
```

#### JetBrains IDEs (IntelliJ, WebStorm, etc.)

1. Go to Settings → Languages & Frameworks → Schemas and DTDs → JSON Schema Mappings
2. Add a new mapping:
   - **Name**: macOS Install Config
   - **Schema file or URL**: Point to `schema.yaml`
   - **File path pattern**: `install.example.yaml`, `*.install.yaml`

#### Emacs with lsp-mode

Add to your configuration:

```elisp
(with-eval-after-load 'lsp-yaml
  (lsp-yaml-set-buffer-schema "file:///path/to/schema.yaml"))
```

## Schema Features

The schema provides:

### 🎯 **Autocompletion**
- Property names (`checklist`, `install_groups`, `software`, etc.)
- Installation methods (`brew`, `cask`, `mas`, `npm`, `gem`, `run`, `script`)
- Configuration methods (`run`, `script`, `ignore_errors`)
- Boolean values for `optional` field

### ✅ **Validation**
- Required properties (e.g., `checklist`, `group`, `name`, `artifact`)
- Property types (strings, arrays, booleans)
- Valid values (e.g., `ignore_errors` must be "true" or "false")
- Pattern matching (e.g., MAS IDs must be numeric)

### 📚 **Documentation**
- Hover descriptions for all properties
- Examples for common use cases
- Pattern explanations

### 🚨 **Error Detection**
- Missing required fields
- Invalid property names
- Type mismatches
- Invalid values

## Schema Structure

```yaml
# Root level
checklist: string           # Required: Path to checklist file
install_groups: array       # Required: Array of install groups

# Install group level  
group: string              # Required: Group name
optional: boolean          # Optional: Whether to prompt (default: true)
software: array            # Required: Array of software definitions

# Software level
name: string               # Required: Software name
artifact: string           # Required: Path to artifact
note: string               # Optional: User-facing note
install: array             # Optional: Installation steps
configure: array           # Optional: Configuration steps  
checklist: array           # Optional: Manual steps

# Installation methods (one per step)
brew: string               # Homebrew package
cask: string               # Homebrew cask
mas: string                # Mac App Store ID
npm: string                # NPM package
gem: string                # Ruby gem
run: string                # Shell command
script: string             # Shell script path

# Configuration methods (one per step)
run: string                # Shell command
script: string             # Shell script path
ignore_errors: "true"|"false"  # Ignore subsequent errors
```

## Validation Examples

### ✅ Valid Configuration

```yaml
checklist: $HOME/SystemSetup.md
install_groups:
  - group: Core Tools
    optional: false
    software:
      - name: Git
        artifact: $BREW/bin/git
        install:
          - brew: git
        configure:
          - run: git config --global init.defaultBranch main
        checklist:
          - Configure Git user settings
```

### ❌ Invalid Configurations

**Missing required field:**
```yaml
install_groups:
  - group: Tools
    software:
      - name: Git
        # Missing 'artifact' field - will show error
        install:
          - brew: git
```

**Invalid installation method:**
```yaml
software:
  - name: Git
    artifact: $BREW/bin/git
    install:
      - invalid_method: git  # Unknown method - will show error
```

**Invalid boolean value:**
```yaml
install_groups:
  - group: Tools
    optional: "yes"  # Should be true/false - will show error
```

## Variable Support

The schema recognizes these variable patterns:
- `$HOME` - User home directory
- `$BREW` - Homebrew prefix

These are documented in the schema and will be highlighted appropriately by your editor.

## Contributing to the Schema

When adding new features to the configuration format:

1. Update `schema.yaml` with new properties
2. Add examples and descriptions
3. Test with your editor to ensure autocompletion works
4. Update this documentation

## Troubleshooting

### Schema Not Working

1. **Check file association**: Ensure your file matches the pattern (`install.yaml`, etc.)
2. **Reload editor**: Some editors need to be reloaded after schema changes
3. **Check extension**: Ensure YAML language support is installed
4. **Validate schema**: Use online JSON Schema validators to check schema syntax

### Performance Issues

If the schema causes performance issues with large files:
1. Disable validation temporarily: Set `yaml.validate: false`
2. Use schema only for smaller config files
3. Consider splitting large configurations into multiple files