export function initializeSidebarCustomizer(): void {
  function cloneSidebarClose(): boolean {
    const originalElement = document.querySelector('.side-bar__close');
    
    if (originalElement) {
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
        const headerBottom = document.querySelector('.side-side-panel__header__bottom') as HTMLElement;
        
        if (sidePanel && sideBar && content && headerBottom) {
          const isOpen = sidePanel.style.width !== '0px';
          const width = isOpen ? '0px' : '340px';
          const opacity = isOpen ? '0' : '1';
          
          sideBar.style.width = width;
          sidePanel.style.width = width;
          content.style.opacity = opacity;
          
          // Toggle visibility of header bottom
          headerBottom.style.visibility = isOpen ? 'hidden' : 'visible';
        }
      });
      
      if (originalElement.parentNode) {
        originalElement.parentNode.insertBefore(cleanCopy, originalElement);
        // Remove the original element
        originalElement.remove();
        return true;
      } else {
        console.warn('No parent node found for the original element');
        return false;
      }
    } else {
      console.warn('Could not find .side-bar__close element in the DOM');
      return false;
    }
  }

  // Initial formatting and cloning with retry mechanism
  let retryCount = 0;
  const maxRetries = 5;
  const retryInterval = 1000; // 1 second

  function attemptClone() {
    if (retryCount >= maxRetries) {
      console.warn('Max retries reached for sidebar cloning');
      return;
    }

    const success = cloneSidebarClose();
    if (!success) {
      retryCount++;
      setTimeout(attemptClone, retryInterval);
    }
  }

  // Wait for DOM to be ready before cloning
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', attemptClone);
  } else {
    attemptClone();
  }
} 