export function initializeSidebarCustomizer(): void {
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

        // Toggle sidebar state
        const sidePanel = document.querySelector('.side-panel--container') as HTMLElement;
        const sideBar = document.querySelector('.side-bar') as HTMLElement;
        const content = document.querySelector('.side-panel__content') as HTMLElement;
        
        if (sidePanel && sideBar && content) {
          const isOpen = sidePanel.style.width !== '0px';
          const width = isOpen ? '0px' : '340px';
          const opacity = isOpen ? '0' : '1';
          
          sideBar.style.width = width;

          sidePanel.style.width = width;

          content.style.opacity = opacity;

        }
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

  // Initial formatting and cloning
  console.log('Starting initialization...');
  
  // Wait for DOM to be ready before cloning
  setTimeout(() => {
    console.log('Attempting to clone sidebar after delay...');
    cloneSidebarClose();
  }, 1000);
 
} 