import React, { useState } from 'react';
import { ThemeProvider } from 'styled-components';
import { IntlProvider } from 'react-intl';
import { messages } from '@kepler.gl/localization';
import { theme } from '@kepler.gl/styles';
import StyledModal from '@kepler.gl/components/dist/common/modal';
import { InputLight, PanelLabel, StyledModalContent, Button } from '@kepler.gl/components/dist/common/styled-components';

interface NotificationsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const NotificationsModal: React.FC<NotificationsModalProps> = ({ isOpen, onClose }) => {
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [result, setResult] = useState<'created' | 'updated' | null>(null);

  const handleSubmit = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!email || status === 'loading') return;
    setStatus('loading');
    fetch(process.env.API_URL ? `${process.env.API_URL}/v1/newsletter` : '/api/v1/newsletter', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email })
    })
      .then(async (res) => {
        const data = await res.json().catch(() => ({} as any));
        const apiResult: 'created' | 'updated' | undefined =
          (data && data.result) || (res.status === 204 ? 'updated' : res.status === 201 ? 'created' : undefined);
        if (res.ok) {
          setResult(apiResult === 'updated' ? 'updated' : 'created');
          setStatus('success');
        } else {
          setResult(null);
          setStatus('error');
        }
      })
      .catch(() => setStatus('error'));
  };

  return (
    <IntlProvider locale="en" messages={messages.en}>
    <ThemeProvider theme={theme}>
      <style>{`.modal--title{font-family:${theme.fontFamily};font-size:18px}`}</style>
      <StyledModal
        isOpen={isOpen}
        onCancel={onClose}
        onRequestClose={onClose}
        shouldCloseOnEsc
        shouldCloseOnOverlayClick
        // we'll render actions inside the content to match Kepler's example (left-aligned primary button)
        footer={false}
        title="Soyez notifié des prochaines fonctionnalités"
        cssStyle={`overflow: hidden;`}
        confirmButton={{
          large: true,
          disabled: !email,
          children: "S'abonner"
        }}
        theme={theme}
      >
        <StyledModalContent>
          <div style={{ color: theme.textColor, fontFamily: theme.fontFamily, fontSize: 15, lineHeight: 1.6, marginBottom: 16 }}>
            <div style={{ marginBottom: 6 }}>
              Prochaines évolutions :
            </div>
            <ul style={{ margin: 0, paddingLeft: 16 }}>
              <li>Affichage du lien d’inscription aux tableaux du tournoi dans l’infobulle quand il est disponible</li>
              <li>Fonctionnalité d'alertes push/sms/email dès l’ouverture des inscriptions aux tableaux du tournoi (notifications par département, région ou partout)</li>
              <li>Applications mobile iOS et Android (en cours de développement)</li>
            </ul>
          </div>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 10, paddingTop: 6, fontFamily: theme.fontFamily, fontSize: 15 }}>
            <PanelLabel style={{ marginBottom: 4 }}>Adresse email</PanelLabel>
            <InputLight
              type="email"
              placeholder="Votre adresse email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              style={{ width: 420 }}
            />
            <Button
              type="submit"
              large
              cta
              disabled={!email || status === 'loading'}
              style={{ width: 172, marginTop: 12, fontSize: 14 }}
            >
              {status === 'loading' ? 'Envoi…' : "S'abonner"}
            </Button>
            <div aria-live="polite" style={{ marginTop: 8, minHeight: 20, color: status === 'success' ? theme.activeColor : status === 'error' ? theme.errorColor : 'inherit' }}>
              {status === 'success' && result === 'created' && "Inscription confirmée. Bienvenue !"}
              {status === 'success' && result === 'updated' && "Vous étiez déjà inscrit. Vos préférences ont été mises à jour."}
              {status === 'error' && "Une erreur est survenue. Réessayez dans quelques instants."}
            </div>
          </form>
        </StyledModalContent>
      </StyledModal>
    </ThemeProvider>
    </IntlProvider>
  );
};
