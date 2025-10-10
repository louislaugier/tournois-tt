#!/bin/bash

# Script pour mettre à jour le feed HTML statique
# À exécuter via cron pour maintenir les pages à jour

set -e

# Configuration
PROJECT_ROOT="/Users/dev/tournois-tt"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
API_DIR="$PROJECT_ROOT/api"
LOG_FILE="$PROJECT_ROOT/logs/feed-update.log"

# Créer le répertoire de logs s'il n'existe pas
mkdir -p "$(dirname "$LOG_FILE")"

# Fonction de logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🚀 Début de la mise à jour du feed HTML"

# Vérifier que nous sommes dans le bon répertoire
if [ ! -d "$PROJECT_ROOT" ]; then
    log "❌ Erreur: Répertoire du projet non trouvé: $PROJECT_ROOT"
    exit 1
fi

# Aller dans le répertoire du projet
cd "$PROJECT_ROOT"

# Vérifier si l'API a des données récentes (moins de 24h)
if [ -f "$API_DIR/cache/data.json" ]; then
    DATA_AGE=$(find "$API_DIR/cache/data.json" -mtime -1 | wc -l)
    if [ "$DATA_AGE" -eq 0 ]; then
        log "⚠️  Les données de l'API sont anciennes (plus de 24h). Vérifiez le processus de mise à jour."
    else
        log "✅ Données de l'API récentes trouvées"
    fi
else
    log "❌ Fichier de données de l'API non trouvé: $API_DIR/cache/data.json"
    exit 1
fi

# Aller dans le répertoire frontend
cd "$FRONTEND_DIR"

# Vérifier que Node.js est disponible
if ! command -v node &> /dev/null; then
    log "❌ Erreur: Node.js n'est pas installé ou pas dans le PATH"
    exit 1
fi

# Vérifier que npm est disponible
if ! command -v npm &> /dev/null; then
    log "❌ Erreur: npm n'est pas installé ou pas dans le PATH"
    exit 1
fi

# Installer les dépendances si nécessaire
if [ ! -d "node_modules" ]; then
    log "📦 Installation des dépendances npm..."
    npm install
fi

# Générer le feed HTML statique et le sitemap
log "🔨 Génération du feed HTML statique et du sitemap..."
if npm run build:feed; then
    log "✅ Feed HTML et sitemap générés avec succès"
    
    # Compter les fichiers générés
    FEED_DIR="$FRONTEND_DIR/build/feed"
    if [ -d "$FEED_DIR" ]; then
        HTML_COUNT=$(find "$FEED_DIR" -name "*.html" | wc -l)
        log "📊 $HTML_COUNT fichiers HTML générés dans $FEED_DIR"
    fi
    
    # Vérifier la taille du répertoire
    if [ -d "$FEED_DIR" ]; then
        FEED_SIZE=$(du -sh "$FEED_DIR" | cut -f1)
        log "💾 Taille du feed: $FEED_SIZE"
    fi
    
    # Vérifier le sitemap principal
    SITEMAP_PATH="$FRONTEND_DIR/public/sitemap.xml"
    if [ -f "$SITEMAP_PATH" ]; then
        SITEMAP_SIZE=$(du -sh "$SITEMAP_PATH" | cut -f1)
        log "🗺️  Sitemap principal généré: $SITEMAP_SIZE"
    fi
    
else
    log "❌ Erreur lors de la génération du feed HTML"
    exit 1
fi

log "🎉 Mise à jour du feed terminée avec succès"

# Nettoyer les anciens logs (garder seulement les 30 derniers jours)
find "$(dirname "$LOG_FILE")" -name "*.log" -mtime +30 -delete 2>/dev/null || true

log "🧹 Nettoyage des anciens logs terminé"
