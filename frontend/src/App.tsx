/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { Routes, Route } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import Header from './components/Header';
import ErrorBoundary from './components/ErrorBoundary';
import AIAgent from './components/AIAgent';
import { useAuth } from './features/auth/hooks/useAuth';

// Lazy loading для компонентов
const Inventory = lazy(() => import('./components/Inventory'));
const Statistics = lazy(() => import('./components/Statistics'));
const AddPart = lazy(() => import('./components/AddPart'));
const AddCar = lazy(() => import('./components/AddCar'));
const DefectReport = lazy(() => import('./components/DefectReport'));
const AdminPanel = lazy(() => import('./components/AdminPanel'));
const Readme = lazy(() => import('./components/Readme'));
const Login = lazy(() => import('./components/Login'));
const Orders = lazy(() => import('./components/Orders'));

// Компонент загрузки для lazy loading
const LoadingSpinner = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-foreground"></div>
  </div>
);

function App() {
  const { user, isLoading, logout } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-foreground"></div>
      </div>
    );
  }

  const handleLogin = () => {
    // Redirect to inventory after login
    window.location.href = '/';
  };

  if (!user) {
    return <Login onLogin={handleLogin} />;
  }

  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-background pb-4">
        <Header user={user} onLogout={logout} />
        <Suspense fallback={<LoadingSpinner />}>
          <Routes>
            <Route path="/" element={<Inventory />} />
            <Route path="/inventory" element={<Inventory />} />
            <Route path="/statistics" element={<Statistics />} />
            <Route path="/add-car" element={<AddCar />} />
            <Route path="/defect-report" element={<DefectReport />} />
            <Route path="/add-part" element={<AddPart />} />
            <Route path="/admin" element={<AdminPanel />} />
            <Route path="/readme" element={<Readme user={user} />} />
            <Route path="/login" element={<Login onLogin={handleLogin} />} />
            <Route path="/orders" element={<Orders />} />
          </Routes>
        </Suspense>

        {/* ИИ-помощник */}
        <AIAgent />
      </div>
    </ErrorBoundary>
  );
}

export default App