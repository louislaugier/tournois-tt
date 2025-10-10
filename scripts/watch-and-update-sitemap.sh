#!/bin/bash

# Script pour surveiller les changements du fichier data.json et régénérer le sitemap
# À exécuter en arrière-plan pour une mise à jour automatique

set -e

# Configuration
PROJECT_ROOT="/Users/dev/tournois-tt"
API_DATA_FILE="$PROJECT_ROOT/api/cache/data.json"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
LOG_FILE="$PROJECT_ROOT/logs/sitemap-watch.log"

# Créer le répertoire de logs s'il n'existe pas
mkdir -p "$(dirname "$LOG_FILE")"

# Fonction de logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🔍 Démarrage de la surveillance du fichier data.json"

# Vérifier que le fichier existe
if [ ! -f "$API_DATA_FILE" ]; then
    log "❌ Fichier de données non trouvé: $API_DATA_FILE"
    exit 1
fi

# Fonction pour régénérer le sitemap
regenerate_sitemap() {
    log "🔄 Fichier data.json modifié, régénération du sitemap..."
    
    cd "$FRONTEND_DIR"
    
    if npm run generate-sitemap; then
        log "✅ Sitemap régénéré avec succès"
    else
        log "❌ Erreur lors de la régénération du sitemap"
    fi
}

# Obtenir le timestamp initial
last_modified=$(stat -f %m "$API_DATA_FILE" 2>/dev/null || stat -c %Y "$API_DATA_FILE" 2>/dev/null)

log "📊 Surveillance active du fichier: $API_DATA_FILE"
log "⏰ Dernière modification: $(date -r $last_modified)"

# Boucle de surveillance
while true; do
    # Attendre 30 secondes
    sleep 30
    
    # Vérifier la modification du fichier
    current_modified=$(stat -f %m "$API_DATA_FILE" 2>/dev/null || stat -c %Y "$API_DATA_FILE" 2>/dev/null)
    
    if [ "$current_modified" -gt "$last_modified" ]; then
        log "📝 Fichier modifié détecté"
        last_modified=$current_modified
        regenerate_sitemap
    fi
done
