import { Helmet } from "react-helmet";

export default () => <Helmet>
    <title>Carte des Tournois de Tennis de Table homologués FFTT</title>
    <meta name="description" content="Découvrez tous les tournois de tennis de table en France sur une carte interactive. Filtrez par date, région, et catégorie. Informations détaillées sur les règlements, dates et lieux des compétitions FFTT." />
    <meta name="keywords" content="tennis de table, tournois, FFTT, ping pong, France, carte interactive, compétitions" />
    <meta property="og:title" content="Carte des Tournois de Tennis de Table en France | FFTT" />
    <meta property="og:description" content="Découvrez tous les tournois de tennis de table en France sur une carte interactive. Filtrez par date, région, et catégorie." />
    <meta property="og:type" content="website" />
    <meta property="og:url" content="https://tournois-tt.fr" />
    <meta property="og:image" content="https://cdn-icons-png.flaticon.com/512/9978/9978844.png" />
    <meta property="og:site_name" content="Carte des Tournois FFTT" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="Carte des Tournois de Tennis de Table en France" />
    <meta name="twitter:description" content="Découvrez tous les tournois de tennis de table en France sur une carte interactive. Filtrez par date, région, et catégorie." />
    <meta name="twitter:image" content="https://cdn-icons-png.flaticon.com/512/9978/9978844.png" />
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
                "description": "Carte interactive des tournois de tennis de table en France avec filtres par date, région et catégorie",
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