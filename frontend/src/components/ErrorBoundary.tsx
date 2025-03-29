import { Component } from "react";
import { Spinner } from "./Spinner";
import Map  from "./Map";

// Error boundary component
class ErrorBoundary extends Component<{ children: React.ReactNode }, { hasError: boolean }> {
    constructor(props) {
        super(props);
        this.state = { hasError: false };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true };
    }

    componentDidCatch(error, errorInfo) {
        console.error("Uncaught error:", error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return <h1>Something went wrong. Please refresh the page.</h1>;
        }

        return this.props.children;
    }
}

export default (props: any) => {
    const { isLoading } = props
    return (
        <ErrorBoundary>
            {isLoading ? (
                <>
                    <Spinner />
                    <div style={{
                        position: 'fixed',
                        top: '50%',
                        left: '50%',
                        transform: 'translate(-50%, -50%)',
                        zIndex: 10000,
                        color: 'red',
                        fontSize: '24px'
                    }}>
                    </div>
                </>
            ) : (
                <div style={{ position: 'absolute', width: '100%', height: '100%' }}>
                    <div aria-hidden="true" style={{
                        position: 'absolute',
                        width: '1px',
                        height: '1px',
                        padding: '0',
                        margin: '-1px',
                        overflow: 'hidden',
                        clip: 'rect(0, 0, 0, 0)',
                        whiteSpace: 'nowrap',
                        border: '0'
                    }}>
                        <h1>Carte Interactive des Tournois de Tennis de Table en France</h1>
                        <p>
                            Bienvenue sur la carte interactive des tournois de tennis de table en France.
                            Trouvez facilement les tournois FFTT près de chez vous grâce à notre carte interactive.
                            Visualisez tous les tournois homologués par la Fédération Française de Tennis de Table.
                        </p>
                        <h2>Fonctionnalités de la Carte des Tournois FFTT</h2>
                        <ul>
                            <li>Visualisation de tous les tournois FFTT sur une carte de France interactive</li>
                            <li>Filtrage par date des tournois de ping-pong</li>
                            <li>Recherche par ville, département et code postal</li>
                            <li>Filtrage par montant de dotation et type de tournoi</li>
                            <li>Accès direct aux règlements des tournois homologués</li>
                            <li>Informations détaillées : club organisateur, dates, adresse, tables</li>
                            <li>Mise à jour en temps réel des compétitions de tennis de table</li>
                        </ul>
                        <h2>Prochains Tournois de Tennis de Table en France</h2>
                        <p>
                            Découvrez les prochains tournois de tennis de table homologués par la FFTT.
                            Notre carte est mise à jour en temps réel avec les dernières informations des clubs.
                            Filtrez par région, date ou type de compétition pour trouver le tournoi qui vous convient.
                        </p>
                        <h2>Compétitions de Tennis de Table par Région</h2>
                        <p>
                            Explorez les tournois de ping-pong dans toute la France :
                            Île-de-France, Auvergne-Rhône-Alpes, Nouvelle-Aquitaine, Occitanie,
                            Hauts-de-France, Grand Est, Provence-Alpes-Côte d'Azur,
                            Normandie, Bretagne, Pays de la Loire, Bourgogne-Franche-Comté,
                            Centre-Val de Loire, Corse et départements d'Outre-mer.
                        </p>
                        <h2>Informations sur les Tournois</h2>
                        <p>
                            Pour chaque tournoi de tennis de table, retrouvez :
                        </p>
                        <ul>
                            <li>Le nom et le type du tournoi FFTT</li>
                            <li>Les dates de début et de fin de la compétition</li>
                            <li>Le montant de la dotation et des récompenses</li>
                            <li>L'adresse complète du gymnase ou de la salle</li>
                            <li>Le club organisateur et son numéro d'affiliation</li>
                            <li>Le règlement officiel du tournoi homologué</li>
                            <li>Le nombre de tables disponibles</li>
                        </ul>
                        <h2>À Propos de la Carte des Tournois</h2>
                        <p>
                            Service gratuit de visualisation des tournois de tennis de table en France.
                            Données officielles issues de la FFTT, mises à jour automatiquement.
                            Interface intuitive pour trouver rapidement les compétitions près de chez vous.
                        </p>
                    </div>
                   <Map />
                    <div style={{
                        position: 'absolute',
                        bottom: 5,
                        background: '#242730',
                        height: 20,
                        width: 300,
                        left: 20,
                        color: 'white',
                        fontSize: 10,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                    }}>Données mises à jour en temps réel. | <a style={{ color: 'white', marginLeft: '5px' }} href="/cookies">Cookies et vie privée</a></div>
                </div>
            )}
        </ErrorBoundary>
    )
}