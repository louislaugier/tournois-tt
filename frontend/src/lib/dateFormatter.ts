export function initializeDateFormatter(): void {
  function formatDate(element: HTMLElement): void {
    const text: string = element.textContent || '';
    // Split by space to separate date and time
    const [datePart] = text.split(' ');
    if (!datePart) return;
    
    // For all elements, convert from MM/DD/YYYY to DD/MM/YYYY
    const [month, day, year] = datePart.split('/');
    if (parseInt(month) <= 12 && !element.hasAttribute('data-formatted')) {
      element.textContent = `${day}/${month}/${year}`;
      element.setAttribute('data-formatted', 'true');
    }
  }

  function translateMonths(): void {
    const monthTranslations: { [key: string]: string } = {
      'January': 'Janvier',
      'February': 'Février',
      'March': 'Mars',
      'April': 'Avril',
      'May': 'Mai',
      'June': 'Juin',
      'July': 'Juillet',
      'August': 'Août',
      'September': 'Septembre',
      'October': 'Octobre',
      'November': 'Novembre',
      'December': 'Décembre'
    };

    const timeSliderElements = document.querySelectorAll('.time-slider-marker text');
    timeSliderElements.forEach((element) => {
      const text = element.textContent || '';
      const translatedMonth = monthTranslations[text];
      if (translatedMonth && !element.hasAttribute('data-translated')) {
        element.textContent = translatedMonth;
        element.setAttribute('data-translated', 'true');
      }
    });
  }

  function formatAllDates(): void {
    const dateElements: NodeListOf<HTMLElement> = document.querySelectorAll('.animation-control__time-domain span, .animation-control__time-display__top');
    dateElements.forEach(formatDate);
    translateMonths();
  }

  // Initial formatting and cloning
  console.log('Starting initialization...');
  formatAllDates();

  // Create an observer instance to watch for changes
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'characterData') {
        const element = mutation.target.parentElement as HTMLElement;
        if (element && (element.closest('.animation-control__time-domain') || element.closest('.animation-control__time-display__top'))) {
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

  // Add periodic check for new elements that need translation
  setInterval(translateMonths, 1000);
} 