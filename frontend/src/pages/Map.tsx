import React, { useState, useEffect } from "react";
import { Tournament } from "../lib/api/types";
import { initializeDOMCleaner } from "../lib/domCleaner";
import { getMapConfig } from "../lib/map/config";
import { initializeSidebarCustomizer } from "../lib/sidebarCustomizer";
import { store } from "../lib/store";
import { formatPostcode } from "../lib/utils/address";
import { initializeDateTranslator } from "../lib/utils/date";
import { addDataToMap } from '@kepler.gl/actions';
import { loadTournaments } from "../lib/tournament/load";
import { CustomTooltipObserver } from "../components/CustomTooltipObserver";
import { getTournamentRows, tournamentFields } from "../lib/map/datasets";
import ErrorBoundary from "../components/ErrorBoundary";
import Meta from "../components/Meta";

export const Map: React.FC = () => {
    const [currentTournaments, setCurrentTournaments] = useState<Tournament[]>([]);
    const [pastCurrentTournaments, setPastCurrentTournaments] = useState<Tournament[]>([]);
    const [pastTournaments, setPastTournaments] = useState<Tournament[]>([]);
    const [isLoading, setIsLoading] = useState<boolean>(true);

    useEffect(() => {
        initializeSidebarCustomizer();
        initializeDateTranslator();
        initializeDOMCleaner();
    }, []);

    useEffect(() => {
        loadTournaments(setIsLoading, setCurrentTournaments, setPastCurrentTournaments, setPastTournaments);
    }, []);

    useEffect(() => {
        const observer = CustomTooltipObserver()

        observer.observe(document.body, {
            childList: true,
            subtree: true
        });

        return () => observer.disconnect();
    }, []);

    useEffect(() => {
        if (!currentTournaments?.length) {
            return
        }

        // const tournamentsWithoutCoordinates = currentTournaments.filter(
        //     t => !t.address?.latitude || !t.address?.longitude
        // );

        const tournamentsDataset = {
            fields: tournamentFields,
            rows: getTournamentRows([
                ...currentTournaments.filter(
                    t => t.address?.latitude && t.address?.longitude
                ),
                // ...tournamentsWithoutCoordinates.map(t => ({
                //     ...t,
                //     address: {
                //         ...t.address,
                //         latitude: 46.777138,
                //         longitude: 2.804568,
                //         approximate: true
                //     }
                // }))
            ].map(t => ({
                ...t,
                address: {
                    ...t.address,
                    postalCode: formatPostcode(t.address.postalCode)
                }
            })))
        };

        const pastCurrentTournamentsDataset = {
            fields: tournamentFields,
            rows: getTournamentRows([
                ...pastCurrentTournaments.filter(
                    t => t.address?.latitude && t.address?.longitude
                ),
            ].map(t => ({
                ...t,
                address: {
                    ...t.address,
                    postalCode: formatPostcode(t.address.postalCode)
                }
            })))
        };

        const pastTournamentsDataset = {
            fields: tournamentFields,
            rows: getTournamentRows([
                ...pastTournaments.filter(
                    t => t.address?.latitude && t.address?.longitude
                ),
            ].map(t => ({
                ...t,
                address: {
                    ...t.address,
                    postalCode: formatPostcode(t.address.postalCode)
                }
            })))
        };

        store.dispatch(
            addDataToMap({
                datasets: [
                    {
                        info: {
                            id: 'current_tournaments'
                        },
                        data: tournamentsDataset
                    },
                    {
                        info: {
                            id: 'past_current_tournaments'
                        },
                        data: pastCurrentTournamentsDataset
                    },
                    {
                        info: {
                            id: 'past_tournaments'
                        },
                        data: pastTournamentsDataset
                    }
                ],
                options: {
                    centerMap: false,
                    readOnly: false
                },
                config: getMapConfig(currentTournaments)
            })
        );
    }, [currentTournaments, pastTournaments]);

    // Wrap the entire component with error boundary
    return (
        <>
            <Meta isMainPage={true} />
            <ErrorBoundary isLoading={isLoading} />
        </>
    );
};