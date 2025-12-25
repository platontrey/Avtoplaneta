/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { Routes, Route } from 'react-router-dom';
import { lazy, Suspense, useState } from 'react';
import Header from './components/Header';
import ErrorBoundary from './components/ErrorBoundary';
import AIAgent from './components/AIAgent';
import PWAInstallPrompt from './components/PWAInstallPrompt';
import MobileBottomNav from './components/MobileBottomNav';
import { useAuth } from './features/auth/hooks/useAuth';

// Lazy loading для компонентов
const Inventory = lazy(() => import('./components/Inventory'));
const Statistics = lazy(() => import('./components/Statistics'));
const AddPart = lazy(() => import('./components/AddPart'));
const AddCar = lazy(() => import('./components/AddCar'));
const DefectReport = lazy(() => import('./components/DefectReport'));
const AdminPanel = lazy(() => import('./components/AdminPanel'));
const Readme = lazy(() => import('./components/Readme'));
const OperatorInstructions = lazy(() => import('./components/OperatorInstructions'));
const ManagerInstructions = lazy(() => import('./components/ManagerInstructions'));
const Login = lazy(() => import('./components/Login'));
const Orders = lazy(() => import('./components/Orders'));
const MessagesPage = lazy(() => import('./components/MessagesPage'));

// Компонент загрузки для lazy loading
const LoadingSpinner = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-foreground"></div>
  </div>
);

function App() {
  const { user, isLoading, logout } = useAuth();
  const [isAIOpen, setIsAIOpen] = useState(false);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-foreground"></div>
      </div>
    );
  }

  if (!user) {
    return <Login />;
  }

  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-background pb-20 md:pb-4">
        {/* Предложение установки PWA сверху */}
        <PWAInstallPrompt />

        <Header user={user} onLogout={logout} />
        <Suspense fallback={<LoadingSpinner />}>
          <Routes>
             <Route path="/" element={<Inventory />} />
             <Route path="/inventory" element={<Inventory />} />
             <Route path="/statistics" element={<Statistics />} />
             <Route path="/messages" element={<MessagesPage />} />
             <Route path="/add-car" element={<AddCar />} />
             <Route path="/defect-report" element={<DefectReport />} />
             <Route path="/add-part" element={<AddPart />} />
             <Route path="/admin" element={<AdminPanel />} />
             <Route path="/readme" element={<Readme user={user} />} />
             <Route path="/operator-instructions" element={<OperatorInstructions />} />
             <Route path="/manager-instructions" element={<ManagerInstructions />} />
             <Route path="/login" element={<Login />} />
             <Route path="/orders" element={<Orders />} />
            </Routes>
        </Suspense>

        {/* ИИ-помощник */}
        <AIAgent isOpen={isAIOpen} onToggle={() => setIsAIOpen(!isAIOpen)} />

        {/* Мобильная нижняя навигация */}
        <MobileBottomNav onOpenAI={() => setIsAIOpen(true)} />
      </div>
    </ErrorBoundary>
  );
}

export default App