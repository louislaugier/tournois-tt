
export const formatDateDDMMYYYY = (date: Date | string) => {
  const d = new Date(date);
  const day = d.getDate().toString().padStart(2, '0');
  const month = (d.getMonth() + 1).toString().padStart(2, '0');
  const year = d.getFullYear();
  return `${day}/${month}/${year}`;
};

export function formatDateQueryParam(date: Date): string {
  return date.toISOString().split('.')[0];
}

export function setToMidnight(date: Date): Date {
  date.setHours(0, 0, 0, 0);
  return date;
}

export function setToYesterday(date: Date): Date {
  date.setDate(date.getDate() - 1);
  return date;
}

export function getYesterday(): Date {
  return setToYesterday(new Date())
}

export function getYesterdayMidnight(): Date {
  return setToMidnight(getYesterday())
}

export function getTodayMidnight(): Date {
  return setToMidnight(new Date())
}

export const normalizeDate = (date) => {
  const d = new Date(date);
  setToMidnight(d)
  return d.getTime();
};

export function getCurrentMonth() {
  const now = new Date();
  const currentMonth = now.getMonth() + 1; // JavaScript months are 0-based

  return currentMonth
}

export function getCurrentYear(date?: Date) {
  return (date || new Date()).getFullYear();
}

export function initializeDateTranslator(): void {
  
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