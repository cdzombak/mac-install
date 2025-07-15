### **Software Design Document: macOS Automated Setup System**

**Version:** 2.0
**Date:** 2025-07-07

### 1. Introduction

#### 1.1 Purpose
This document outlines the design, requirements, and behavior of an automated system for installing and configuring software on a macOS environment. The primary goal of the system is to ensure a consistent, repeatable, and maintainable setup process, reducing manual intervention and ensuring a desired state is achieved efficiently.

#### 1.2 Scope
The system is responsible for the following activities on a target macOS machine:
- Installation of command-line tools and graphical user interface (GUI) applications from multiple sources (Homebrew, Mac App Store, package managers, archives, scripts).
- Application of system-level and application-specific configurations.
- Management of prerequisites and dependencies required for the setup process itself (internal artifacts).
- Generation of a checklist for required manual configuration steps.
- Persistent state management for user choices (optional per software).
- Colored terminal output for enhanced user experience.

#### 1.3 Audience
This document is intended for the developer and user of the dotfiles repository, who will use this system to provision new machines or update existing ones.

#### 1.4 Definitions
- **Artifact:** Any piece of software to be installed, such as a command-line tool, a GUI application, or a system font. The file or directory that indicates successful installation.
- **Idempotent Operation:** An operation that can be applied multiple times without changing the result beyond the initial application.
- **Configuration:** A set of parameters or settings applied to the operating system or an application.
- **Checklist:** A user-facing document (`SystemSetup.md`) that tracks required manual actions.
- **Internal Artifacts:** Dependencies required by the system itself (e.g., Homebrew, brew-caveats) that are automatically installed when needed.
- **Persistence:** The ability to remember user choices about whether to install optional software across program runs.
- **Optional Group:** A group of software where users are prompted for each item (can be overridden with `optional: false`).
- **Archive Installation:** Installation method that downloads and extracts archives (.dmg, .zip) to install applications.

---

### 2. System Overview
The macOS Automated Setup System is a Go-based program designed to provision a new or existing macOS environment to a predefined, user-specific state. It automates the installation of a comprehensive suite of software from various sources and applies a consistent set of configurations with colored terminal output for enhanced user experience.

The system is designed around the core principles of:
- **Idempotency:** The program can be re-run safely at any time to install missing components or update configurations without causing errors or unintended side effects.
- **Modularity:** Functionality is broken into distinct Go packages (orchestrator, installer, config, checklist, state, colors).
- **Configurability:** The system allows for user interaction to include or exclude optional components, with configurable persistence of choices to avoid repetitive prompting. A command line flag allows skipping all optional groups entirely.
- **Multi-Source Support:** Supports installation from Homebrew, Mac App Store, package managers (npm, gem), archives (.dmg, .zip), and custom scripts.
- **Internal Dependency Management:** Automatically handles installation of prerequisites like Homebrew when needed.
- **Colored Output:** Provides visual feedback through tasteful, optional colored terminal output.

---

### 3. Functional Requirements
The system shall meet the following functional requirements:

| ID     | Requirement                               | Description                                                                                                                                     |
| :----- | :---------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------- |
| **FR-1** | **Automated Software Installation**       | The system must be capable of installing a predefined list of both command-line tools and GUI applications without manual user intervention.      |
| **FR-2** | **Idempotent Installation**               | The system must check for the existence of an artifact before attempting installation. If the artifact is already present, the installation step must be skipped. |
| **FR-3** | **Multi-Source Package Management**       | The system must integrate with multiple package and software distribution mechanisms. This includes, but is not limited to, Homebrew, Mac App Store (`mas`), and language-specific package managers (`npm`, `gem`). |
| **FR-4** | **Automated Post-Install Configuration**  | The system must programmatically apply configurations to installed applications and the operating system using system utilities like `defaults` and `osascript`. |
| **FR-5** | **Interactive Component Selection**       | For optional software components, the system must prompt the user for a decision (e.g., via a "y/N" prompt) on whether to proceed with the installation. |
| **FR-6** | **User Choice Persistence**               | The system must store user decisions from interactive prompts when `persist: true` is configured for the software. On subsequent runs, it must read this stored state and skip prompting for components the user has previously chosen to exclude. Software with `persist: false` (default) will be prompted about on every run. |
| **FR-7** | **Manual Action Tracking**                | The system must generate and update a user-facing checklist (`SystemSetup.md`) with items that require manual configuration (e.g., licensing, UI-based settings, account sign-in). This checklist must also be idempotent, preventing duplicate entries. |
| **FR-8** | **Platform Verification**                 | The system must detect the host operating system and execute macOS-specific logic only when running on a Darwin-based system. |
| **FR-9** | **Prerequisite Management**               | The system must ensure its own dependencies (e.g., Homebrew, Rosetta 2) are present and configured before proceeding with the main installation tasks. |
| **FR-10**| **Privileged Operation Handling**         | The system must handle operations requiring elevated privileges (e.g., via `sudo`) in a controlled manner, such as for setting system-wide paths or fixing file permissions. |
| **FR-11**| **Custom Script-Based Installation**      | The system must support the execution of external shell scripts to install artifacts. This allows for the installation of software that does not have a package available through the other supported package management systems. The script execution must be integrated into the standard idempotent installation workflow. |
| **FR-12**| **Archive-Based Installation**            | The system must support downloading and extracting archives (.dmg, .zip, .tar.gz) to install applications. It must download the archive to a temporary location, extract or mount it, and either copy a specified file/directory to the target location or extract all contents to the directory containing the artifact. |
| **FR-13**| **Checklist Backfill for Existing Software** | The system must automatically generate checklist entries for software that is already installed but has missing checklist headers. This ensures manual setup steps are always available. |
| **FR-14**| **Colored Terminal Output**               | The system must provide colored terminal output for enhanced user experience, with automatic detection of terminal capabilities and respect for NO_COLOR environment variable. |
| **FR-15**| **Optional vs Required Groups**           | The system must support both optional groups (where users are prompted for each software item) and required groups (where software is installed automatically without prompting). |

---

### 4. Non-Functional Requirements

| ID      | Requirement     | Description                                                                                                                               |
| :------ | :-------------- | :---------------------------------------------------------------------------------------------------------------------------------------- |
| **NFR-1** | **Maintainability** | The system’s architecture must be modular and its scripts clear, allowing the user to easily add, remove, or update software components. |
| **NFR-2** | **Reliability**     | Operations must be robust. The idempotent nature of the design contributes to this by allowing safe recovery from partial runs.         |
| **NFR-3** | **Usability**       | The system must provide clear, color-coded feedback to the user regarding its progress, successes, and failures.                          |

---

### 5. System Architecture and Design

#### 5.1 Core Components

1.  **Orchestrator:** A Go package that serves as the primary coordination layer. It processes internal artifacts first, then iterates through software groups, handling optional vs required groups, user prompting, state persistence, and checklist generation.

2.  **Config Manager:** A Go package that loads and parses YAML configuration files, handles variable expansion ($HOME, $BREW), manages embedded internal configuration, and provides methods for checking Homebrew requirements.

3.  **Installer:** A Go package that handles multiple installation methods including Homebrew (brew/cask), Mac App Store (mas), package managers (npm/gem), shell commands, scripts, and archive downloads/extraction.

4.  **Checklist Manager:** A Go package that manages the manual action checklist, ensures idempotent header creation, integrates Homebrew caveats, and handles checklist backfill for existing software.

5.  **State Store:** A Go package that manages user choice persistence in `~/.config/dotfiles/software/` directory using flag files, with configurable persistence per software item.

6.  **Colors:** A Go package that provides colored terminal output with automatic capability detection and NO_COLOR environment variable support.

#### 5.2 Key Processes and Workflows

**1. Overall Process:**
1. **Platform Check:** Verify the system is running on macOS (Darwin).
2. **Internal Artifacts:** If any software requires Homebrew, automatically install Homebrew and brew-caveats tool from embedded internal.yaml configuration.
3. **Group Processing:** Process each software group in order, respecting the `optional` flag for user prompting.

**2. Individual Software Processing:**
1. **State Check:** For optional groups, check if software was previously excluded (only if `persist: true`).
2. **Artifact Check:** Check if the target artifact (e.g., `/Applications/Foo.app`) exists.
3. **Skip or Install:**
   - **If it exists:** Report "already installed". If checklist steps are defined and no header exists in the checklist file, create missing checklist entries with any applicable Homebrew caveats.
   - **If it does not exist:** 
     - If no install steps are defined, add "Install [software]" to checklist.
     - For optional groups, prompt user "Install [software]? (y/N)" in colored text.
     - If user declines and `persist: true`, save choice to state store.
     - If user accepts or group is required, execute installation (brew, cask, mas, npm, gem, run, script, archive).
4. **Artifact Verification:** Ensure the artifact exists after installation.
5. **Configuration:** If post-install configuration steps exist and artifact is present, apply them (supports `ignore_errors: true`).
6. **Checklist Update:** If software was just installed and has checklist steps, add them to the checklist with any Homebrew caveats.

#### Error Handling

The `mac-install` program fails if the installation or configuration process for any piece of software fails. The idempotent nature of the program makes re-running it to resolve errors safe.

---

### 6. Data Design

#### 6.1 User Choices Data
-   **Storage:** Plain files within the `~/.config/dotfiles/software/` directory.
-   **Format:** A file's existence acts as a boolean flag. For example, the presence of `~/.config/dotfiles/software/no-tool` indicates the user has chosen not to install `tool`.
-   **Schema:** The file naming convention is `no-<normalized-software-name>`, where normalization converts to lowercase and replaces spaces and slashes with hyphens.
-   **Persistence:** Only created when `persist: true` is configured for the software. Software with `persist: false` (default) will not create state files.

#### 6.2 Manual Action Checklist
-   **Storage:** A single Markdown file located at the path specified by the `checklist` configuration field.
-   **Format:** Standard Markdown, using headers for categorization and checkboxes (`- [ ]`) for individual to-do items.
-   **Generation:** Checklist entries are appended idempotently by first checking if a header for the artifact already exists.
-   **Headers:** Headers use artifact display names. If the artifact path contains "/Applications/" or "/bin/", only the base name is used. Otherwise, the full path is used.
-   **Backfill:** For software that is already installed but has missing checklist headers, entries are automatically created when the program runs.
-   **Caveats Integration:** Homebrew caveats are automatically included in checklist entries for brew/cask installed software.

#### Installation Spec

The software installation spec is a single YAML file with the following format. Ordering is important in all lists in the file:

```
checklist: /Users/cdzombak/SystemSetup.md

install_groups:
  - group: Human-readable Group name (e.g. Core, Dev Tools)
    optional: true  # Optional: defaults to true. Set to false for always-install groups
    software:
      - name: Human-Readable Software Name
        artifact: /path/to/artifact.app
        persist: true  # Optional: remember user's choice not to install (defaults to false)
        install:
          - brew: packagename
          - cask: packagename
          - mas: "id"  # Must be quoted. Can be either an app ID (e.g., "1502933106") or an App Store URL (e.g., "https://apps.apple.com/us/app/bear/id1091189122")
          - npm: packagename
          - gem: packagename
          - run: command-to-run
          - script: /path/to/.dotfiles/mac/install/artifact.sh
          - archive: https://example.com/app.dmg
            file: Application.app
          - archive: https://fonts.example.com/fonts.zip  # Extracts all files to artifact directory
        configure:
          - run: echo "foo" > ~/.config/app.txt
          - script: /path/to/.dotfiles/mac/configure/artifact.sh
        checklist:
          - item 1
          - item 2

      - artifact: /Applications/Tool.app  # name defaults to "Tool.app"
        install:
          - cask: tool-package
        checklist:
          - Configure tool settings

```

##### Groups

The file consists of one or more groups. Each group contains definitions for one or more pieces of software to install. Each group must have a name and a list of software.

Groups have the following properties:
- `group`: Human-readable group name (required)
- `software`: List of software definitions (required)
- `optional`: Boolean indicating whether to prompt for each software item in the group (optional, defaults to true)

##### Software

Each software item must contain an artifact. All other keys are optional, including name. Keys are:

- `name`: human-readable software name (optional, defaults to artifact display name)
- `note`: optional note displayed to the user when working with this software (optional)
- `persist`: boolean indicating whether to remember the user's choice not to install this software (optional, defaults to false)
- `install`: a list of installation steps. Each step is a key/value pair. The key must be one of:
    - `brew`: install software using `brew install packagename`
    - `cask`: install software using `brew install --cask packagename`
    - `mas`: install software using `mas install id` (id must be quoted)
    - `npm`: install software using `npm install -g packagename`
    - `gem`: install software using `gem install packagename`
    - `pipx`: install software using `pipx install packagename`
    - `dl`: download file from URL and save directly to artifact path
    - `run`: run the given command, assuming it will produce the artifact (working directory: config file directory)
    - `script`: run the given shell script, assuming it will produce the artifact (working directory: config file directory)
    - `archive`: download and extract archive. If `file` parameter is provided, copies specific file/directory from archive. If `file` is omitted, extracts all archive contents to the directory containing the artifact.
- `configure`: a list of configuration steps to be run if the software artifact exists. Each step is a key/value pair. The key must be one of:
    - `ignore_errors`: if `true`, ignore errors produced by the remaining configuration steps, for this software only.
    - `run`: run the given command (working directory: config file directory)
    - `script`: run the given shell script (working directory: config file directory)
- `checklist`: a list of human-readable post-installation steps. After installing the software, these steps are written to the checklist, under a header for the artifact name.

Artifact names can contain the following variables, which are evaluated as follows:

- `$HOME`: the absolute path to the user's home directory
- `$BREW`: the output of `$(brew --prefix)`
- `$ENV_VARIABLE_NAME`: environment variables using the `$ENV_` prefix (e.g., `$ENV_ASDF_PY` expands to the value of the `ASDF_PY` environment variable). If the environment variable is not set, configuration loading will fail with an error.

If a software definition has no installation steps, and the artifact does not exist, simply add a checklist step "- [ ] Install <software name> to the checklist."

##### Working Directory for Scripts

When executing `run` or `script` commands (both in install and configure sections), the working directory is set to the directory containing the configuration YAML file. This allows scripts to use relative paths from the config file location. For example:
- If the config file is at `/Users/me/dotfiles/install.yaml`
- Scripts executed via `run` or `script` will have working directory `/Users/me/dotfiles/`
- This applies to both installation and configuration steps

##### Automatic Application Launch

When all of the following conditions are met:
1. A software item was just installed (not already present)
2. The artifact path ends with `.app`
3. The configuration steps include at least one `run` or `script` command

The system will automatically open the application using `open -a /path/to/artifact` before running the configuration steps. This ensures applications that need to be running for configuration (e.g., to accept preferences or perform initial setup) are launched automatically. The system waits 2 seconds after opening the application before proceeding with configuration.

---

### 4. Command Line Interface

The program accepts the following command line options:

- `-config <file>`: Specifies the path to the configuration YAML file (default: `./install.yaml`)
- `-skip-optional`: When set, completely skips all optional sections. No installation, configuration, or checklist related actions are taken for items in optional groups. This flag is useful for automated or non-interactive installations where only required software should be installed.
- `-only <name>`: When set, installs only a single piece of software from the configuration file. The system searches for software whose artifact basename or user-chosen name contains the provided value as a substring (case-insensitive). If multiple matches are found, the program lists all candidates and exits with an error, requiring the user to be more specific. When this flag is used, core dependencies setup is skipped, and only the matched software is processed (install, configure, and checklist updates as needed). Cannot be used together with `-skip-optional`.
