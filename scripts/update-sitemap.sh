#!/bin/bash

# Script simple pour mettre à jour le sitemap
# Peut être appelé depuis l'API Go après chaque mise à jour des données

set -e

# Configuration
PROJECT_ROOT="/Users/dev/tournois-tt"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
LOG_FILE="$PROJECT_ROOT/logs/sitemap-update.log"

# Créer le répertoire de logs s'il n'existe pas
mkdir -p "$(dirname "$LOG_FILE")"

# Fonction de logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🔄 Mise à jour du sitemap et du RSS demandée"

# Aller dans le répertoire frontend
cd "$FRONTEND_DIR"

# Générer le sitemap et le RSS
if npm run generate-sitemap && npm run generate-rss; then
    log "✅ Sitemap et RSS mis à jour avec succès"
else
    log "❌ Erreur lors de la mise à jour du sitemap ou du RSS"
    exit 1
fi
