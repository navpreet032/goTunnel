# 📋 Publishing Checklist for GoTunnel

Quick reference guide to make your tunnel service a downloadable Go package.

## 🎯 What You'll Achieve

Users will be able to install your tunnel service with:
```bash
go install github.com/YOUR_USERNAME/goTunnel/cmd/...@latest
```

Just like ngrok, but your own! 🚀

## ✅ Pre-Publishing Checklist

### 1. Update Module Path

**File:** `go.mod`

```diff
- module goTunnel
+ module github.com/YOUR_USERNAME/goTunnel
```

### 2. Update All Imports

Run this command to update all Go files:

```bash
find . -type f -name "*.go" -exec sed -i '' 's|goTunnel/|github.com/YOUR_USERNAME/goTunnel/|g' {} +
```

**Manually verify these files:**
- `cmd/server/main.go`
- `cmd/client/main.go`
- `pkg/server/http_handler.go`
- `pkg/client/proxy.go`
- All files in `pkg/*`

### 3. Rename Command Directories (Optional but Recommended)

For better user experience:

```bash
mv cmd/server cmd/tunnel-server
mv cmd/client cmd/tunnel-client
```

Update `scripts/start-server.sh` and `scripts/start-client.sh` if needed.

### 4. Test Build

```bash
go mod tidy
go build ./cmd/server
go build ./cmd/client

# Test both binaries
./server --help
./client --help
```

### 5. Update Documentation

**Files to update:**
- [x] `README.md` - Already updated with installation instructions
- [x] `.goreleaser.yaml` - Already created
- [ ] Replace `YOUR_USERNAME` with your actual GitHub username in:
  - `README.md`
  - `.goreleaser.yaml`
  - `DISTRIBUTION.md`

### 6. Create GitHub Repository

1. Go to https://github.com/new
2. Repository name: `goTunnel` (or `go-tunnel`)
3. Description: "A lightweight tunnel service written in Go - expose localhost to the internet"
4. Public or Private: **Public** (for go install to work)
5. **Don't** initialize with README (you already have one)

### 7. Push to GitHub

```bash
# Initialize git (if not already)
git init

# Add remote
git remote add origin https://github.com/YOUR_USERNAME/goTunnel.git

# Stage all files
git add .

# Commit
git commit -m "Initial release: GoTunnel v1.0.0"

# Push
git push -u origin main
```

### 8. Create First Release

```bash
# Create tag
git tag -a v1.0.0 -m "Release v1.0.0: Initial stable release"

# Push tag
git push origin v1.0.0
```

**On GitHub:**
1. Go to your repository
2. Click "Releases" → "Create a new release"
3. Choose tag: `v1.0.0`
4. Release title: `v1.0.0 - Initial Release`
5. Description:
   ```markdown
   ## GoTunnel v1.0.0 🚀

   A lightweight, production-ready tunnel service written in Go.

   ### Features
   - ✅ WebSocket-based HTTP tunneling
   - ✅ Auto-reconnect with exponential backoff
   - ✅ Multiple concurrent clients
   - ✅ Production-ready logging
   - ✅ Configurable via CLI or env vars

   ### Installation
   \`\`\`bash
   go install github.com/YOUR_USERNAME/goTunnel/cmd/...@v1.0.0
   \`\`\`

   ### Quick Start
   \`\`\`bash
   # Start server (on your VPS)
   tunnel-server --port 8080 --domain tunnel.example.com

   # Start client (on your local machine)
   tunnel-client --local-port 3000
   \`\`\`

   See [README](https://github.com/YOUR_USERNAME/goTunnel#readme) for full documentation.
   ```
6. Click "Publish release"

### 9. Test Installation

On a different machine (or remove local version):

```bash
go install github.com/YOUR_USERNAME/goTunnel/cmd/...@v1.0.0

# Verify installation
which tunnel-server
which tunnel-client

# Test commands
tunnel-server --help
tunnel-client --help
```

## 🎁 Optional: Multi-Platform Binaries with GoReleaser

For pre-built binaries (Linux, macOS, Windows):

### Install GoReleaser

```bash
# macOS
brew install goreleaser

# Or download from https://goreleaser.com/install/
```

### Test Locally

```bash
goreleaser release --snapshot --clean
```

Check `dist/` folder for binaries.

### Create Release

```bash
# Get GitHub token from: https://github.com/settings/tokens
# Needs: repo (all), write:packages
export GITHUB_TOKEN="your_token_here"

# Create and push tag
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin v1.0.1

# Build and release
goreleaser release --clean
```

This automatically:
- ✅ Builds for Linux, macOS, Windows (amd64, arm64)
- ✅ Creates archives (.tar.gz, .zip)
- ✅ Generates checksums
- ✅ Uploads to GitHub Releases
- ✅ Creates release notes

## 📊 After Publishing

### Track Your Package

Your package will automatically appear on:
- **pkg.go.dev**: https://pkg.go.dev/github.com/YOUR_USERNAME/goTunnel
- **Go Package Search**: Users can discover it

### Add Badges to README

```markdown
[![Go Reference](https://pkg.go.dev/badge/github.com/YOUR_USERNAME/goTunnel.svg)](https://pkg.go.dev/github.com/YOUR_USERNAME/goTunnel)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USERNAME/goTunnel)](https://goreportcard.com/report/github.com/YOUR_USERNAME/goTunnel)
[![GitHub release](https://img.shields.io/github/v/release/YOUR_USERNAME/goTunnel.svg)](https://github.com/YOUR_USERNAME/goTunnel/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

### Add GitHub Topics

On your GitHub repository page:
- Click ⚙️ (Settings icon) next to "About"
- Add topics: `go`, `golang`, `tunnel`, `ngrok`, `networking`, `websocket`, `reverse-proxy`

### Share Your Project

- Reddit: r/golang
- Twitter/X: #golang #webdev
- Dev.to: Write a blog post
- Hacker News: Show HN

## 🔄 Future Releases

When you add new features:

```bash
# Make changes
git add .
git commit -m "feat: add custom client IDs"

# Create new version
git tag -a v1.1.0 -m "Release v1.1.0: Custom client IDs"
git push origin v1.1.0

# If using GoReleaser
goreleaser release --clean
```

Users update with:
```bash
go install github.com/YOUR_USERNAME/goTunnel/cmd/...@latest
```

## 📝 Version Numbering (Semantic Versioning)

- `v1.0.0` → `v1.0.1`: Bug fixes, patches
- `v1.0.0` → `v1.1.0`: New features (backward compatible)
- `v1.0.0` → `v2.0.0`: Breaking changes

## 🎉 Success Criteria

Your package is successfully published when:

- [x] Module path uses GitHub URL
- [x] Code pushed to public GitHub repository
- [x] v1.0.0 tag created
- [x] GitHub release published
- [x] `go install github.com/YOUR_USERNAME/goTunnel/cmd/...@latest` works
- [x] Package appears on pkg.go.dev
- [x] Users can download and run your binaries

## 🆘 Need Help?

See [DISTRIBUTION.md](DISTRIBUTION.md) for detailed explanations of each step.

## 📞 Questions?

Common issues:
- **"go: module not found"** → Make sure repository is public
- **"no Go files"** → Check module path in go.mod matches GitHub URL
- **"command not found"** → Add `$GOPATH/bin` to your `$PATH`

---

**Remember:** Once you push to GitHub with the correct module path, anyone in the world can install your tunnel service! 🌍🚀
