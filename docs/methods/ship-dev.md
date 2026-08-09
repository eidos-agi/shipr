<!-- Absorbed from eidos-agi/ship-forge (.claude/skills/ship-dev.md) into shipr.
     Prefer: shipr model/attempt/frontier (Go). Agents run proofs themselves.
     Do not treat this as an executable ship runner. -->

# /ship-dev — Local Development Workflow

## When to Use
When developing a Python package that's published on PyPI and used by downstream projects. This skill sets up the local dev loop so changes are testable instantly, and guides the pre-release → release lifecycle.

## What It Does
Configures editable installs, validates the dev environment, and walks through the confidence layers from local edit to PyPI release.

## Execution Steps

### 1. Check Current State

```bash
# What version is installed?
pip show <package-name>

# Is it an editable install?
pip show -f <package-name> | head -5
# Editable installs show "Location: /path/to/source" not site-packages

# What version is on PyPI?
pip index versions <package-name> 2>/dev/null || pip install <package-name>==nonexistent 2>&1 | grep "from versions"
```

### 2. Set Up Editable Install

```bash
cd /path/to/package/source
pip install -e ".[dev]"    # or pip install -e . if no dev extras
```

Verify:
```bash
python -c "import <package>; print(<package>.__file__)"
# Should point to your source directory, not site-packages
```

**What this does:** Creates a link from site-packages to your source directory. Edits to source files take effect immediately — no reinstall needed. This is the standard for all library development.

### 3. Develop and Test Locally

```bash
# Run tests
pytest

# Run the package (if it's an MCP server, CLI, etc.)
python -m <package>.server
```

Every save to a source file is instantly reflected. This is your inner loop — keep it fast.

### 4. Test in a Downstream Project

**Option A: Git branch (quick validation)**

Push your changes to a branch, then in the downstream project:

```bash
# Install from your branch instead of PyPI
pip install "package-name @ git+https://github.com/org/repo.git@branch-name"

# Run the downstream project's tests
pytest
```

**Option B: Local path (fastest, same machine)**

```bash
# In the downstream project's venv
pip install -e /path/to/package/source

# Now both projects share the same source
```

### 5. Pre-Release to PyPI

When you're confident but want broader testing:

```bash
# Bump version with pre-release tag
# In pyproject.toml: version = "1.5.0a1"

# Build
python -m build

# Upload to PyPI (pre-releases are invisible to normal pip install)
twine upload dist/*
```

Downstream projects opt in:
```bash
pip install --pre package-name          # latest pre-release
pip install package-name==1.5.0a1       # exact pre-release
```

### 6. Full Release

```bash
# Remove pre-release tag
# In pyproject.toml: version = "1.5.0"

# Tag and push
git tag v1.5.0
git push --tags

# If CI/CD is set up (publish-verified.yml), it publishes automatically
# Otherwise: python -m build && twine upload dist/*
```

### 7. Verify the Release

```bash
# Install in a clean venv
python -m venv /tmp/verify-release && source /tmp/verify-release/bin/activate
pip install package-name
python -c "import package; print(package.__version__)"

# Run smoke tests
# Clean up
deactivate && rm -rf /tmp/verify-release
```

## Common Problems

| Problem | Fix |
|---------|-----|
| Editable install not reflecting changes | Restart the Python process (MCP servers need restart) |
| `pip install -e .` fails | Check pyproject.toml has `[build-system]` and correct package discovery |
| Pre-release visible to normal install | Version tag must contain `a`, `b`, or `rc` (e.g., `1.5.0a1`) |
| Downstream project not picking up branch | Clear pip cache: `pip cache purge` then reinstall |
| Multiple venvs confused | Check `which python` and `pip show` to verify you're in the right env |

## Rules
- **Never develop against the installed PyPI version.** Always use editable install for active development.
- **Never publish without testing in at least one downstream project.** Git branch or local path install.
- **Pre-release tags protect users.** Use them for anything that hasn't been validated in production.
- **Tag versions in git.** The git tag is the source of truth, not pyproject.toml alone.
- **Test what you ship.** Build the wheel, install it in a clean venv, verify it imports. See `/ship-check`.
