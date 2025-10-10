import { useState, useEffect } from 'react';
import { Tournament } from '../api/types';
import { fetchAllTournaments } from '../api/tournaments';

export const useTournaments = () => {
  const [tournaments, setTournaments] = useState<Tournament[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTournaments = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await fetchAllTournaments();
        setTournaments(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erreur lors du chargement des tournois');
      } finally {
        setLoading(false);
      }
    };

    fetchTournaments();
  }, []);

  return { tournaments, loading, error };
};
