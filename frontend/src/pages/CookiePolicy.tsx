import React from 'react';
import { Helmet } from 'react-helmet';

const CookiePolicy: React.FC = () => {
  return (
    <>
      <style>
        {`
          body {
            overflow: visible;
          }
        `}
      </style>
      <Helmet>
        <title>Politique de Confidentialité et Cookies | Carte des Tournois FFTT</title>
        <meta name="description" content="Politique de confidentialité et utilisation des cookies pour la carte des tournois de tennis de table. Informations sur l'utilisation de Google Analytics et la protection de vos données." />
        <meta name="robots" content="index, follow" />
        <link rel="canonical" href="https://tournois-tt.fr/cookies" />
      </Helmet>
      <div style={{
        padding: '20px',
        color: 'white',
        backgroundColor: '#242730',
        minHeight: '100vh',
        fontFamily: 'Arial, sans-serif'
      }}>
        <h1>Politique de confidentialité et cookies</h1>
        
        <section style={{ marginBottom: '30px' }}>
          <h2>Cookies et publicités</h2>
          <p>
            Ce site utilise Google Analytics pour récolter des statistiques anonymes. Ce service peut utiliser des cookies pour fonctionner correctement.
          </p>
        </section>

        <section style={{ marginBottom: '30px' }}>
          <h2>Données collectées</h2>
          <p>
            Les données collectées par Google Analyticspeuvent inclure :
          </p>
          <ul style={{ marginLeft: '20px', lineHeight: '1.5' }}>
            <li>Informations sur votre navigateur</li>
            <li>Adresse IP</li>
            <li>Durée de visite</li>
            <li>Localisation approximative</li>
            <li>Intérêt pour la publicité</li>
          </ul>
        </section>

        <section style={{ marginBottom: '30px' }}>
          <h2>Non-affiliation</h2>
          <p>
            Ce site n'est pas affilié à la Fédération Française de Tennis de Table (FFTT).
            Il s'agit d'un service indépendant qui utilise les données publiques de la FFTT
            pour fournir une visualisation des tournois.
          </p>
        </section>
        <a 
          href="/"
          style={{
            display: 'inline-block',
            marginTop: '20px',
            color: 'white',
            textDecoration: 'none',
            padding: '10px 20px',
            backgroundColor: '#1a1b1e',
            borderRadius: '4px'
          }}
        >
          Retour à la carte
        </a>
      </div>
    </>
  );
};

export default CookiePolicy; 