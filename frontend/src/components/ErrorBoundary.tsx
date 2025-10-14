import React, { Component } from "react";
import { Spinner } from "./Spinner";
import Map from "./Map";
import { MapControlButton } from "@kepler.gl/components";
import { ThemeProvider } from "styled-components";
import { theme } from "@kepler.gl/styles";
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
                <ThemeProvider theme={theme}>
                    <div style={{
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center',
                        height: '100vh',
                        backgroundColor: '#242730',
                        color: theme.titleColorLT,
                        fontFamily: theme.fontFamily
                    }}>
                        <div style={{
                            background: '#29323C',
                            border: '1px solid #3A414C',
                            boxShadow: '0 6px 12px 0 rgba(0,0,0,0.16)',
                            borderRadius: 2,
                            padding: '24px 28px',
                            maxWidth: 520,
                            textAlign: 'center'
                        }}>
                            <h1 style={{ margin: 0, marginBottom: 12, fontSize: 18, color: theme.textColorHl }}>Une erreur s'est produite</h1>
                            <p style={{ margin: 0, marginBottom: 20, color: theme.textColor }}>Veuillez actualiser la page pour continuer.</p>
                            <button
                                onClick={() => window.location.reload()}
                                className="map-control-button"
                                style={{
                                    backgroundColor: theme.primaryBtnBgd,
                                    color: theme.primaryBtnColor,
                                    border: theme.primaryBtnBorder,
                                    borderRadius: theme.primaryBtnRadius,
                                    padding: '10px 16px',
                                    fontSize: theme.primaryBtnFontSizeDefault,
                                    cursor: 'pointer'
                                }}
                            >
                                Actualiser la page
                            </button>
                        </div>
                    </div>
                </ThemeProvider>
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
                    @keyframes pulse {
                        0%, 100% { box-shadow: 0 6px 12px 0 rgba(0, 0, 0, 0.16); }
                        50% { box-shadow: 0 6px 12px 0 rgba(0, 0, 0, 0.16), 0 0 0 6px rgba(31, 186, 214, 0.28); }
                    }
                    .notification-bell-button { animation: pulse 2.4s ease-in-out infinite; }
                    .notification-bell-button:hover {
                        animation: none;
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
                        <ThemeProvider theme={theme}>
                            <MapControlButton
                                onClick={() => {
                                    if (typeof window !== 'undefined' && (window as any).gtag) {
                                        (window as any).gtag('event', 'click', {
                                            event_category: 'ui_action',
                                            event_label: 'bell_notifications',
                                            link_type: 'notification_bell',
                                            content_group: 'ui',
                                        });
                                    }
                                    setIsModalOpen(true)
                                }}
                                className="map-control-button notification-bell-button"
                                style={{
                                    position: 'absolute',
                                    top: '50px',
                                    left: sidebarOpen ? '330px' : '15px',
                                    zIndex: 1000,
                                    transition: 'left 0.2s ease',
                                }}
                            >
                                <svg 
                                    width="16" 
                                    height="16" 
                                    fill="currentColor" 
                                    viewBox="0 0 24 24"
                                >
                                    <path d="M12 22c1.1 0 2-.9 2-2h-4c0 1.1.9 2 2 2zm6-6v-5c0-3.07-1.64-5.64-4.5-6.32V4c0-.83-.67-1.5-1.5-1.5s-1.5.67-1.5 1.5v.68C7.63 5.36 6 7.92 6 11v5l-2 2v1h16v-1l-2-2z"/>
                                </svg>
                            </MapControlButton>
                        </ThemeProvider>

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
                        }}>Données mises à jour en temps réel. | <a style={{ color: 'white', margin: '0 3px' }} href="/a-propos">A propos                            </a> | <a style={{ color: 'white', margin: '0 5px', display: 'inline-flex', alignItems: 'center' }} href="https://instagram.com/tournoistt?utm_source=tournois-tt.fr&utm_medium=website&utm_campaign=social_instagram&utm_content=footer" title="Instagram @tournoistt" target="_blank" rel="noopener noreferrer" onClick={() => {
                            if (typeof window !== 'undefined' && window.gtag) {
                                window.gtag('event', 'click', {
                                    event_category: 'social_link',
                                    event_label: 'Instagram Footer',
                                    value: 'https://instagram.com/tournoistt',
                                    link_type: 'instagram',
                                    content_group: 'footer'
                                });
                            }
                        }}>
                                <svg style={{ width: '14px', height: '14px' }} fill="currentColor" viewBox="0 0 24 24">
                                    <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838c-3.403 0-6.162 2.759-6.162 6.162s2.759 6.163 6.162 6.163 6.162-2.759 6.162-6.163c0-3.403-2.759-6.162-6.162-6.162zm0 10.162c-2.209 0-4-1.79-4-4 0-2.209 1.791-4 4-4s4 1.791 4 4c0 2.21-1.791 4-4 4zm6.406-11.845c-.796 0-1.441.645-1.441 1.44s.645 1.44 1.441 1.44c.795 0 1.439-.645 1.439-1.44s-.644-1.44-1.439-1.44z" />
                                </svg>
                            </a> <a style={{ color: 'white', margin: '0 3px', display: 'inline-flex', alignItems: 'center' }} href="/rss.xml" title="Flux RSS">
                                <svg style={{ width: '12px', height: '12px' }} fill="currentColor" viewBox="0 0 24 24">
                                    <path d="M6.503 20.752c0 1.794-1.456 3.248-3.251 3.248S0 22.546 0 20.752s1.456-3.248 3.251-3.248 3.252 1.454 3.252 3.248zm-6.503-12.572v4.811c6.05.062 10.96 4.966 11.022 11.009h4.817c-.062-8.71-7.118-15.758-15.839-15.82zm0-3.368c10.58.046 19.152 8.594 19.183 19.188h4.817c-.03-13.231-10.755-23.954-24-24v4.812z" />
                                </svg>
                            </a></div>

                        <NotificationsModal 
                            isOpen={isModalOpen} 
                            onClose={() => setIsModalOpen(false)} 
                        />
                    </div>
                )}
              
            </ErrorBoundary>
        </>
    )
}