declare module 'classnames' {
  type ClassValue = string | number | boolean | undefined | null | Record<string, any>;
  
  interface ClassNamesFn {
    (...classes: ClassValue[]): string;
    (...classes: ClassValue[][]): string;
    (arg: Record<string, any>): string;
  }

  const classNames: ClassNamesFn;
  export default classNames;
  export = classNames;
} 