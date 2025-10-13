# Configuration Analytics pour le Tracking des Liens

## Vue d'ensemble

Ce document décrit la configuration nécessaire pour analyser les données de tracking des liens externes implémentées sur le site tournois-tt.fr.

## Événements Trackés

### 1. Liens Instagram
- **Événement**: `click`
- **Catégorie**: `social_link`
- **Label**: `Instagram Footer`
- **UTM**: `utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=social_instagram&utm_content=footer`

### 2. Liens de Règlements FFTT
- **Événement**: `click`
- **Catégorie**: `external_link`
- **Labels**:
  - `FFTT Rules - Feed List` (liste des tournois)
  - `FFTT Rules - Feed Detail` (page détail tournoi)
  - `Map Tooltip External Link` (tooltip carte)
- **UTM**: `utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=fftt_rules&utm_content=[feed_list|feed_detail|map_tooltip]`

### 3. Liens d'Inscription FFTT
- **Événement**: `click`
- **Catégorie**: `external_link`
- **Label**: `Map Tooltip External Link`
- **UTM**: `utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=fftt_signup&utm_content=map_tooltip`

## Configuration Google Analytics

### 1. Événements Personnalisés

Dans Google Analytics 4, créez les événements personnalisés suivants :

```javascript
// Événement pour les clics sur Instagram
gtag('event', 'social_click', {
  social_network: 'instagram',
  content_group: 'footer',
  source: 'tournois-tt.fr'
});

// Événement pour les clics sur les règlements
gtag('event', 'external_link_click', {
  link_type: 'rules',
  link_source: 'fftt',
  content_group: 'feed_list' // ou 'feed_detail' ou 'map_tooltip'
});

// Événement pour les clics sur les inscriptions
gtag('event', 'external_link_click', {
  link_type: 'signup',
  link_source: 'fftt',
  content_group: 'map_tooltip'
});
```

### 2. Dimensions Personnalisées

Créez les dimensions personnalisées suivantes dans GA4 :

- `link_type` (règlement, inscription, social)
- `link_source` (fftt, instagram)
- `content_group` (footer, feed_list, feed_detail, map_tooltip)
- `tournament_name` (nom du tournoi pour les liens de règlement)

### 3. Métriques Personnalisées

- `external_clicks_total` : Nombre total de clics sur liens externes
- `social_clicks_total` : Nombre total de clics sur liens sociaux
- `rules_downloads_total` : Nombre total de téléchargements de règlements

## Rapports Recommandés

### 1. Rapport des Liens Externes
- **Dimensions**: `link_type`, `content_group`, `tournament_name`
- **Métriques**: `external_clicks_total`, `sessions`
- **Filtre**: `event_name = 'click' AND event_category = 'external_link'`

### 2. Rapport des Liens Sociaux
- **Dimensions**: `social_network`, `content_group`
- **Métriques**: `social_clicks_total`, `sessions`
- **Filtre**: `event_name = 'click' AND event_category = 'social_link'`

### 3. Rapport des Règlements par Tournoi
- **Dimensions**: `tournament_name`, `link_type`
- **Métriques**: `external_clicks_total`
- **Filtre**: `event_name = 'click' AND link_type = 'rules'`

## Configuration UTM

### Paramètres UTM Utilisés

| Paramètre | Valeur | Description |
|-----------|--------|-------------|
| `utm_source` | `tournois-tt.fr` | Source du trafic |
| `utm_medium` | `website` | Medium du trafic |
| `utm_campaign` | `social_instagram`, `fftt_rules`, `fftt_signup` | Campagne |
| `utm_content` | `footer`, `feed_list`, `feed_detail`, `map_tooltip` | Contenu spécifique |

### Exemples d'URLs avec UTM

```
# Instagram
https://instagram.com/tournoistt?utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=social_instagram&utm_content=footer

# Règlement FFTT
https://apiv2.fftt.com/api/files/123456/reglement.pdf?utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=fftt_rules&utm_content=feed_list

# Inscription FFTT
https://helloasso.com/event?utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=fftt_signup&utm_content=map_tooltip
```

## Dashboard Recommandé

### Métriques Clés
1. **Taux de clic sur Instagram** : `social_clicks_total / sessions`
2. **Taux de téléchargement de règlements** : `rules_downloads_total / sessions`
3. **Top 10 des tournois avec le plus de clics sur règlements**
4. **Répartition des clics par source** (feed_list, feed_detail, map_tooltip)

### Alertes
- Augmentation significative des clics sur Instagram
- Tournois avec beaucoup de clics sur règlements mais peu d'inscriptions
- Liens externes avec taux de clic anormalement élevé

## Maintenance

### Vérifications Régulières
1. **Validation des UTM** : Vérifier que tous les liens externes ont bien des paramètres UTM
2. **Test des événements** : Tester le tracking sur tous les types de liens
3. **Nettoyage des données** : Supprimer les événements de test

### Améliorations Futures
1. **A/B Testing** : Tester différents textes de liens pour optimiser le CTR
2. **Heatmaps** : Ajouter des heatmaps pour voir où les utilisateurs cliquent
3. **Funnel Analysis** : Analyser le parcours de l'utilisateur depuis le clic jusqu'à l'inscription

## Code de Test

Pour tester le tracking, utilisez ce code dans la console du navigateur :

```javascript
// Test du tracking Instagram
gtag('event', 'click', {
  event_category: 'social_link',
  event_label: 'Instagram Footer',
  value: 'https://instagram.com/tournoistt'
});

// Test du tracking des règlements
gtag('event', 'click', {
  event_category: 'external_link',
  event_label: 'FFTT Rules - Feed List',
  value: 'https://apiv2.fftt.com/api/files/123456/reglement.pdf',
  link_url: 'https://apiv2.fftt.com/api/files/123456/reglement.pdf',
  link_text: 'Voir le règlement',
  tournament_name: 'Test Tournament'
});
```

## Support

Pour toute question sur la configuration analytics, consultez :
- [Documentation Google Analytics 4](https://developers.google.com/analytics/devguides/collection/ga4)
- [Guide des événements personnalisés GA4](https://support.google.com/analytics/answer/9322688)
