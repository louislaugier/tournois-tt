#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

/**
 * Script pour générer les pages statiques du feed HTML
 * Ce script peut être exécuté via cron pour maintenir les pages à jour
 */

const API_DATA_PATH = path.join(__dirname, '../../api/cache/data.json');
const BUILD_DIR = path.join(__dirname, '../build');
const FEED_DIR = path.join(BUILD_DIR, 'feed');

// Fonction pour formater les dates
function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
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

// Fonction pour générer le HTML d'une page de tournoi
function generateTournamentHTML(tournament) {
  const fullAddress = [
    tournament.address.streetAddress,
    tournament.address.postalCode,
    tournament.address.addressLocality
  ].filter(Boolean).join(', ');

  const title = `${tournament.name} - Tournoi ${mapTournamentType(tournament.type)} | FFTT`;
  const description = `Tournoi de tennis de table ${mapTournamentType(tournament.type)} organisé par ${tournament.club.name} le ${formatDate(tournament.startDate)}${tournament.startDate !== tournament.endDate ? ` au ${formatDate(tournament.endDate)}` : ''} à ${tournament.address.addressLocality}. ${tournament.endowment > 0 ? `Dotation: ${(tournament.endowment / 100).toLocaleString('fr-FR')}€.` : ''} Informations pratiques, règlement et inscription.`;

  return `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${title}</title>
    <meta name="description" content="${description}">
    <meta name="keywords" content="tennis de table, tournoi, ${mapTournamentType(tournament.type)}, ${tournament.club.name}, ${tournament.address.addressLocality}, FFTT, ping pong, compétition">
    <meta property="og:title" content="${title}">
    <meta property="og:description" content="${description}">
    <meta property="og:type" content="sports_event">
    <meta property="og:url" content="https://tournois-tt.fr/feed/${tournament.id}">
    <meta property="og:image" content="https://tournois-tt.fr/thumbnail.png">
    <meta property="og:site_name" content="Carte des Tournois FFTT">
    <meta property="og:event:start_time" content="${tournament.startDate}">
    <meta property="og:event:end_time" content="${tournament.endDate}">
    <meta property="og:event:location" content="${fullAddress}">
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:site" content="@tournoistt">
    <meta name="twitter:title" content="${title}">
    <meta name="twitter:description" content="${description}">
    <meta name="twitter:image" content="https://tournois-tt.fr/thumbnail.png">
    <meta name="robots" content="index, follow">
    <link rel="canonical" href="https://tournois-tt.fr/feed/${tournament.id}">
    <script src="https://cdn.tailwindcss.com"></script>
    <script type="application/ld+json">
    {
        "@context": "https://schema.org",
        "@type": "SportsEvent",
        "name": "${tournament.name}",
        "description": "${description}",
        "startDate": "${tournament.startDate}",
        "endDate": "${tournament.endDate}",
        "location": {
            "@type": "Place",
            "name": "${tournament.address.disambiguatingDescription || tournament.address.addressLocality}",
            "address": {
                "@type": "PostalAddress",
                "streetAddress": "${tournament.address.streetAddress}",
                "addressLocality": "${tournament.address.addressLocality}",
                "postalCode": "${tournament.address.postalCode}",
                "addressCountry": "FR"
            }${tournament.address.latitude && tournament.address.longitude ? `,
            "geo": {
                "@type": "GeoCoordinates",
                "latitude": ${tournament.address.latitude},
                "longitude": ${tournament.address.longitude}
            }` : ''}
        },
        "organizer": {
            "@type": "Organization",
            "name": "Fédération Française de Tennis de Table",
            "alternateName": "FFTT",
            "url": "https://www.fftt.com"
        },
        "sport": "Tennis de Table",
        "url": "https://tournois-tt.fr/feed/${tournament.id}",
        "eventStatus": "EventScheduled",
        "eventAttendanceMode": "OfflineEventAttendanceMode",
        "performer": {
            "@type": "SportsTeam",
            "name": "Joueurs de Tennis de Table"
        },
        "image": "https://www.fftt.com/site/medias/header2017/logo.png"${tournament.endowment > 0 ? `,
        "offers": {
            "@type": "Offer",
            "url": "https://tournois-tt.fr/feed/${tournament.id}",
            "price": "${tournament.endowment / 100}",
            "priceCurrency": "EUR",
            "availability": "https://schema.org/InStock"
        }` : ''}
    }
    </script>
</head>
<body class="min-h-screen bg-gray-50">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <!-- Breadcrumb -->
        <nav class="mb-8">
            <ol class="flex items-center space-x-2 text-sm text-gray-500">
                <li><a href="/feed" class="hover:text-gray-700 transition-colors duration-200">Tournois</a></li>
                <li><svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd"></path></svg></li>
                <li class="text-gray-900 font-medium truncate">${tournament.name}</li>
            </ol>
        </nav>

        <!-- Main content -->
        <article class="bg-white rounded-lg shadow-lg overflow-hidden">
            <!-- Header -->
            <div class="bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-8 text-white">
                <div class="flex items-center gap-3 mb-4">
                    <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-white bg-opacity-20">
                        ${mapTournamentType(tournament.type)}
                    </span>
                    ${tournament.endowment > 0 ? `<span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-500 bg-opacity-20">
                        Dotation: ${(tournament.endowment / 100).toLocaleString('fr-FR')}€
                    </span>` : ''}
                </div>
                <h1 class="text-3xl font-bold mb-2">${tournament.name}</h1>
                <p class="text-blue-100 text-lg">Organisé par ${tournament.club.name}</p>
            </div>

            <!-- Content -->
            <div class="p-6">
                <div class="grid gap-6 md:grid-cols-2">
                    <!-- Informations principales -->
                    <div class="space-y-6">
                        <section>
                            <h2 class="text-xl font-semibold text-gray-900 mb-4 flex items-center">
                                <svg style="width: 12px; height: 12px; margin-right: 8px; color: #2563eb; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                </svg>
                                Informations du tournoi
                            </h2>
                            <dl class="space-y-3">
                                <div>
                                    <dt class="text-sm font-medium text-gray-500">Type de tournoi</dt>
                                    <dd class="text-sm text-gray-900">${mapTournamentType(tournament.type)}</dd>
                                </div>
                                <div>
                                    <dt class="text-sm font-medium text-gray-500">Club organisateur</dt>
                                    <dd class="text-sm text-gray-900">${tournament.club.name}</dd>
                                </div>
                                <div>
                                    <dt class="text-sm font-medium text-gray-500">Dates</dt>
                                    <dd class="text-sm text-gray-900">
                                        ${formatDate(tournament.startDate)}${tournament.startDate !== tournament.endDate ? ` - ${formatDate(tournament.endDate)}` : ''}
                                    </dd>
                                </div>
                                ${tournament.endowment > 0 ? `<div>
                                    <dt class="text-sm font-medium text-gray-500">Dotation</dt>
                                    <dd class="text-sm text-gray-900 font-semibold">
                                        ${(tournament.endowment / 100).toLocaleString('fr-FR')}€
                                    </dd>
                                </div>` : ''}
                            </dl>
                        </section>

                        <!-- Adresse -->
                        <section>
                            <h2 class="text-xl font-semibold text-gray-900 mb-4 flex items-center">
                                <svg style="width: 12px; height: 12px; margin-right: 8px; color: #2563eb; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
                                </svg>
                                Lieu du tournoi
                            </h2>
                            <div class="bg-gray-50 rounded-lg p-4">
                                <p class="text-sm text-gray-900">${fullAddress}</p>
                                ${tournament.address.disambiguatingDescription ? `<p class="text-sm text-gray-600 mt-1">
                                    ${tournament.address.disambiguatingDescription}
                                </p>` : ''}
                            </div>
                        </section>
                    </div>

                    <!-- Actions et liens -->
                    <div class="space-y-6">
                        ${tournament.rules?.url ? `<!-- Règlement -->
                        <section>
                            <h2 class="text-xl font-semibold text-gray-900 mb-4 flex items-center">
                                <svg style="width: 12px; height: 12px; margin-right: 8px; color: #2563eb; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                                </svg>
                                Règlement
                            </h2>
                            <a href="${tournament.rules.url}" target="_blank" rel="noopener noreferrer" class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors duration-200">
                                <svg style="width: 12px; height: 12px; margin-right: 8px; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                                </svg>
                                Télécharger le règlement (PDF)
                            </a>
                        </section>` : ''}

                        ${(tournament.page && new Date(tournament.startDate) >= new Date()) ? `<!-- Inscription -->
                        <section>
                            <h2 class="text-xl font-semibold text-gray-900 mb-4 flex items-center">
                                <svg style="width: 12px; height: 12px; margin-right: 8px; color: #2563eb; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                                </svg>
                                Inscription
                            </h2>
                            <a href="${tournament.page}" target="_blank" rel="noopener noreferrer" class="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors duration-200">
                                <svg style="width: 12px; height: 12px; margin-right: 8px; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                                </svg>
                                S'inscrire au tournoi
                            </a>
                        </section>` : (!tournament.page && new Date(tournament.startDate) >= new Date()) ? `<!-- Inscription -->
                        <section>
                            <h2 class="text-xl font-semibold text-gray-900 mb-4 flex items-center">
                                <svg style="width: 12px; height: 12px; margin-right: 8px; color: #2563eb; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                                </svg>
                                Inscription
                            </h2>
                            <p class="text-sm text-gray-500">Pas encore de lien d'inscription</p>
                        </section>` : ''}

                        <!-- Navigation -->
                        <section>
                            <h2 class="text-xl font-semibold text-gray-900 mb-4">Navigation</h2>
                            <div class="space-y-2">
                                <a href="/feed" class="block w-full text-center px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition-colors duration-200">
                                    ← Retour à la liste des tournois
                                </a>
                                <a href="/" class="block w-full text-center px-4 py-2 bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors duration-200">
                                    Voir sur la carte
                                </a>
                            </div>
                        </section>
                    </div>
                </div>
            </div>
        </article>
    </div>
</body>
</html>`;
}

// Fonction pour générer l'index HTML
function generateIndexHTML(tournaments) {
  const title = `Liste des Tournois de Tennis de Table`;
  const description = `Découvrez ${tournaments.length} tournoi${tournaments.length > 1 ? 's' : ''} de tennis de table en France. Informations détaillées sur les dates, lieux, dotations et règlements des compétitions FFTT.`;

  // Sort tournaments by start date: furthest in the future first, then furthest in the past
  const sortedTournaments = [...tournaments].sort((a, b) => {
    return new Date(b.startDate) - new Date(a.startDate);
  });

  const tournamentCards = sortedTournaments.map(tournament => `
    <article class="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-200 p-6">
        <div class="mb-4">
            <h2 class="text-xl font-semibold text-gray-900 mb-2">
                <a href="/feed/${tournament.id}" class="hover:text-blue-600 transition-colors duration-200">
                    ${tournament.name}
                </a>
            </h2>
            <div class="flex items-center gap-2 mb-2">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    ${mapTournamentType(tournament.type)}
                </span>
                ${tournament.endowment > 0 ? `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    ${(tournament.endowment / 100).toLocaleString('fr-FR')}€
                </span>` : ''}
            </div>
        </div>

        <div class="space-y-2 text-sm text-gray-600">
            <div class="flex items-center">
                <svg style="width: 10px; height: 10px; margin-right: 5px; color: #666; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
                </svg>
                <span class="font-medium">${tournament.club.name}</span>
            </div>
            
            <div class="flex items-center">
                <svg style="width: 10px; height: 10px; margin-right: 5px; color: #666; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
                </svg>
                <span>
                    ${formatDate(tournament.startDate)}${tournament.startDate !== tournament.endDate ? ` - ${formatDate(tournament.endDate)}` : ''}
                </span>
            </div>
            
            <div class="flex items-center">
                <svg style="width: 10px; height: 10px; margin-right: 5px; color: #666; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
                </svg>
                <span>${tournament.address.postalCode} ${tournament.address.addressLocality}</span>
            </div>
        </div>

        <div class="mt-4 pt-4 border-t border-gray-200">
            <a href="/feed/${tournament.id}" class="inline-flex items-center text-blue-600 hover:text-blue-800 font-medium text-sm transition-colors duration-200">
                Voir les détails
                <svg style="width: 10px; height: 10px; margin-left: 5px; display: inline-block; vertical-align: middle;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                </svg>
            </a>
        </div>
    </article>
  `).join('');

  return `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${title}</title>
    <meta name="description" content="${description}">
    <meta name="keywords" content="tennis de table, tournois, FFTT, ping pong, France, liste, compétitions, dates, lieux">
    <meta property="og:title" content="Liste des Tournois de Tennis de Table en France">
    <meta property="og:description" content="${tournaments.length} tournoi${tournaments.length > 1 ? 's' : ''} de tennis de table disponibles en France avec toutes les informations pratiques.">
    <meta property="og:type" content="website">
    <meta property="og:url" content="https://tournois-tt.fr/feed">
    <meta property="og:image" content="https://tournois-tt.fr/thumbnail.png">
    <meta property="og:site_name" content="Carte des Tournois FFTT">
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:site" content="@tournoistt">
    <meta name="twitter:title" content="Liste des Tournois de Tennis de Table en France">
    <meta name="twitter:description" content="${tournaments.length} tournoi${tournaments.length > 1 ? 's' : ''} de tennis de table disponibles en France.">
    <meta name="twitter:image" content="https://tournois-tt.fr/thumbnail.png">
    <meta name="robots" content="index, follow">
    <link rel="canonical" href="https://tournois-tt.fr/feed">
    <script src="https://cdn.tailwindcss.com"></script>
    <script type="application/ld+json">
    {
        "@context": "https://schema.org",
        "@type": "ItemList",
        "name": "Tournois de Tennis de Table en France",
        "description": "Liste de ${tournaments.length} tournoi${tournaments.length > 1 ? 's' : ''} de tennis de table en France",
        "numberOfItems": ${tournaments.length},
        "itemListElement": []
    }
    </script>
</head>
<body class="min-h-screen bg-gray-50">
    <div style="max-width: 1280px; margin: 0 auto; padding: 32px 16px;">
        <header style="margin-bottom: 32px;">
            <h1 style="font-size: 36px; font-weight: bold; color: #111; margin-bottom: 16px;">
                Tournois FFTT homologués
            </h1>
        </header>

        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 24px;">
            ${tournamentCards}
        </div>

        ${tournaments.length === 0 ? `<div class="text-center py-12">
            <p class="text-gray-500 text-lg">Aucun tournoi disponible pour le moment.</p>
        </div>` : ''}
    </div>
</body>
</html>`;
}

// Fonction principale
function generateStaticFeed() {
  console.log('🚀 Génération du feed HTML statique...');

  try {
    // Vérifier que le fichier de données existe
    if (!fs.existsSync(API_DATA_PATH)) {
      throw new Error(`Fichier de données non trouvé: ${API_DATA_PATH}`);
    }

    // Lire les données des tournois
    const tournamentsData = JSON.parse(fs.readFileSync(API_DATA_PATH, 'utf8'));
    console.log(`📊 ${tournamentsData.length} tournois trouvés`);

    // Créer le répertoire de sortie
    if (!fs.existsSync(FEED_DIR)) {
      fs.mkdirSync(FEED_DIR, { recursive: true });
    }

    // Générer l'index
    const indexHTML = generateIndexHTML(tournamentsData);
    fs.writeFileSync(path.join(FEED_DIR, 'index.html'), indexHTML);
    console.log('✅ Index généré: /feed/index.html');

    // Générer les pages individuelles
    let generatedCount = 0;
    for (const tournament of tournamentsData) {
      const tournamentHTML = generateTournamentHTML(tournament);
      const filename = `${tournament.id}.html`;
      fs.writeFileSync(path.join(FEED_DIR, filename), tournamentHTML);
      generatedCount++;
    }
    console.log(`✅ ${generatedCount} pages de tournois générées`);

    // Générer un sitemap pour le feed uniquement
    const feedSitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>https://tournois-tt.fr/feed/</loc>
        <lastmod>${new Date().toISOString()}</lastmod>
        <changefreq>daily</changefreq>
        <priority>0.8</priority>
    </url>
    ${tournamentsData.map(tournament => `    <url>
        <loc>https://tournois-tt.fr/feed/${tournament.id}</loc>
        <lastmod>${new Date().toISOString()}</lastmod>
        <changefreq>weekly</changefreq>
        <priority>0.6</priority>
    </url>`).join('\n')}
</urlset>`;

    fs.writeFileSync(path.join(FEED_DIR, 'sitemap.xml'), feedSitemap);
    console.log('✅ Sitemap du feed généré: /feed/sitemap.xml');

    console.log('🎉 Génération terminée avec succès!');
    console.log(`📁 Fichiers générés dans: ${FEED_DIR}`);

  } catch (error) {
    console.error('❌ Erreur lors de la génération:', error.message);
    process.exit(1);
  }
}

// Exécuter le script si appelé directement
if (require.main === module) {
  generateStaticFeed();
}

module.exports = { generateStaticFeed };
