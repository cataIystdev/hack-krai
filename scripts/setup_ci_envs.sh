#!/bin/bash
# ==============================================================================
# Deep Krai — Server Environment Setup Script for CI/CD
# Run this script ONCE on the target server (141.98.7.225) as root.
# ==============================================================================

set -e

echo "Starting Deep Krai Server Setup..."

CONFIG_DIR="/opt/deepkrai/configs"
mkdir -p "$CONFIG_DIR"

declare -a ENVS=("prod" "dev" "test-catalyst" "test-lowcoware")

# Base port offsets for each environment (e.g. prod=0, dev=100, catalyst=200, lowcoware=300)
declare -A PORT_OFFSETS
PORT_OFFSETS=( ["prod"]=0 ["dev"]=100 ["test-catalyst"]=200 ["test-lowcoware"]=300 )

# Create base .env files
for ENV in "${ENVS[@]}"; do
    OFFSET=${PORT_OFFSETS[$ENV]}
    ENV_FILE="$CONFIG_DIR/.env.$ENV"
    
    echo "Generating $ENV_FILE with port offset +$OFFSET"
    
    cat <<EOF > "$ENV_FILE"
# Environment: $ENV

APP_NAME=deep-krai-api-$ENV
APP_ENV=$ENV
APP_PORT=$((8080 + OFFSET))
APP_LOG_LEVEL=debug

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=$((5432 + OFFSET))
POSTGRES_USER=deepkrai
POSTGRES_PASSWORD=deepkrai_secret_2026
POSTGRES_DB=deepkrai
POSTGRES_MAX_CONNS=20
POSTGRES_MIN_CONNS=5

# Redis
REDIS_HOST=localhost
REDIS_PORT=$((6379 + OFFSET))
REDIS_PASSWORD=redis_secret_2026
REDIS_DB=0

# Qdrant
QDRANT_HOST=localhost
QDRANT_HTTP_PORT=$((6333 + OFFSET))
QDRANT_GRPC_PORT=$((6334 + OFFSET))

# Neo4j
NEO4J_HOST=localhost
NEO4J_BOLT_PORT=$((7687 + OFFSET))
NEO4J_HTTP_PORT=$((7474 + OFFSET))
NEO4J_USER=neo4j
NEO4J_PASSWORD=neo4j_secret_2026

# ClickHouse
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=$((9000 + OFFSET))
CLICKHOUSE_HTTP_PORT=$((8123 + OFFSET))
CLICKHOUSE_USER=deepkrai
CLICKHOUSE_PASSWORD=clickhouse_secret_2026
CLICKHOUSE_DB=deepkrai_analytics

# MinIO
MINIO_HOST=localhost
MINIO_API_PORT=$((9002 + OFFSET))
MINIO_CONSOLE_PORT=$((9001 + OFFSET))
MINIO_ROOT_USER=deepkrai_minio
MINIO_ROOT_PASSWORD=minio_secret_2026
MINIO_BUCKET=deepkrai-media
MINIO_USE_SSL=false

# JWT
JWT_SECRET=deep-krai-super-secret-key-2026-change-${ENV}
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168
JWT_ISSUER=deep-krai-api-$ENV

# AI
AI_BASE_URL=https://api.onlysq.ru/ai/openai/
AI_API_KEY=sq-P9j2PA2fUwFrjyfCKPdQ8SaecHSGrY22
AI_WHISPER_MODEL=whisper-1
AI_LLM_MODEL=gpt-5.1
AI_EMBEDDINGS_MODEL=local-minilm
EMBEDDINGS_PORT=$((7997 + OFFSET))
AI_EMBEDDINGS_BASE_URL=http://localhost:$((7997 + OFFSET))/

# Vosk STT
VOSK_HOST=127.0.0.1
VOSK_PORT=$((2700 + OFFSET))
EOF

    # Check if we are running as root on the actual server before creating systemd services
    if [ "$EUID" -eq 0 ]; then
        mkdir -p "/opt/deepkrai/$ENV"
    fi
done

# Create Systemd Template Service if running as root
if [ "$EUID" -eq 0 ]; then
    echo "Installing systemd template service deepkrai-api@.service..."
    cat <<'EOF' > /etc/systemd/system/deepkrai-api@.service
[Unit]
Description=Deep Krai API (%i environment)
After=network.target docker.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/deepkrai/%i/api
ExecStart=/opt/deepkrai/%i/api/api_bin
Restart=always
RestartSec=5
LimitNOFILE=65536
EnvironmentFile=/opt/deepkrai/configs/.env.%i

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    echo "✅ Systemd template loaded. You can now use: systemctl restart deepkrai-api@dev"
else
    echo "⚠️ Not running as root. Skipping systemd service template creation and final directory creation."
    echo "Please copy the .env files to /opt/deepkrai/configs on your target server and create the systemd template manually."
fi

echo "✅ Setup script completed successfully."
