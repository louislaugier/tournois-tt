export const Spinner = () => {
    return (
        <div
            id="loading-spinner"
            style={{
                position: 'fixed',
                top: 0,
                left: 0,
                width: '100vw',
                height: '100vh',
                backgroundColor: '#0E0E0E',
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                zIndex: 9999
            }}
        >
            <div
                style={{
                    width: '100px',
                    height: '100px',
                    border: '10px solid #1FBAD6', // Base turquoise
                    borderTop: '10px solid rgba(31, 186, 214, 0.3)', // Lighter turquoise for spinning effect
                    borderRight: '10px solid rgba(31, 186, 214, 0.6)', // Slightly darker turquoise
                    borderBottom: '10px solid rgba(31, 186, 214, 0.8)', // Even darker turquoise
                    borderRadius: '50%',
                    animation: 'spin 1s linear infinite'
                }}
            />
            <style>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
        </div>
    );
};