export const formatPostcode = (postcode: string | undefined): string => {
    if (!postcode) return '';
    // Add a text prefix to force Kepler.gl to treat it as a string
    // This should prevent histogram visualization
    return postcode.toString() + '‎';
};

export const formatCityName = (city: string): string => {
    const upperCity = city?.toUpperCase() || '';

    // Special case for VILLENEUVE D ASCQ
    if (upperCity === 'VILLENEUVE D ASCQ') {
        return "VILLENEUVE D'ASCQ";
    // Special case for 51000 - CHALONS EN CHAMPAGNE
    } else if (upperCity === '51000 - CHALONS EN CHAMPAGNE') {
        return "CHALONS-EN-CHAMPAGNE";
    }

    // Replace STE and ST with full words
    const cityWithFullWords = upperCity
        .replace(/\bSTE\b/g, 'SAINTE')
        .replace(/\bST\b/g, 'SAINT');

    // Split the string into words
    const words = cityWithFullWords.split(' ');
    const result: string[] = [];

    for (let i = 0; i < words.length; i++) {
        const currentWord = words[i];
        const nextWord = words[i + 1];

        // Add current word
        result.push(currentWord);

        // If there's a next word and current word is not LA/LE/LES, add hyphen
        if (nextWord && !['LA', 'LE', 'LES'].includes(currentWord)) {
            result.push('-');
        } else if (nextWord) {
            // If it is LA/LE/LES, add space
            result.push(' ');
        }
    }

    return result.join('');
};

export const getRegionFromPostalCode = (postalCode: string): string => {
    const prefix = postalCode.substring(0, 2);
    const prefixNum = parseInt(prefix, 10);

    // Handle special cases for overseas departments
    if (['971', '972', '973', '974', '976'].includes(postalCode.substring(0, 3))) {
        return 'Outre-Mer';
    }

    // Map department prefixes to regions
    if ([1, 3, 7, 15, 26, 38, 42, 43, 63, 69, 73, 74].includes(prefixNum)) {
        return 'Auvergne-Rhône-Alpes';
    } else if ([21, 25, 39, 58, 70, 71, 89, 90].includes(prefixNum)) {
        return 'Bourgogne-Franche-Comté';
    } else if ([22, 29, 35, 56].includes(prefixNum)) {
        return 'Bretagne';
    } else if ([18, 28, 36, 37, 41, 45].includes(prefixNum)) {
        return 'Centre-Val de Loire';
    } else if ([20].includes(prefixNum)) {
        return 'Corse';
    } else if ([8, 10, 51, 52, 54, 55, 57, 67, 68, 88].includes(prefixNum)) {
        return 'Grand Est';
    } else if ([2, 59, 60, 62, 80].includes(prefixNum)) {
        return 'Hauts-de-France';
    } else if ([75, 77, 78, 91, 92, 93, 94, 95].includes(prefixNum)) {
        return 'Île-de-France';
    } else if ([14, 27, 50, 61, 76].includes(prefixNum)) {
        return 'Normandie';
    } else if ([16, 17, 19, 23, 24, 33, 40, 47, 64, 79, 86, 87].includes(prefixNum)) {
        return 'Nouvelle-Aquitaine';
    } else if ([9, 11, 12, 30, 31, 32, 34, 46, 48, 65, 66, 81, 82].includes(prefixNum)) {
        return 'Occitanie';
    } else if ([44, 49, 53, 72, 85].includes(prefixNum)) {
        return 'Pays de la Loire';
    } else if ([4, 5, 6, 13, 83, 84].includes(prefixNum)) {
        return 'Provence-Alpes-Côte d\'Azur';
    }

    return 'Région inconnue';
};
