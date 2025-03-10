import { Routes, Route, Navigate, useLocation, BrowserRouter } from "react-router-dom";
import Cookies from "../pages/CookiePolicy";
import { Map } from "../pages/Map";
import { useEffect } from "react";
import React from "react";

const Router: React.FC = () => {
    const location = useLocation();

    useEffect(() => {
        if (location.pathname === '/') document.body.style.overflow = 'hidden'; // no scroll on map page

        return () => {
            document.body.style.overflow = '';
        };
    }, [location]);

    return (
        <Routes>
            <Route path="/" element={<Map />} />
            <Route path="/cookies" element={<Cookies />} />
            <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
    );
};

export default () => (
    <BrowserRouter>
        <Router />
    </BrowserRouter>
);