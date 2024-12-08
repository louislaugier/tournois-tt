import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import './styles.css';

// Suppress React development warnings
const originalError = console.error;
const originalWarn = console.warn;

// Override error logging before React loads
const consoleFilter = function(callback: typeof console.error) {
  return function(msg: any, ...args: any[]) {
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

const container = document.getElementById('root');
if (!container) throw new Error('Failed to find the root element');

const root = createRoot(container);
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
); 