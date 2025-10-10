#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

/**
 * Script pour générer un flux RSS des tournois
 * Utilise exactement les mêmes données que le feed HTML statique
 */

const API_DATA_PATH = process.env.OUTPUT_DIR === '/usr/share/nginx/html' 
  ? '/app/api/cache/data.json'  // Production Docker
  : path.join(__dirname, '../../api/cache/data.json'); // Local development
const OUTPUT_DIR = process.env.OUTPUT_DIR || path.join(__dirname, '../public');
const RSS_PATH = path.join(OUTPUT_DIR, 'rss.xml');
const BUILD_RSS_PATH = path.join(__dirname, '../build/rss.xml');

// Ensure output directory exists
if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

// Fonction pour formater les dates au format RSS
function formatRSSDate(dateString) {
  const date = new Date(dateString);
  return date.toUTCString();
}

// Fonction pour mapper les types de tournois
function mapTournamentType(type) {
  const types = {
    'I': 'International',
    'A': 'National A',
    'B': 'National B',
    'R': 'Régional',
    'D': 'Départemental',
    'P': 'Promotionnel'
  };
  return types[type] || type;
}

// Fonction pour formater les dates en français
function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
}

// Fonction pour générer le flux RSS
function generateRSS(tournaments) {
  const now = new Date();
  const buildDate = formatRSSDate(now.toISOString());
  
  // Trier les tournois par date de début (plus récents en premier)
  const sortedTournaments = [...tournaments].sort((a, b) => {
    return new Date(b.startDate) - new Date(a.startDate);
  });

  // Prendre tous les tournois pour le RSS
  const recentTournaments = sortedTournaments;

  const items = recentTournaments.map(tournament => {
    // Utiliser exactement les mêmes données que le feed HTML
    const title = `${tournament.name} - Tournoi ${mapTournamentType(tournament.type)} | FFTT`;
    
    // Construire la description avec le lien du règlement si disponible
    let description = `Tournoi de tennis de table ${mapTournamentType(tournament.type)} organisé par ${tournament.club.name} le ${formatDate(tournament.startDate)}${tournament.startDate !== tournament.endDate ? ` au ${formatDate(tournament.endDate)}` : ''} à ${tournament.address.addressLocality}. ${tournament.endowment > 0 ? `Dotation: ${tournament.endowment.toLocaleString('fr-FR')}€.` : ''} Informations pratiques`;
    
    if (tournament.rules && tournament.rules.url) {
      description += `, <a href="${tournament.rules.url}">règlement (PDF)</a>`;
    } else {
      description += ', règlement';
    }
    
    description += ' et inscription.';

    const itemDate = formatRSSDate(tournament.startDate);
    const guid = `tournoi-${tournament.id}`;

    return `    <item>
      <title><![CDATA[${title}]]></title>
      <link>https://tournois-tt.fr/feed/${tournament.id}</link>
      <guid isPermaLink="false">${guid}</guid>
      <pubDate>${itemDate}</pubDate>
      <description><![CDATA[${description}]]></description>
      <category><![CDATA[${mapTournamentType(tournament.type)}]]></category>
      <category><![CDATA[${tournament.club.name}]]></category>
      <category><![CDATA[${tournament.address.addressLocality}]]></category>
      <category><![CDATA[${tournament.address.postalCode}]]></category>
      ${tournament.endowment > 0 ? `<category><![CDATA[Dotation: ${tournament.endowment.toLocaleString('fr-FR')}€]]></category>` : ''}
      ${tournament.rules && tournament.rules.url ? `<category><![CDATA[Règlement: ${tournament.rules.url}]]></category>` : ''}
    </item>`;
  }).join('\n');

  return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Tournois de Tennis de Table FFTT</title>
    <link>https://tournois-tt.fr</link>
    <description>Carte interactive des tournois de tennis de table en France. Trouvez les prochains tournois FFTT près de chez vous. Tri & recherche par type de tournoi, club organisateur, date, région, code postal, ville, dotation.</description>
    <language>fr-FR</language>
    <lastBuildDate>${buildDate}</lastBuildDate>
    <pubDate>${buildDate}</pubDate>
    <ttl>60</ttl>
    <atom:link href="https://tournois-tt.fr/rss.xml" rel="self" type="application/rss+xml"/>
    <managingEditor>noreply@tournois-tt.fr (Tournois FFTT)</managingEditor>
    <webMaster>noreply@tournois-tt.fr (Tournois FFTT)</webMaster>
    <category>Sports</category>
    <category>Tennis de Table</category>
    <category>France</category>
    <image>
      <url>https://tournois-tt.fr/thumbnail.png</url>
      <title>Tournois FFTT</title>
      <link>https://tournois-tt.fr</link>
      <width>1200</width>
      <height>630</height>
    </image>
${items}
  </channel>
</rss>`;
}

// Fonction principale
function generateRSSFeed() {
  console.log('📡 Génération du flux RSS...');

  try {
    // Vérifier que le fichier de données existe
    if (!fs.existsSync(API_DATA_PATH)) {
      throw new Error(`Fichier de données non trouvé: ${API_DATA_PATH}`);
    }

    // Lire les données des tournois
    const tournamentsData = JSON.parse(fs.readFileSync(API_DATA_PATH, 'utf8'));
    console.log(`📊 ${tournamentsData.length} tournois trouvés pour le RSS`);

    // Générer le flux RSS
    const rssXML = generateRSS(tournamentsData);
    
    // Écrire le RSS dans le répertoire public (pour le développement)
    fs.writeFileSync(RSS_PATH, rssXML);
    console.log(`✅ RSS généré: ${RSS_PATH}`);

    // Écrire aussi dans le répertoire build (pour la production)
    if (fs.existsSync(path.dirname(BUILD_RSS_PATH))) {
      fs.writeFileSync(BUILD_RSS_PATH, rssXML);
      console.log(`✅ RSS généré: ${BUILD_RSS_PATH}`);
    }

    // Statistiques
    const recentCount = tournamentsData.length;
    console.log(`📈 ${recentCount} tournois récents dans le RSS`);

    // Vérifier la taille du fichier
    const stats = fs.statSync(RSS_PATH);
    const fileSizeKB = (stats.size / 1024).toFixed(2);
    console.log(`💾 Taille du RSS: ${fileSizeKB} KB`);

    console.log('🎉 Flux RSS généré avec succès!');

  } catch (error) {
    console.error('❌ Erreur lors de la génération du RSS:', error.message);
    process.exit(1);
  }
}

// Exécuter le script si appelé directement
if (require.main === module) {
  generateRSSFeed();
}

module.exports = { generateRSSFeed };
