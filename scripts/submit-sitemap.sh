#!/bin/bash

# Script pour soumettre le sitemap aux moteurs de recherche
# À exécuter après la génération du sitemap

set -e

# Configuration
SITEMAP_URL="https://tournois-tt.fr/sitemap.xml"
LOG_FILE="/Users/dev/tournois-tt/logs/sitemap-submission.log"

# Créer le répertoire de logs s'il n'existe pas
mkdir -p "$(dirname "$LOG_FILE")"

# Fonction de logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🚀 Soumission du sitemap aux moteurs de recherche"

# Vérifier que curl est disponible
if ! command -v curl &> /dev/null; then
    log "❌ Erreur: curl n'est pas installé"
    exit 1
fi

# Soumettre à Google Search Console (si l'API key est configurée)
if [ -n "$GOOGLE_SEARCH_CONSOLE_API_KEY" ]; then
    log "📤 Soumission à Google Search Console..."
    # Note: L'API Google Search Console nécessite une authentification OAuth2
    # Cette partie nécessiterait une implémentation plus complexe
    log "⚠️  Soumission Google: Configuration manuelle requise"
else
    log "ℹ️  Google Search Console: Pas de clé API configurée"
fi

# Soumettre à Bing Webmaster Tools (si l'API key est configurée)
if [ -n "$BING_API_KEY" ]; then
    log "📤 Soumission à Bing Webmaster Tools..."
    # Note: L'API Bing nécessite aussi une authentification
    log "⚠️  Soumission Bing: Configuration manuelle requise"
else
    log "ℹ️  Bing Webmaster Tools: Pas de clé API configurée"
fi

# Vérifier que le sitemap est accessible
log "🔍 Vérification de l'accessibilité du sitemap..."
if curl -s --head "$SITEMAP_URL" | grep -q "200 OK"; then
    log "✅ Sitemap accessible: $SITEMAP_URL"
else
    log "❌ Sitemap non accessible: $SITEMAP_URL"
    exit 1
fi

# Instructions pour la soumission manuelle
log "📋 Instructions pour la soumission manuelle:"
log "   Google Search Console: https://search.google.com/search-console"
log "   - Ajouter la propriété: $SITEMAP_URL"
log "   - Soumettre le sitemap: $SITEMAP_URL"
log ""
log "   Bing Webmaster Tools: https://www.bing.com/webmasters"
log "   - Ajouter le site: https://tournois-tt.fr"
log "   - Soumettre le sitemap: $SITEMAP_URL"
log ""
log "   Yandex Webmaster: https://webmaster.yandex.com"
log "   - Ajouter le site: https://tournois-tt.fr"
log "   - Soumettre le sitemap: $SITEMAP_URL"

log "🎉 Soumission du sitemap terminée"

# Nettoyer les anciens logs (garder seulement les 30 derniers jours)
find "$(dirname "$LOG_FILE")" -name "*.log" -mtime +30 -delete 2>/dev/null || true
