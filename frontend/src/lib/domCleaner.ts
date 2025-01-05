export function initializeDOMCleaner(): void {
  const selectors = [
    // '.bottom-widget__y-axis',
    // '.playback-controls',
    // '.animation-control__time-display__bottom',
    '.maplibre-attribution-container',
    '.map-control',
    '.select-geometry',
    '.popover-arrow-left',
    '.popover-pin',
    '.map-popover__layer-info>div:last-child',
    '.edit-feature.toolbar-item.edit-feature'
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