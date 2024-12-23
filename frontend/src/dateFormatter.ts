export function initializeDateFormatter(): void {
  function swapMonthDay(): void {
    const dateElements: NodeListOf<HTMLElement> = document.querySelectorAll('.animation-control__time-display__top');
    dateElements.forEach((element: HTMLElement) => {
      const text: string = element.textContent || '';
      const [month, day, year] = text.split('/');
      if (parseInt(month) <= 12 && !element.hasAttribute('data-formatted')) {
        element.textContent = `${day}/${month}/${year}`;
        element.setAttribute('data-formatted', 'true');
      }
    });
  }

  setInterval(swapMonthDay, 100);
} 