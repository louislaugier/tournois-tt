import React, { Component } from "react";
import { Spinner } from "./Spinner";
import Map from "./Map";
import { Button } from "@heroui/react";
import { NotificationsModal } from "./NotificationsModal";

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
            return (
                <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: '100vh',
                    backgroundColor: '#242730',
                    color: 'white',
                    fontFamily: 'system-ui, -apple-system, sans-serif'
                }}>
                    <h1 style={{ marginBottom: '20px', fontSize: '24px' }}>
                        Une erreur s'est produite
                    </h1>
                    <p style={{ marginBottom: '20px', textAlign: 'center' }}>
                        Veuillez actualiser la page pour continuer.
                    </p>
                    <button
                        onClick={() => window.location.reload()}
                        style={{
                            padding: '10px 20px',
                            backgroundColor: '#007bff',
                            color: 'white',
                            border: 'none',
                            borderRadius: '5px',
                            cursor: 'pointer',
                            fontSize: '16px'
                        }}
                    >
                        Actualiser la page
                    </button>
                </div>
            );
        }

        return this.props.children;
    }
}

export default (props: any) => {
    const { isLoading } = props
    const [sidebarOpen, setSidebarOpen] = React.useState(true)
    const [isModalOpen, setIsModalOpen] = React.useState(false)

    React.useEffect(() => {
        const checkSidebarState = () => {
            // Use the same logic as sidebarCustomizer.ts
            const sidePanel = document.querySelector('.side-panel--container') as HTMLElement
            if (sidePanel) {
                const isOpen = sidePanel.style.width !== '0px'
                setSidebarOpen(isOpen)
            }
        }

        // Check initially
        checkSidebarState()

        // Set up observer to watch for style changes on the sidebar
        const observer = new MutationObserver(checkSidebarState)
        observer.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: true,
            attributeFilter: ['style']
        })

        return () => {
            observer.disconnect()
        }
    }, [])

    return (
        <>
            <style>
                {`
                    @keyframes shake {
                        0%, 100% { transform: translateY(0); }
                        20% { transform: translateY(-2px); }
                        40% { transform: translateY(2px); }
                        60% { transform: translateY(-1px); }
                        80% { transform: translateY(1px); }
                    }
                `}
            </style>
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
                            <h1>Tournois FFTT</h1>
                            <p>Carte interactive des tournois de tennis de table en France. Trouvez les prochains tournois FFTT près de chez vous. Tri & recherche par type de tournoi, club organisateur, date, région, code postal, ville, dotation.</p>
                        </div>
                        <Map />
                        {/* Bell notification button */}
                        {/* <Button
                            isIconOnly
                            style={{
                                position: 'absolute',
                                top: '50px',
                                left: sidebarOpen ? '330px' : '15px',
                                zIndex: 1000,
                                background: '#242730',
                                border: 'none',
                                borderRadius: '50%',
                                width: '48px',
                                height: '48px',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                cursor: 'pointer',
                                boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)',
                                transition: 'all 0.2s ease',
                                animation: 'shake 1.5s ease-in-out infinite',
                            }}
                            onPress={() => setIsModalOpen(true)}
                            title="Notifications"
                            aria-label="Notifications"
                        >
                            <svg
                                width="20"
                                height="20"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="white"
                                strokeWidth="2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                            >
                                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
                                <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
                            </svg>
                        </Button> */}
                        
                        <div style={{
                            position: 'absolute',
                            bottom: 5,
                            background: '#242730',
                            height: 20,
                            width: 290,
                            left: 20,
                            color: 'white',
                            fontSize: 11,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'flex-start',
                            padding: '0 5px',
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis'
                        }}>Données mises à jour en temps réel. | <a style={{ color: 'white', margin: '0 3px' }} href="/cookies">Cookies</a> | <a style={{ color: 'white', margin: '0 5px', display: 'inline-flex', alignItems: 'center' }} href="/rss.xml" title="Flux RSS">
                                <svg style={{ width: '8px', height: '8px', marginRight: '2px' }} fill="currentColor" viewBox="0 0 24 24">
                                    <path d="M6.503 20.752c0 1.794-1.456 3.248-3.251 3.248S0 22.546 0 20.752s1.456-3.248 3.251-3.248 3.252 1.454 3.252 3.248zm-6.503-12.572v4.811c6.05.062 10.96 4.966 11.022 11.009h4.817c-.062-8.71-7.118-15.758-15.839-15.82zm0-3.368c10.58.046 19.152 8.594 19.183 19.188h4.817c-.03-13.231-10.755-23.954-24-24v4.812z" />
                                </svg>
                                RSS
                            </a></div>
                    </div>
                )}
                {/* <NotificationsModal
                    isOpen={true}
                    onClose={() => setIsModalOpen(false)}
                /> */}
            </ErrorBoundary>
        </>
    )
}