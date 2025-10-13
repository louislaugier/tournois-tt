#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

/**
 * Script pour générer un sitemap XML complet incluant toutes les pages
 * - Pages principales du site
 * - Toutes les pages de feed des tournois
 */

const API_DATA_PATH = process.env.OUTPUT_DIR === '/usr/share/nginx/html' 
  ? '/app/api/cache/data.json'  // Production Docker
  : path.join(__dirname, '../../api/cache/data.json'); // Local development
const OUTPUT_DIR = process.env.OUTPUT_DIR || path.join(__dirname, '../public');
const SITEMAP_PATH = path.join(OUTPUT_DIR, 'sitemap.xml');
const BUILD_SITEMAP_PATH = path.join(__dirname, '../build/sitemap.xml');

// Ensure output directory exists
if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

// Pages principales du site
const MAIN_PAGES = [
  {
    url: 'https://tournois-tt.fr/',
    changefreq: 'always',
    priority: '1.0',
    lastmod: new Date().toISOString()
  },
  {
    url: 'https://tournois-tt.fr/a-propos',
    changefreq: 'monthly',
    priority: '0.3',
    lastmod: new Date().toISOString()
  },
  {
    url: 'https://tournois-tt.fr/feed/',
    changefreq: 'daily',
    priority: '0.8',
    lastmod: new Date().toISOString()
  },
  {
    url: 'https://tournois-tt.fr/rss.xml',
    changefreq: 'hourly',
    priority: '0.7',
    lastmod: new Date().toISOString()
  }
];

// Fonction pour générer le sitemap XML
function generateSitemap(tournaments = []) {
  const now = new Date().toISOString();
  
  let sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`;

  // Ajouter les pages principales
  MAIN_PAGES.forEach(page => {
    sitemap += `  <url>
    <loc>${page.url}</loc>
    <lastmod>${page.lastmod}</lastmod>
    <changefreq>${page.changefreq}</changefreq>
    <priority>${page.priority}</priority>
  </url>
`;
  });

  // Ajouter toutes les pages de tournois
  tournaments.forEach(tournament => {
    sitemap += `  <url>
    <loc>https://tournois-tt.fr/feed/${tournament.id}</loc>
    <lastmod>${now}</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.6</priority>
  </url>
`;
  });

  sitemap += `</urlset>`;

  return sitemap;
}

// Fonction principale
function generateCompleteSitemap() {
  console.log('🗺️  Génération du sitemap XML complet...');

  try {
    // Vérifier que le fichier de données existe
    if (!fs.existsSync(API_DATA_PATH)) {
      throw new Error(`Fichier de données non trouvé: ${API_DATA_PATH}`);
    }

    // Lire les données des tournois
    const tournamentsData = JSON.parse(fs.readFileSync(API_DATA_PATH, 'utf8'));
    console.log(`📊 ${tournamentsData.length} tournois trouvés pour le sitemap`);

    // Générer le sitemap
    const sitemapXML = generateSitemap(tournamentsData);
    
    // Écrire le sitemap dans le répertoire public (pour le développement)
    fs.writeFileSync(SITEMAP_PATH, sitemapXML);
    console.log(`✅ Sitemap généré: ${SITEMAP_PATH}`);

    // Écrire aussi dans le répertoire build (pour la production)
    if (fs.existsSync(path.dirname(BUILD_SITEMAP_PATH))) {
      fs.writeFileSync(BUILD_SITEMAP_PATH, sitemapXML);
      console.log(`✅ Sitemap généré: ${BUILD_SITEMAP_PATH}`);
    }

    // Statistiques
    const totalUrls = MAIN_PAGES.length + tournamentsData.length;
    console.log(`📈 Total URLs dans le sitemap: ${totalUrls}`);
    console.log(`   - Pages principales: ${MAIN_PAGES.length}`);
    console.log(`   - Pages de tournois: ${tournamentsData.length}`);

    // Vérifier la taille du fichier
    const stats = fs.statSync(SITEMAP_PATH);
    const fileSizeKB = (stats.size / 1024).toFixed(2);
    console.log(`💾 Taille du sitemap: ${fileSizeKB} KB`);

    console.log('🎉 Sitemap généré avec succès!');

  } catch (error) {
    console.error('❌ Erreur lors de la génération du sitemap:', error.message);
    process.exit(1);
  }
}

// Exécuter le script si appelé directement
if (require.main === module) {
  generateCompleteSitemap();
}

module.exports = { generateCompleteSitemap };
