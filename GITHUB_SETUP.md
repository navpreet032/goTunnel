# GitHub Repository Setup - Complete! ✅

## What's Ready

✅ Git repository initialized  
✅ `.gitignore` created (excludes binaries, IDE files, artifacts)  
✅ `README.md` with full documentation  
✅ Start scripts for server and client  
✅ `DEPLOYMENT.md` with EC2 setup guide  
✅ All code committed to git

## Next Steps for GitHub

### 1. Create GitHub Repository

Go to https://github.com/new and create a new repository named `goTunnel`

### 2. Push to GitHub

```bash
cd /Users/admin/Desktop/goTunnel

# Add your GitHub repository as remote
git remote add origin https://github.com/navpreet032/goTunnel.git

# Push to GitHub
git branch -M main
git push -u origin main
```

### 3. Clone on Your Server

```bash
# SSH into your EC2 server, then:
git clone https://github.com/navpreet032/goTunnel.git
cd goTunnel
go mod download
./scripts/start-server.sh 8080
```

## Quick Start Scripts

### On Server (EC2)

```bash
./scripts/start-server.sh                    # Default port 8080
./scripts/start-server.sh 9090               # Custom port
./scripts/start-server.sh 8080 example.com   # With domain
```

### On Local Machine

```bash
# Start your local app first (e.g., on port 3000)
npm run dev

# Then start tunnel client
./scripts/start-client.sh ws://YOUR-EC2-IP:8080/tunnel 3000
```

## Testing Your Setup

1. **Start local web app** (e.g., Vite on port 5173)
2. **Start server** on EC2
3. **Start client** pointing to server
4. **Access**: `http://YOUR-EC2-IP:8080/client-1/`

See [DEPLOYMENT.md](file:///Users/admin/Desktop/goTunnel/DEPLOYMENT.md) for detailed instructions!
