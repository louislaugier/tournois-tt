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
            }
        });
    });
}