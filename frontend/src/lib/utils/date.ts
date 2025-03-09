
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