# Distribution Guide: Making GoTunnel a Downloadable Package

This guide shows how to distribute GoTunnel as installable binaries like ngrok.

## 📋 Prerequisites

- GitHub account (or GitLab, Gitea, etc.)
- Go 1.21+ installed
- Git initialized in this directory

## 🔧 Step 1: Update Module Path

Replace `goTunnel` with your repository path in `go.mod`:

```go
// Before:
module goTunnel

// After (replace navpreet032):
module github.com/navpreet032/goTunnel
```

**Example:**

```go
module github.com/johndoe/goTunnel
```

Then update all imports in your code:

```bash
# Find and replace all imports
find . -type f -name "*.go" -exec sed -i '' 's|goTunnel/|github.com/navpreet032/goTunnel/|g' {} +

# Or manually update each file:
# From: import "goTunnel/pkg/protocol"
# To:   import "github.com/navpreet032/goTunnel/pkg/protocol"
```

Run `go mod tidy` to update dependencies:

```bash
go mod tidy
```

## 📦 Step 2: Prepare for Binary Installation

Your project is already structured correctly for `go install`:

```
cmd/
  ├── server/main.go    → installs as 'server'
  └── client/main.go    → installs as 'client'
```

**Rename binaries** for better user experience:

```bash
# Move files to create better command names
mv cmd/server cmd/tunnel-server
mv cmd/client cmd/tunnel-client
```

Update directory structure:

```
cmd/
  ├── tunnel-server/main.go
  └── tunnel-client/main.go
```

## 🚀 Step 3: Push to GitHub

```bash
# Initialize git (if not already)
git init

# Add remote repository (create repo on GitHub first)
git remote add origin https://github.com/navpreet032/goTunnel.git

# Add all files
git add .

# Commit
git commit -m "Initial release: GoTunnel v1.0.0"

# Push to GitHub
git push -u origin main
```

## 📥 Step 4: Users Can Install Your Package

Once pushed to GitHub, users can install directly:

### Option A: Install Latest Version

```bash
# Install server
go install github.com/navpreet032/goTunnel/cmd/tunnel-server@latest

# Install client
go install github.com/navpreet032/goTunnel/cmd/tunnel-client@latest

# Or install both at once
go install github.com/navpreet032/goTunnel/cmd/...@latest
```

Commands will be available as `tunnel-server` and `tunnel-client` in `$GOPATH/bin` (usually `~/go/bin`).

### Option B: Install Specific Version

```bash
go install github.com/navpreet032/goTunnel/cmd/tunnel-server@v1.0.0
```

## 🏷️ Step 5: Create GitHub Releases (Recommended)

Create semantic versioned tags:

```bash
# Tag your release
git tag -a v1.0.0 -m "Release v1.0.0: Initial stable release"
git push origin v1.0.0
```

On GitHub:

1. Go to your repository
2. Click "Releases" → "Create a new release"
3. Select the tag (v1.0.0)
4. Add release notes
5. Publish release

## 🎁 Step 6: Multi-Platform Binaries with GoReleaser (Optional)

For pre-built binaries (Linux, macOS, Windows) like ngrok:

### Install GoReleaser

```bash
# macOS
brew install goreleaser

# Or download from https://goreleaser.com/install/
```

### Create `.goreleaser.yaml`

```yaml
# .goreleaser.yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: tunnel-server
    binary: tunnel-server
    main: ./cmd/tunnel-server
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

  - id: tunnel-client
    binary: tunnel-client
    main: ./cmd/tunnel-client
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else }}{{ .Arch }}{{ end }}
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"

release:
  github:
    owner: navpreet032
    name: goTunnel
```

### Build and Release

```bash
# Test locally (doesn't push)
goreleaser release --snapshot --clean

# Create release on GitHub
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Build and upload to GitHub
export GITHUB_TOKEN="your_github_personal_access_token"
goreleaser release --clean
```

This creates:

- Pre-built binaries for all platforms
- Checksums
- Automatic release notes
- Downloadable archives

## 📖 Step 7: Update Your README

Add installation instructions to your README.md:

````markdown
## 🚀 Installation

### Quick Install (requires Go 1.21+)

```bash
# Install both server and client
go install github.com/navpreet032/goTunnel/cmd/...@latest

# Or install individually
go install github.com/navpreet032/goTunnel/cmd/tunnel-server@latest
go install github.com/navpreet032/goTunnel/cmd/tunnel-client@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `$PATH`.

### Download Pre-built Binaries

Download from [Releases](https://github.com/navpreet032/goTunnel/releases):

**Linux (amd64):**

```bash
wget https://github.com/navpreet032/goTunnel/releases/download/v1.0.0/goTunnel_1.0.0_Linux_x86_64.tar.gz
tar -xzf goTunnel_1.0.0_Linux_x86_64.tar.gz
sudo mv tunnel-* /usr/local/bin/
```

**macOS (Apple Silicon):**

```bash
wget https://github.com/navpreet032/goTunnel/releases/download/v1.0.0/goTunnel_1.0.0_Darwin_arm64.tar.gz
tar -xzf goTunnel_1.0.0_Darwin_arm64.tar.gz
sudo mv tunnel-* /usr/local/bin/
```

**Windows:**
Download the `.zip` file and add to your PATH.

### From Source

```bash
git clone https://github.com/navpreet032/goTunnel.git
cd goTunnel
go build -o tunnel-server ./cmd/tunnel-server
go build -o tunnel-client ./cmd/tunnel-client
```
````

## 🎯 Final Checklist

Before publishing:

- [ ] Update module path in `go.mod`
- [ ] Update all imports in `.go` files
- [ ] Rename `cmd/server` → `cmd/tunnel-server`
- [ ] Rename `cmd/client` → `cmd/tunnel-client`
- [ ] Run `go mod tidy`
- [ ] Test builds: `go build ./cmd/...`
- [ ] Create GitHub repository
- [ ] Push code to GitHub
- [ ] Create v1.0.0 tag
- [ ] Create GitHub release
- [ ] (Optional) Set up GoReleaser
- [ ] Update README with installation instructions
- [ ] Test installation: `go install github.com/navpreet032/goTunnel/cmd/...@latest`

## 📊 Usage Statistics

Want to track downloads? GitHub provides:

- Release download counts (for binary releases)
- Go package statistics via [pkg.go.dev](https://pkg.go.dev)

## 🌟 Promoting Your Package

1. **Submit to pkg.go.dev** - Automatic after first `go get`
2. **Add badges** to README:
   ```markdown
   [![Go Reference](https://pkg.go.dev/badge/github.com/navpreet032/goTunnel.svg)](https://pkg.go.dev/github.com/navpreet032/goTunnel)
   [![Go Report Card](https://goreportcard.com/badge/github.com/navpreet032/goTunnel)](https://goreportcard.com/report/github.com/navpreet032/goTunnel)
   ```
3. **Share on** Reddit (r/golang), Twitter, Dev.to
4. **Add topics** on GitHub: `go`, `golang`, `tunnel`, `ngrok`, `networking`

## 🔄 Future Updates

When releasing new versions:

```bash
# Make changes, commit
git add .
git commit -m "feat: add new feature"

# Create new tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# If using GoReleaser
goreleaser release --clean
```

Users update with:

```bash
go install github.com/navpreet032/goTunnel/cmd/...@latest
```

## 📝 Example: Complete Workflow

```bash
# 1. Update module
sed -i '' 's/module goTunnel/module github.com\/johndoe\/goTunnel/' go.mod

# 2. Update imports
find . -type f -name "*.go" -exec sed -i '' 's|goTunnel/|github.com/johndoe/goTunnel/|g' {} +

# 3. Rename directories
mv cmd/server cmd/tunnel-server
mv cmd/client cmd/tunnel-client

# 4. Tidy
go mod tidy

# 5. Test build
go build ./cmd/...

# 6. Commit and push
git add .
git commit -m "chore: prepare for distribution"
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin main
git push origin v1.0.0

# 7. Test installation
go install github.com/johndoe/goTunnel/cmd/...@v1.0.0

# 8. Run it!
tunnel-server --help
tunnel-client --help
```

---

## 🎉 Result

Users worldwide can now install your tunnel service with a single command:

```bash
go install github.com/navpreet032/goTunnel/cmd/...@latest
```

Just like ngrok, but built by you! 🚀
