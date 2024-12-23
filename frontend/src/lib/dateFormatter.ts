export function initializeDateFormatter(): void {
  function formatDate(element: HTMLElement): void {
    const text: string = element.textContent || '';
    // Split by space to separate date and time
    const [datePart] = text.split(' ');
    if (!datePart) return;
    
    const [month, day, year] = datePart.split('/');
    if (parseInt(month) <= 12 && !element.hasAttribute('data-formatted')) {
      element.textContent = `${day}/${month}/${year}`;
      element.setAttribute('data-formatted', 'true');
    }
  }

  function formatAllDates(): void {
    const dateElements: NodeListOf<HTMLElement> = document.querySelectorAll('.animation-control__time-domain span');
    dateElements.forEach(formatDate);
  }

  // Initial formatting
  formatAllDates();

  // Create an observer instance to watch for changes
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'characterData') {
        const element = mutation.target.parentElement as HTMLElement;
        if (element && element.closest('.animation-control__time-domain')) {
          element.removeAttribute('data-formatted');
          formatDate(element);
        }
      } else if (mutation.type === 'childList') {
        formatAllDates();
      }
    });
  });

  // Observe the entire document body for changes
  observer.observe(document.body, {
    characterData: true,
    childList: true,
    subtree: true
  });
} 