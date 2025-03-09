import franceBordersRaw from './../../assets/metropole-et-outre-mer.json';

export const franceBorders = typeof franceBordersRaw === 'string' ? JSON.parse(franceBordersRaw) : franceBordersRaw;
 