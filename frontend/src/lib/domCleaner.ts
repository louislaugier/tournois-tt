export function initializeDOMCleaner(): void {
  const selectors = [
    '.side-side-panel__header__bottom',
    '.filter-manager>div:first-child',
    '.filter-manager>div:nth-child(2)',
    '.filter-manager>div:nth-child(3)',
    '.layer-manager-header>button',
    '.side-panel__top__actions',
    '.side-panel__panel-header.side-side-panel__header.side-panel__panel-header',
    '.sc-cVzyXs',
    '.bottom-widget__y-axis',
    '.playback-controls',
    '.animation-control__time-display__bottom',
    '.maplibre-attribution-container',
    '.sc-ikkxIA.fOZEjb:nth-child(2)',
    '.map-control',
    '.data-source-selector.side-panel-section.data-source-selector',
    '.side-panel-section>div:nth-child(7)',
    '.side-panel-panel__label',
    '.edit-feature.toolbar-item.edit-feature',
    '.panel--header__action',
    '.select-geometry',
    '.popover-arrow-left',
    '.popover-pin',
    '.map-popover__layer-info>div:last-child'
  ];

  function removeHiddenElements() {
    selectors.forEach(selector => {
      document.querySelectorAll(selector).forEach(element => {
        if (element instanceof HTMLElement) {
          // Check if the element is actually hidden via CSS
          const computedStyle = window.getComputedStyle(element);
          if (computedStyle.display === 'none') {
            element.remove();
          }
        }
      });
    });
  }

  // Create a MutationObserver to watch for new elements
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'childList') {
        removeHiddenElements();
      }
    });
  });

  // Initial cleanup
  removeHiddenElements();

  // Start observing the document for DOM changes
  observer.observe(document.body, {
    childList: true,
    subtree: true
  });

  // Also run cleanup periodically to catch any elements that might have been missed
  setInterval(removeHiddenElements, 1000);
} 