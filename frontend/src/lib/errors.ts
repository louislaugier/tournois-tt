export function handleErrorOverlayForEnv() {
    if (process.env.NODE_ENV === 'production') {
        console.error = () => { };
        window.addEventListener('error', (e) => {
          e.stopPropagation();
          e.preventDefault();
        });
        window.addEventListener('unhandledrejection', (e) => {
          e.stopPropagation();
          e.preventDefault();
        });
      }
      
}