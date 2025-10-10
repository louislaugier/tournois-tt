import { Helmet } from "react-helmet";
import { Tournament } from "../lib/api/types";
import { mapTournamentType } from "../lib/utils/tournament";
import { formatDateDDMMYYYY } from "../lib/utils/date";

interface FeedMetaProps {
  tournament?: Tournament;
  isIndex?: boolean;
  totalTournaments?: number;
}

const FeedMeta: React.FC<FeedMetaProps> = ({ tournament, isIndex = false, totalTournaments = 0 }) => {
  if (isIndex) {
    return (
      <Helmet>
        <title>Liste des Tournois de Tennis de Table</title>
        <meta 
          name="description" 
          content={`Découvrez ${totalTournaments} tournoi${totalTournaments > 1 ? 's' : ''} de tennis de table en France. Informations détaillées sur les dates, lieux, dotations et règlements des compétitions FFTT.`} 
        />
        <meta name="keywords" content="tennis de table, tournois, FFTT, ping pong, France, liste, compétitions, dates, lieux" />
        <meta property="og:title" content="Liste des Tournois de Tennis de Table en France" />
        <meta 
          property="og:description" 
          content={`${totalTournaments} tournoi${totalTournaments > 1 ? 's' : ''} de tennis de table disponibles en France avec toutes les informations pratiques.`} 
        />
        <meta property="og:type" content="website" />
        <meta property="og:url" content="https://tournois-tt.fr/feed" />
        <meta property="og:image" content="https://tournois-tt.fr/thumbnail.png" />
        <meta property="og:site_name" content="Carte des Tournois FFTT" />
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:title" content="Liste des Tournois de Tennis de Table en France" />
        <meta 
          name="twitter:description" 
          content={`${totalTournaments} tournoi${totalTournaments > 1 ? 's' : ''} de tennis de table disponibles en France.`} 
        />
        <meta name="twitter:image" content="https://tournois-tt.fr/thumbnail.png" />
        <meta name="robots" content="index, follow" />
        <link rel="canonical" href="https://tournois-tt.fr/feed" />
        <script type="application/ld+json">
          {JSON.stringify({
            "@context": "https://schema.org",
            "@type": "ItemList",
            "name": "Tournois de Tennis de Table en France",
            "description": `Liste de ${totalTournaments} tournoi${totalTournaments > 1 ? 's' : ''} de tennis de table en France`,
            "numberOfItems": totalTournaments,
            "itemListElement": []
          })}
        </script>
      </Helmet>
    );
  }

  if (!tournament) return null;

  const fullAddress = [
    tournament.address.streetAddress,
    tournament.address.postalCode,
    tournament.address.addressLocality
  ].filter(Boolean).join(', ');

  const title = `${tournament.name} - Tournoi ${mapTournamentType(tournament.type)} | FFTT`;
  const description = `Tournoi de tennis de table ${mapTournamentType(tournament.type)} organisé par ${tournament.club.name} le ${formatDateDDMMYYYY(tournament.startDate)}${tournament.startDate !== tournament.endDate ? ` au ${formatDateDDMMYYYY(tournament.endDate)}` : ''} à ${tournament.address.addressLocality}. ${tournament.endowment > 0 ? `Dotation: ${tournament.endowment.toLocaleString('fr-FR')}€.` : ''} Informations pratiques, règlement et inscription.`;

  return (
    <Helmet>
      <title>{title}</title>
      <meta name="description" content={description} />
      <meta name="keywords" content={`tennis de table, tournoi, ${mapTournamentType(tournament.type)}, ${tournament.club.name}, ${tournament.address.addressLocality}, FFTT, ping pong, compétition`} />
      <meta property="og:title" content={title} />
      <meta property="og:description" content={description} />
      <meta property="og:type" content="sports_event" />
      <meta property="og:url" content={`https://tournois-tt.fr/feed/${tournament.id}`} />
        <meta property="og:image" content="https://tournois-tt.fr/thumbnail.png" />
        <meta property="og:site_name" content="Carte des Tournois FFTT" />
        <meta property="og:event:start_time" content={tournament.startDate} />
        <meta property="og:event:end_time" content={tournament.endDate} />
        <meta property="og:event:location" content={fullAddress} />
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:title" content={title} />
        <meta name="twitter:description" content={description} />
        <meta name="twitter:image" content="https://tournois-tt.fr/thumbnail.png" />
      <meta name="robots" content="index, follow" />
      <link rel="canonical" href={`https://tournois-tt.fr/feed/${tournament.id}`} />
      <script type="application/ld+json">
        {JSON.stringify({
          "@context": "https://schema.org",
          "@type": "SportsEvent",
          "name": tournament.name,
          "description": description,
          "startDate": tournament.startDate,
          "endDate": tournament.endDate,
          "location": {
            "@type": "Place",
            "name": tournament.address.disambiguatingDescription || tournament.address.addressLocality,
            "address": {
              "@type": "PostalAddress",
              "streetAddress": tournament.address.streetAddress,
              "addressLocality": tournament.address.addressLocality,
              "postalCode": tournament.address.postalCode,
              "addressCountry": "FR"
            },
            "geo": tournament.address.latitude && tournament.address.longitude ? {
              "@type": "GeoCoordinates",
              "latitude": tournament.address.latitude,
              "longitude": tournament.address.longitude
            } : undefined
          },
          "organizer": {
            "@type": "Organization",
            "name": tournament.club.name
          },
          "sport": "Tennis de Table",
          "url": `https://tournois-tt.fr/feed/${tournament.id}`,
          "eventStatus": "EventScheduled",
          "eventAttendanceMode": "OfflineEventAttendanceMode",
          "offers": tournament.endowment > 0 ? {
            "@type": "Offer",
            "price": tournament.endowment.toString(),
            "priceCurrency": "EUR",
            "availability": "https://schema.org/InStock"
          } : undefined
        })}
      </script>
    </Helmet>
  );
};

export default FeedMeta;
