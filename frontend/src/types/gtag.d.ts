/**
 * Type definitions for Google Analytics gtag function
 */

declare global {
  interface Window {
    gtag: (
      command: 'config' | 'event' | 'js' | 'set',
      targetId: string | Date,
      config?: {
        [key: string]: any;
      }
    ) => void;
  }
}

declare function gtag(
  command: 'config' | 'event' | 'js' | 'set',
  targetId: string | Date,
  config?: {
    [key: string]: any;
  }
): void;

export {};
