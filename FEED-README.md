# Feed HTML Statique - Tournois de Tennis de Table

Ce système génère automatiquement des pages HTML statiques pour chaque tournoi, rendant le contenu indexable par les moteurs de recherche.

## 🎯 Objectif

Créer un "shadow site" ou "feed HTML" statique qui permet aux crawlers d'indexer les données des tournois de tennis de table. Chaque tournoi dispose d'une page HTML dédiée avec toutes les informations essentielles.

## 📁 Structure

```
/feed/                    # Index des tournois
/feed/123.html           # Page individuelle du tournoi ID 123
/feed/456.html           # Page individuelle du tournoi ID 456
/feed/sitemap.xml        # Sitemap pour les moteurs de recherche
```

## 🚀 Utilisation

### Génération manuelle

```bash
# Depuis le répertoire frontend
cd frontend
npm run generate-feed
```

### Génération automatique (recommandé)

```bash
# Exécuter le script de mise à jour
./scripts/update-feed.sh

# Configurer le cron pour automatisation
crontab scripts/crontab-example.txt
```

## 📊 Données incluses

Chaque page de tournoi contient :

- **Nom du tournoi**
- **Type** (International, National A/B, Régional, Départemental, Promotionnel)
- **Club organisateur**
- **Dotation** (si applicable)
- **Dates** (début et fin)
- **Adresse complète**
- **Lien vers le règlement** (PDF)
- **Lien d'inscription** (si disponible)

## 🔍 SEO et Indexation

### Métadonnées incluses

- **Meta tags** : title, description, keywords
- **Open Graph** : pour les réseaux sociaux
- **Twitter Cards** : pour Twitter
- **JSON-LD** : données structurées Schema.org
- **Sitemap XML** : pour les moteurs de recherche
- **Robots.txt** : directives pour les crawlers

### Sitemap XML

Le système génère deux sitemaps :

1. **Sitemap principal** (`/sitemap.xml`) :
   - Pages principales du site (accueil, cookies)
   - Index des tournois (`/feed/`)
   - Toutes les pages individuelles de tournois (`/feed/:id`)
   - **Total : 684 URLs** (3 pages principales + 681 tournois)

2. **Sitemap du feed** (`/feed/sitemap.xml`) :
   - Index des tournois (`/feed/`)
   - Pages individuelles de tournois uniquement
   - **Total : 682 URLs** (1 index + 681 tournois)

### Exemple de données structurées

```json
{
  "@context": "https://schema.org",
  "@type": "SportsEvent",
  "name": "TOURNOI REGIONAL ES ORMES TT",
  "startDate": "2025-06-01T00:00:00",
  "endDate": "2025-06-01T00:00:00",
  "location": {
    "@type": "Place",
    "address": {
      "@type": "PostalAddress",
      "streetAddress": "Chemin des Plantes",
      "postalCode": "45140",
      "addressLocality": "ORMES"
    }
  },
  "organizer": {
    "@type": "Organization",
    "name": "ORMES EVEIL SPORTIF"
  },
  "sport": "Tennis de Table"
}
```

## ⚙️ Configuration

### Scripts npm disponibles

```json
{
  "generate-feed": "node scripts/generate-static-feed.js",
  "generate-sitemap": "node scripts/generate-sitemap.js",
  "build:feed": "npm run build && npm run generate-feed && npm run generate-sitemap"
}
```

### Configuration cron

Le fichier `scripts/crontab-example.txt` contient des exemples de configuration cron :

- **Toutes les 6 heures** : `0 */6 * * *`
- **Tous les jours à 2h** : `0 2 * * *`
- **Toutes les 2 heures** : `0 */2 * * *`

## 📝 Logs

Les logs sont stockés dans `logs/feed-update.log` et incluent :

- Timestamp de chaque opération
- Nombre de tournois traités
- Nombre de fichiers générés
- Taille du répertoire de sortie
- Erreurs éventuelles

## 🔧 Développement

### Structure des fichiers

```
frontend/
├── scripts/
│   └── generate-static-feed.js    # Script principal de génération
├── src/
│   ├── pages/
│   │   ├── FeedIndex.tsx          # Page d'index des tournois
│   │   └── FeedTournament.tsx     # Page individuelle de tournoi
│   ├── components/
│   │   └── FeedMeta.tsx           # Composant pour les métadonnées SEO
│   └── lib/
│       └── hooks/
│           └── useTournaments.ts  # Hook pour charger les données
└── build/
    └── feed/                      # Pages HTML générées
```

### Ajout de nouvelles données

Pour ajouter de nouvelles informations aux pages :

1. Modifier le script `generate-static-feed.js`
2. Mettre à jour les templates HTML
3. Ajouter les métadonnées SEO correspondantes
4. Tester avec `npm run generate-feed`

## 🚀 Déploiement

### Intégration avec le build

Le feed peut être généré automatiquement lors du build :

```bash
npm run build:feed
```

### Serveur web

Les fichiers générés dans `build/feed/` peuvent être servis directement par n'importe quel serveur web statique (Nginx, Apache, etc.).

### Configuration Nginx

```nginx
location /feed/ {
    alias /path/to/build/feed/;
    try_files $uri $uri/ =404;
    
    # Headers pour le SEO
    add_header X-Robots-Tag "index, follow";
    add_header Cache-Control "public, max-age=3600";
}
```

## 📈 Monitoring

### Vérification de l'indexation

- **Google Search Console** : Vérifier l'indexation des pages
- **Bing Webmaster Tools** : Suivre l'indexation Bing
- **Sitemap principal** : Soumettre `/sitemap.xml` aux moteurs (inclut toutes les pages)
- **Sitemap du feed** : `/feed/sitemap.xml` (pages de tournois uniquement)

### Mise à jour automatique du sitemap

Le sitemap peut être mis à jour de plusieurs façons :

#### Option 1: Surveillance automatique (recommandé)
```bash
# Démarrer la surveillance en arrière-plan
nohup ./scripts/watch-and-update-sitemap.sh &
```

#### Option 2: Mise à jour manuelle
```bash
# Mettre à jour le sitemap manuellement
./scripts/update-sitemap.sh
```

#### Option 3: Via cron (comme actuellement)
```bash
# Configuration cron existante
crontab scripts/crontab-example.txt
```

### Soumission du sitemap

```bash
# Soumettre le sitemap aux moteurs de recherche
./scripts/submit-sitemap.sh
```

### Métriques à surveiller

- Nombre de pages indexées
- Temps de génération
- Taille du répertoire de sortie
- Erreurs dans les logs

## 🐛 Dépannage

### Problèmes courants

1. **Données manquantes** : Vérifier que `api/cache/data.json` existe et est à jour
2. **Permissions** : S'assurer que le script a les droits d'écriture
3. **Node.js** : Vérifier la version (>= 20.18.1)
4. **Dépendances** : Exécuter `npm install` si nécessaire

### Logs de débogage

```bash
# Voir les logs récents
tail -f logs/feed-update.log

# Vérifier les tâches cron
crontab -l

# Tester la génération manuellement
cd frontend && npm run generate-feed
```

## 🔄 Maintenance

### Nettoyage automatique

Le script nettoie automatiquement les logs de plus de 30 jours.

### Mise à jour des données

Les données sont automatiquement mises à jour via le processus cron, mais peuvent aussi être rafraîchies manuellement en relançant le script de l'API.

## 📞 Support

Pour toute question ou problème, consulter :

1. Les logs dans `logs/feed-update.log`
2. La documentation du script `scripts/generate-static-feed.js`
3. Les exemples de configuration dans `scripts/crontab-example.txt`
