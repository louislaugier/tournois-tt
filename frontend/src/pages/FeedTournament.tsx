import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { Tournament } from '../lib/api/types';
import { mapTournamentType } from '../lib/utils/tournament';
import { formatDateDDMMYYYY } from '../lib/utils/date';
import Meta from '../components/Meta';
import { useTournaments } from '../lib/hooks/useTournaments';

const FeedTournament: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { tournaments, loading, error } = useTournaments();
  const tournament = tournaments.find(t => t.id.toString() === id);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Chargement...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Erreur: {error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Réessayer
          </button>
        </div>
      </div>
    );
  }

  if (!tournament) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Tournoi non trouvé</h1>
          <p className="text-gray-600 mb-6">Le tournoi demandé n'existe pas ou n'est plus disponible.</p>
          <Link
            to="/feed"
            className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors duration-200"
          >
            Retour à la liste des tournois
          </Link>
        </div>
      </div>
    );
  }

  const fullAddress = [
    tournament.address.streetAddress,
    tournament.address.postalCode,
    tournament.address.addressLocality
  ].filter(Boolean).join(', ');

  return (
    <>
      <Meta tournament={tournament} />
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Breadcrumb */}
          <nav className="mb-8">
            <Link to="/feed" className="hover:text-gray-700 transition-colors duration-200">
              Retour aux tournois
            </Link>
          </nav>

          {/* Main content */}
          <article
            key={tournament.id}
            className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-200 p-6"
          >
            <div className="mb-4">
              <h2 className="text-xl font-semibold text-gray-900" style={{ marginBottom: 0 }}>
                <Link
                  to={`/feed/${tournament.id}`}
                  className="hover:text-blue-600 transition-colors duration-200"
                >
                  {tournament.name}
                </Link>
              </h2>
              <div className="flex items-center gap-2 mb-2">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  Type de tournoi : {mapTournamentType(tournament.type)}
                </span>
              </div>
              <div className="flex items-center gap-2 mb-2">
                {tournament.endowment > 0 && (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    Dotation totale : {tournament.endowment.toLocaleString('fr-FR')}€
                  </span>
                )}
              </div>
            </div>

            <div className="space-y-2 text-sm text-gray-600">
              <div className="flex items-center">

                <span className="font-medium">{tournament.club.name}</span>
              </div>

              <div className="flex items-center">

                <span>
                  {formatDateDDMMYYYY(tournament.startDate)}
                  {tournament.startDate !== tournament.endDate && (
                    <span> - {formatDateDDMMYYYY(tournament.endDate)}</span>
                  )}
                </span>
              </div>

              <div className="flex items-center">
                <span>
                  {tournament.address.streetAddress}, {tournament.address.postalCode} {tournament.address.addressLocality}
                </span>
              </div>
            </div>

            {tournament.rules?.url && (
              <div className="mt-4 pt-4 border-t border-gray-200">
                <a href={tournament.rules.url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center text-blue-600 hover:text-blue-800 font-medium text-sm transition-colors duration-200">
                  Voir le règlement
                </a>
              </div>
            )}

          </article>

          {/* Structured Data JSON-LD */}
          <script
            type="application/ld+json"
            dangerouslySetInnerHTML={{
              __html: JSON.stringify({
                "@context": "https://schema.org",
                "@type": "SportsEvent",
                "name": tournament.name,
                "description": `Tournoi de tennis de table ${mapTournamentType(tournament.type)} organisé par ${tournament.club.name}`,
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
                  }
                },
                "organizer": {
                  "@type": "Organization",
                  "name": tournament.club.name
                },
                "sport": "Tennis de Table",
                "url": `${window.location.origin}/feed/${tournament.id}`
              })
            }}
          />
        </div>
      </div>
    </>
  );
};

export default FeedTournament;
