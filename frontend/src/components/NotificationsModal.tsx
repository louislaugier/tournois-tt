import React from 'react';
// import {
//   Modal,
//   ModalContent,
//   ModalHeader,
//   ModalBody,
//   ModalFooter,
//   Button,
// } from '@heroui/react';

interface NotificationsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const NotificationsModal: React.FC<NotificationsModalProps> = ({ isOpen, onClose }) => {
  return (
    // <Modal 
    //   isOpen={isOpen} 
    //   onOpenChange={(open) => !open && onClose()} 
    //   size="md"
    //   style={{ zIndex: 10000 }}
    // >
    //   <ModalContent style={{ backgroundColor: 'white', color: 'black' }}>
    //     {(onClose) => (
    //       <>
    //         <ModalHeader style={{ backgroundColor: 'white', color: 'black', borderBottom: '1px solid #e5e7eb' }}>
    //           🔔 Notifications
    //         </ModalHeader>
    //         <ModalBody style={{ backgroundColor: 'white', color: 'black' }}>
    //           <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
    //             <div style={{ padding: '16px', backgroundColor: '#dbeafe', borderRadius: '8px', border: '1px solid #93c5fd' }}>
    //               <h4 style={{ fontWeight: '600', color: '#1e40af', margin: '0 0 8px 0' }}>
    //                 Nouveaux tournois disponibles
    //               </h4>
    //               <p style={{ fontSize: '14px', color: '#1d4ed8', margin: '0' }}>
    //                 3 nouveaux tournois ont été ajoutés dans votre région cette semaine.
    //               </p>
    //             </div>
                
    //             <div style={{ padding: '16px', backgroundColor: '#dcfce7', borderRadius: '8px', border: '1px solid #86efac' }}>
    //               <h4 style={{ fontWeight: '600', color: '#166534', margin: '0 0 8px 0' }}>
    //                 Mise à jour des données
    //               </h4>
    //               <p style={{ fontSize: '14px', color: '#15803d', margin: '0' }}>
    //                 Les données des tournois ont été mises à jour avec les dernières informations FFTT.
    //               </p>
    //             </div>
                
    //             <div style={{ padding: '16px', backgroundColor: '#fef3c7', borderRadius: '8px', border: '1px solid #fcd34d' }}>
    //               <h4 style={{ fontWeight: '600', color: '#92400e', margin: '0 0 8px 0' }}>
    //                 Maintenance programmée
    //               </h4>
    //               <p style={{ fontSize: '14px', color: '#b45309', margin: '0' }}>
    //                 Une maintenance est prévue dimanche de 2h à 4h du matin.
    //               </p>
    //             </div>
    //           </div>
    //         </ModalBody>
    //         <ModalFooter style={{ backgroundColor: 'white', color: 'black', borderTop: '1px solid #e5e7eb' }}>
    //           <Button color="danger" variant="light" onPress={onClose}>
    //             Fermer
    //           </Button>
    //           <Button color="primary" onPress={onClose}>
    //             Marquer comme lu
    //           </Button>
    //         </ModalFooter>
    //       </>
    //     )}
    //   </ModalContent>
    // </Modal>
    <></>
  );
};
