# Deployment Guide

## Production Deployment

### 1. Create directory on server:
```bash
mkdir -p ~/servers-stats
cd ~/servers-stats
```

### 2. Download files:
```bash
# Download docker-compose
wget https://raw.githubusercontent.com/koss-shtukert/servers-stats/refs/heads/master/docker-compose.yml

# Download configuration template
wget https://raw.githubusercontent.com/koss-shtukert/servers-stats/refs/heads/master/config.example.yaml -O config.yaml
```

### 3. Edit config.yaml:
```bash
editor config.yaml
# Change:
# - tgbot_api_key: "YOUR_BOT_TOKEN_HERE"
# - tgbot_chat_id: "YOUR_CHAT_ID_HERE"
# - Configure cron intervals as needed
```

### 4. Create logs directory:
```bash
mkdir -p logs
```

### 5. Start:
```bash
docker-compose up -d
```

### 6. Check logs:
```bash
docker-compose logs -f
```

### 7. Update:
```bash
docker-compose pull
docker-compose up -d
```

## Local Development

### 1. Clone repository:
```bash
git clone https://github.com/koss-shtukert/servers-stats.git
cd servers-stats
```

### 2. Copy config template:
```bash
cp config.example.yaml config.yaml
# Edit values in config.yaml
```

### 3. Run locally:
```bash
# With Docker build
docker-compose up -d

# Or run Go directly
go run main.go
```
