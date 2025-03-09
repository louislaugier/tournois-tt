import React from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';
import App from './App';
import { handleErrorOverlayForEnv } from './lib/errors';

// Enhanced error logging and debugging
const initializeApp = () => {
  // Validate critical dependencies
  if (!React) {
    throw new Error('React is not loaded correctly');
  }

  if (!createRoot) {
    throw new Error('ReactDOM createRoot is not available');
  }

  // Detailed container initialization
  const container = document.getElementById('root');
  if (!container) {
    console.error('CRITICAL: Failed to find the root element');
    throw new Error('Failed to find the root element');
  }

  try {
    const root = createRoot(container);

    // Disable error overlay in production
    handleErrorOverlayForEnv()

    root.render(
      // <React.StrictMode>
      <App />
      // </React.StrictMode>
    );
  } catch (error) {
    console.error('CRITICAL: Failed to render React application', error);

    // Create an error display element
    const errorDisplay = document.createElement('div');
    errorDisplay.style.position = 'fixed';
    errorDisplay.style.top = '0';
    errorDisplay.style.left = '0';
    errorDisplay.style.width = '100%';
    errorDisplay.style.backgroundColor = 'red';
    errorDisplay.style.color = 'white';
    errorDisplay.style.padding = '10px';
    errorDisplay.style.zIndex = '9999';
    errorDisplay.innerHTML = `
      <h1>Application Initialization Error</h1>
      <p>An error occurred while starting the application.</p>
      <pre>${error?.toString()}</pre>
    `;

    throw error;
  }
};

// Global error handling
window.addEventListener('error', (event) => {
  console.error('Uncaught global error:', event.error);
});

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason);
});

// Suppress React development warnings
const originalError = console.error;
const originalWarn = console.warn;

const consoleFilter = function (callback: typeof console.error) {
  return function (msg: any, ...args: any[]) {
    if (typeof msg === 'string' && (
      msg.includes('findDOMNode') ||
      msg.includes('Warning: %s: Support for defaultProps') ||
      msg.includes('Warning: Setting defaultProps')
    )) {
      return;
    }
    callback.apply(console, [msg, ...args]);
  };
};

console.error = consoleFilter(originalError);
console.warn = consoleFilter(originalWarn);

// Attempt to initialize the app
try {
  initializeApp();
} catch (error) {
  console.error('Failed to initialize application', error);
} 