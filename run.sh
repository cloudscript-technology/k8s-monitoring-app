#!/bin/bash

# K8s Monitoring App - Run Script
# This script ensures the application runs from the correct directory

set -e

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Change to project root
cd "$SCRIPT_DIR"

echo "🚀 Starting K8s Monitoring App..."
echo "📁 Working directory: $SCRIPT_DIR"
echo ""

# Check if web templates exist
if [ ! -d "web/templates" ]; then
    echo "❌ Error: web/templates directory not found!"
    echo "   Make sure you're running this script from the project root."
    exit 1
fi

# Check environment variables
if [ -z "$DB_HOST" ]; then
    echo "⚠️  DB_HOST not set, using default: localhost"
    export DB_HOST="localhost"
fi

if [ -z "$DB_PORT" ]; then
    echo "⚠️  DB_PORT not set, using default: 5432"
    export DB_PORT="5432"
fi

if [ -z "$DB_USER" ]; then
    echo "⚠️  DB_USER not set, using default: monitoring"
    export DB_USER="monitoring"
fi

if [ -z "$DB_PASSWORD" ]; then
    echo "⚠️  DB_PASSWORD not set, using default: monitoring"
    export DB_PASSWORD="monitoring"
fi

if [ -z "$DB_NAME" ]; then
    echo "⚠️  DB_NAME not set, using default: k8s_monitoring"
    export DB_NAME="k8s_monitoring"
fi

if [ -z "$METRICS_RETENTION_DAYS" ]; then
    echo "⚠️  METRICS_RETENTION_DAYS not set, using default: 30 days"
    export METRICS_RETENTION_DAYS="30"
fi

if [ -z "$METRICS_CLEANUP_INTERVAL" ]; then
    echo "⚠️  METRICS_CLEANUP_INTERVAL not set, using default: daily at 2 AM"
    export METRICS_CLEANUP_INTERVAL="0 2 * * *"
fi

if [ -z "$KUBECONFIG" ]; then
    if [ -f "$HOME/.kube/config" ]; then
        echo "✅ Using kubeconfig: $HOME/.kube/config"
        export KUBECONFIG="$HOME/.kube/config"
    else
        echo "⚠️  KUBECONFIG not set and ~/.kube/config not found"
        echo "   The app will try to use in-cluster configuration"
    fi
fi

echo ""
echo "📋 Configuration:"
echo "   DB_HOST=$DB_HOST"
echo "   DB_PORT=$DB_PORT"
echo "   DB_USER=$DB_USER"
echo "   DB_NAME=$DB_NAME"
echo "   METRICS_RETENTION_DAYS=$METRICS_RETENTION_DAYS"
echo "   METRICS_CLEANUP_INTERVAL=$METRICS_CLEANUP_INTERVAL"
echo ""
echo "🌐 Web UI will be available at: http://localhost:8080"
echo "📊 REST API will be available at: http://localhost:8080/api/v1"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Run the application
go run cmd/main.go

