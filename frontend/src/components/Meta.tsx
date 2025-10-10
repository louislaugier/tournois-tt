import { Helmet } from "react-helmet";

export default () => <Helmet>
    <title>Tournois FFTT</title>
    <meta name="description" content="Carte interactive des tournois de tennis de table en France. Trouvez les prochains tournois FFTT près de chez vous. Tri & recherche par type de tournoi, club organisateur, date, région, code postal, ville, dotation." />
    <meta name="keywords" content="tennis de table, ping pong, tournois, FFTT, France, carte, compétition" />
    <meta property="og:title" content="Tournois FFTT" />
    <meta property="og:description" content="Carte interactive des tournois de tennis de table en France. Trouvez les prochains tournois FFTT près de chez vous. Tri & recherche par type de tournoi, club organisateur, date, région, code postal, ville, dotation." />
    <meta property="og:type" content="website" />
    <meta property="og:url" content="https://tournois-tt.fr" />
    <meta property="og:image" content="https://tournois-tt.fr/thumbnail.png" />
    <meta property="og:site_name" content="Tournois FFTT" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="Tournois FFTT" />
    <meta name="twitter:description" content="Carte interactive des tournois de tennis de table en France. Trouvez les prochains tournois FFTT près de chez vous. Tri & recherche par type de tournoi, club organisateur, date, région, code postal, ville, dotation." />
    <meta name="twitter:image" content="https://tournois-tt.fr/thumbnail.png" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="theme-color" content="#242730" />
    <meta name="robots" content="index, follow" />
    <link rel="canonical" href="https://tournois-tt.fr" />
    <script type="application/ld+json">
        {`
            {
                "@context": "https://schema.org",
                "@type": "SportsEvent",
                "name": "Tournois de Tennis de Table en France",
                "description": "Carte interactive des tournois FFTT homologués. Tri & recherche par type de tournoi, club organisateur, date, région, code postal, ville, dotation.",
                "sport": "Tennis de Table",
                "location": {
                    "@type": "Country",
                    "name": "France"
                },
                "organizer": {
                    "@type": "Organization",
                    "name": "Fédération Française de Tennis de Table",
                    "alternateName": "FFTT"
                },
                "eventStatus": "EventScheduled",
                "eventAttendanceMode": "OfflineEventAttendanceMode",
                "offers": {
                    "@type": "Offer",
                    "availability": "https://schema.org/InStock",
                    "price": "0",
                    "priceCurrency": "EUR"
                }
            }
        `}
    </script>
</Helmet>