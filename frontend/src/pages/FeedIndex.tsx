import React from 'react';
import { Link } from 'react-router-dom';
import { Tournament } from '../lib/api/types';
import { mapTournamentType } from '../lib/utils/tournament';
import { formatDateDDMMYYYY } from '../lib/utils/date';
import Meta from '../components/Meta';
import { useTournaments } from '../lib/hooks/useTournaments';

const FeedIndex: React.FC = () => {
  const { tournaments, loading, error } = useTournaments();

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
  return (
    <>
      <Meta isIndex={true} totalTournaments={tournaments.length} />
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div style={{ marginTop: '10px' }}>
            <Link to="/" className="text-blue-600 hover:text-blue-800 transition-colors duration-200">
              Retour à la carte
            </Link>
          </div>
          <header className="mb-8">
            <h1 className="text-4xl font-bold text-gray-900 mb-4">
              Tournois FFTT homologués
            </h1>
          </header>

          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {tournaments.map((tournament) => (
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
                        Dotation totale : {(tournament.endowment / 100).toLocaleString('fr-FR')}€
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
                    <a 
                      href={`${tournament.rules.url}${tournament.rules.url.includes('?') ? '&' : '?'}utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=fftt_rules&utm_content=feed_list`} 
                      target="_blank" 
                      rel="noopener noreferrer" 
                      className="inline-flex items-center text-blue-600 hover:text-blue-800 font-medium text-sm transition-colors duration-200"
                      onClick={() => {
                        if (typeof window !== 'undefined' && window.gtag) {
                          window.gtag('event', 'click', {
                            event_category: 'external_link',
                            event_label: 'FFTT Rules - Feed List',
                            value: tournament.rules?.url,
                            link_url: tournament.rules?.url,
                            link_text: 'Voir le règlement',
                            tournament_name: tournament.name
                          });
                        }
                      }}
                    >
                      Voir le règlement
                    </a>
                  </div>
                )}

                {new Date(tournament.startDate) >= new Date() && (
                  <div className="mt-4 pt-4 border-t border-gray-200">
                    {tournament.page ? (
                      <a 
                        href={`${tournament.page}${tournament.page.includes('?') ? '&' : '?'}utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=page_signup&utm_content=feed_list`} 
                        target="_blank" 
                        rel="noopener noreferrer" 
                        className="inline-flex items-center text-blue-600 hover:text-blue-800 font-medium text-sm transition-colors duration-200"
                        onClick={() => {
                          if (typeof window !== 'undefined' && window.gtag) {
                            window.gtag('event', 'click', {
                              event_category: 'external_link',
                              event_label: 'Page Signup - Feed List',
                              value: tournament.page,
                              link_url: tournament.page,
                              link_text: 'Inscription',
                              tournament_name: tournament.name
                            });
                          }
                        }}
                      >
                        Inscription
                      </a>
                    ) : (
                      <span className="text-gray-500 text-sm">Pas encore de lien d'inscription</span>
                    )}
                  </div>
                )}

              </article>
            ))}
          </div>

          {tournaments.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 text-lg">Aucun tournoi disponible pour le moment.</p>
            </div>
          )}
        </div>
      </div >
    </>
  );
};

export default FeedIndex;
