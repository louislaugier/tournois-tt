
import React from 'react';

const Cookies: React.FC = () => {
  return (
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
          Ce site utilise Google Analytics et Google AdSense pour améliorer votre expérience et
          afficher des publicités pertinentes. Ces services peuvent utiliser des cookies pour
          fonctionner correctement.
        </p>
        <p>
          Les cookies Google Analytics nous permettent d'analyser l'utilisation du site de manière
          anonyme. Les cookies Google AdSense sont utilisés pour personnaliser les annonces en
          fonction de vos centres d'intérêt.
        </p>
      </section>

      <section style={{ marginBottom: '30px' }}>
        <h2>Données collectées</h2>
        <p>
          Les données collectées par Google Analytics et AdSense peuvent inclure :
        </p>
        <ul style={{ marginLeft: '20px', lineHeight: '1.5' }}>
          <li>Informations sur votre navigateur</li>
          <li>Durée de visite</li>
          <li>Pages visitées</li>
          <li>Localisation approximative</li>
          <li>Centres d'intérêt pour la publicité</li>
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
  );
};

export default Cookies; 