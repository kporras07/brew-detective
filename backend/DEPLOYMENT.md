# Deployment Guide

This guide explains how to deploy the Brew Detective backend to Google Cloud Run.

## Project Configuration Independence

All deployment scripts include `--project=brew-detective` flags, making them **independent of your current gcloud configuration**. This means:

✅ **Works regardless of**: 
- Current `gcloud config get-value project`
- Active gcloud configuration
- Multiple Google Cloud accounts

✅ **No need to**:
- Run `gcloud config set project brew-detective`
- Switch gcloud configurations
- Worry about wrong project deployments

## Deployment Scripts

### 1. Full Deployment (`deploy.sh`)

The main deployment script that builds, pushes, and deploys:

```bash
# Standard deployment (builds new Docker image)
./deploy.sh

# Force new revision even if code unchanged
./deploy.sh --force

# Skip Docker build, just update deployment manifest
./deploy.sh --skip-build

# Show help
./deploy.sh --help
```

**When to use:**
- After code changes
- First deployment
- When you need to rebuild the Docker image

### 2. Quick Deployment (`quick-deploy.sh`)

Lightweight script that forces a new Cloud Run revision without rebuilding:

```bash
# Force new revision with timestamp annotation
./quick-deploy.sh
```

**When to use:**
- No code changes but need to restart service
- Force deployment when Cloud Run doesn't detect changes
- Quick restart for configuration changes

### 3. Check Deployment (`check-deployment.sh`)

Diagnostic script that shows current deployment status and configuration:

```bash
# Check service status and recent revisions
./check-deployment.sh
```

**When to use:**
- Verify deployment status
- Check recent revisions
- Troubleshoot deployment issues
- Confirm service configuration

## Why Deployments Might Not Update

Cloud Run optimizes deployments by **not creating new revisions** when:

1. **Source code is identical** to the last deployment
2. **Configuration hasn't changed**
3. **Environment variables are the same**

This is actually a **feature** that saves time and resources!

## Solutions for Force Updates

### Method 1: Use the `--force` flag
```bash
./deploy.sh --force
```

### Method 2: Use quick-deploy
```bash
./quick-deploy.sh
```

### Method 3: Manual gcloud command
```bash
gcloud run services update brew-detective-backend \
    --region=us-central1 \
    --project=brew-detective \
    --update-annotations="run.googleapis.com/last-deploy=$(date +%s)"
```

## Build Timestamp Feature

The deployment manifest includes a build timestamp annotation:
```yaml
run.googleapis.com/build-timestamp: "REPLACE_WITH_TIMESTAMP"
```

This gets replaced with the current timestamp during deployment, ensuring each deploy creates a unique revision.

## Troubleshooting

### Issue: "No new revision created"
**Solution:** Use `./deploy.sh --force` or `./quick-deploy.sh`

### Issue: "Service not updating"
**Causes:**
- Identical source code
- Same configuration
- Cached Docker layers

**Solutions:**
1. Force deployment: `./deploy.sh --force`
2. Quick restart: `./quick-deploy.sh`
3. Check logs: `gcloud run services logs read brew-detective-backend --region=us-central1 --project=brew-detective`

### Issue: Build takes too long
**Solution:** Use `./deploy.sh --skip-build` if only config changed

## Environment Variables

Current environment variables (defined in `deploy.yaml`):
- `GOOGLE_CLOUD_PROJECT`
- `FIRESTORE_DATABASE_ID` 
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET` (from secret)
- `JWT_SECRET` (from secret)
- `GOOGLE_REDIRECT_URL`
- `GIN_MODE`
- `FRONTEND_URL`

## Monitoring

Check deployment status:
```bash
# Service status
gcloud run services describe brew-detective-backend --region=us-central1 --project=brew-detective

# Recent revisions
gcloud run revisions list --service=brew-detective-backend --region=us-central1 --project=brew-detective

# Logs
gcloud run services logs read brew-detective-backend --region=us-central1 --project=brew-detective
```

## 📋 Quick Reference

| Scenario | Command | Description |
|----------|---------|-------------|
| Code changed | `./deploy.sh` | Full build and deploy |
| No code change, need update | `./deploy.sh --force` | Force new revision |
| Config only change | `./deploy.sh --skip-build` | Skip Docker build |
| Quick restart | `./quick-deploy.sh` | Fastest option |
| Check status | `./check-deployment.sh` | View deployment info |

## 🎯 All Scripts Include Project ID

Every gcloud command includes `--project=brew-detective`, ensuring:
- **No dependency** on current gcloud configuration
- **Consistent deployments** across different environments  
- **Safe execution** regardless of active gcloud account/project