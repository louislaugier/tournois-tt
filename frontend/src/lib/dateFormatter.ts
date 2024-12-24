export function initializeDateFormatter(): void {
  function cloneSidebarClose(): void {
    console.log('Attempting to find .side-bar__close element...');
    const originalElement = document.querySelector('.side-bar__close');
    console.log('Original element found:', originalElement);
    
    if (originalElement) {
      console.log('Creating clone of element...');
      const copy = originalElement.cloneNode(true) as HTMLElement;
      copy.classList.add('side-bar__close--cloned');
      
      // Remove all event listeners by replacing the element with its clone
      const cleanCopy = copy.cloneNode(true) as HTMLElement;
      
      // Add new click event to show alert
      cleanCopy.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        alert('Clicked the cloned button!');
      });
      
      console.log('Clone created and modified:', cleanCopy);
      
      if (originalElement.parentNode) {
        console.log('Parent node found, inserting clone...');
        originalElement.parentNode.insertBefore(cleanCopy, originalElement);
        // Remove the original element
        originalElement.remove();
        console.log('Clone inserted successfully and original removed');
      } else {
        console.warn('No parent node found for the original element');
      }
    } else {
      console.warn('Could not find .side-bar__close element in the DOM');
    }
  }

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

  function formatAllDates(): void {
    const dateElements: NodeListOf<HTMLElement> = document.querySelectorAll('.animation-control__time-domain span, .animation-control__time-display__top');
    dateElements.forEach(formatDate);
  }

  // Initial formatting and cloning
  console.log('Starting initialization...');
  formatAllDates();
  
  // Wait for DOM to be ready before cloning
  setTimeout(() => {
    console.log('Attempting to clone sidebar after delay...');
    cloneSidebarClose();
  }, 1000);

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
} 