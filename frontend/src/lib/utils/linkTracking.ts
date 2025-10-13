/**
 * Utilitaires pour le tracking des liens externes et l'ajout de paramètres UTM
 */

export interface LinkTrackingOptions {
  utmSource?: string;
  utmMedium?: string;
  utmCampaign?: string;
  utmContent?: string;
  trackClick?: boolean;
  linkText?: string;
}

/**
 * Crée un lien externe avec des paramètres UTM et du tracking
 */
export const createExternalLink = (url: string, options: LinkTrackingOptions = {}): string => {
  const {
    utmSource = 'tournois-tt.fr',
    utmMedium = 'website',
    utmCampaign,
    utmContent,
    trackClick = true
  } = options;

  // Ne pas ajouter d'UTM aux liens internes
  if (url.startsWith('/') || url.includes('tournois-tt.fr')) {
    return url;
  }

  const utmParams = new URLSearchParams();
  utmParams.set('utm_source', utmSource);
  utmParams.set('utm_medium', utmMedium);
  
  if (utmCampaign) {
    utmParams.set('utm_campaign', utmCampaign);
  }
  
  if (utmContent) {
    utmParams.set('utm_content', utmContent);
  }

  const separator = url.includes('?') ? '&' : '?';
  const finalUrl = `${url}${separator}${utmParams.toString()}`;

  return finalUrl;
};

/**
 * Track un clic sur un lien externe
 */
export const trackExternalClick = (url: string, linkText: string = '', eventCategory: string = 'external_link') => {
  // Google Analytics 4
  if (typeof gtag !== 'undefined') {
    gtag('event', 'click', {
      event_category: eventCategory,
      event_label: linkText || url,
      value: url,
      link_url: url,
      link_text: linkText
    });
  }

  // Google Analytics Universal (fallback)
  if (typeof ga !== 'undefined') {
    ga('send', 'event', eventCategory, 'click', linkText || url);
  }

  // Console log pour debug
  console.log(`External link clicked: ${linkText} -> ${url}`);
};

/**
 * Crée un gestionnaire de clic pour les liens externes
 */
export const createExternalClickHandler = (url: string, linkText: string = '', eventCategory: string = 'external_link') => {
  return (event: React.MouseEvent<HTMLAnchorElement>) => {
    trackExternalClick(url, linkText, eventCategory);
  };
};

/**
 * Constantes pour les campagnes UTM
 */
export const UTM_CAMPAIGNS = {
  SOCIAL_INSTAGRAM: 'social_instagram',
  FFTT_RULES: 'fftt_rules',
  FFTT_SIGNUP: 'fftt_signup',
  SOCIAL_RSS: 'social_rss'
} as const;

/**
 * Constantes pour le contenu UTM
 */
export const UTM_CONTENT = {
  FOOTER: 'footer',
  TOOLTIP: 'tooltip',
  FEED_LIST: 'feed_list',
  FEED_DETAIL: 'feed_detail',
  MAP_TOOLTIP: 'map_tooltip'
} as const;
