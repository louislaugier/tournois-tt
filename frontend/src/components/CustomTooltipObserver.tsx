export const CustomTooltipObserver = (): MutationObserver => {
    return new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            if (mutation.type === 'childList') {
                // Handle map-popover__layer-name
                document.querySelectorAll('.map-popover__layer-name').forEach((el) => {
                    if (el.textContent?.includes('Tournoi')) {
                        (el as HTMLElement).style.display = 'none';
                        // Find and style jlDYGb elements within this tooltip
                        const tooltipContainer = el.closest('.map-popover');
                        tooltipContainer?.querySelectorAll('.jlDYGb').forEach((jlElement) => {
                            (jlElement as HTMLElement).style.marginBottom = '12.5px';
                        });
                    } else {
                        (el as HTMLElement).style.display = '';
                        // Reset margin for non-Tournoi tooltips
                        const tooltipContainer = el.closest('.map-popover');
                        tooltipContainer?.querySelectorAll('.jlDYGb').forEach((jlElement) => {
                            (jlElement as HTMLElement).style.marginBottom = '';
                        });
                    }
                });

                // Handle kWuINq separately
                document.querySelectorAll('.kWuINq').forEach((el) => {
                    const tooltipContainer = el.closest('.map-popover');
                    const tournamentName = tooltipContainer?.querySelector('.row__name')?.textContent;
                    if (tournamentName === 'Nom du tournoi') {
                        (el as HTMLElement).style.display = 'none';
                    } else {
                        (el as HTMLElement).style.display = '';
                    }
                });

                // Add click tracking for external links in tooltips (beacon to avoid navigation loss)
                document.querySelectorAll('.map-popover a[href^="http"]').forEach((link) => {
                    const href = link.getAttribute('href');
                    if (href && !href.includes('tournois-tt.fr') && !(link as HTMLElement).dataset.tracked) {
                        // ensure new tab to give GA time
                        link.setAttribute('target', '_blank');
                        link.setAttribute('rel', 'noopener noreferrer');
                        (link as HTMLElement).dataset.tracked = '1';
                        link.addEventListener('click', () => {
                            if (typeof window !== 'undefined' && window.gtag) {
                                window.gtag('event', 'click', {
                                    event_category: 'external_link',
                                    event_label: 'Map Tooltip External Link',
                                    link_url: href,
                                    link_text: link.textContent || '',
                                    content_group: 'map_tooltip',
                                    link_source: 'rules',
                                    transport_type: 'beacon'
                                });
                            }
                        });
                    }
                });
            }
        });
    });
}