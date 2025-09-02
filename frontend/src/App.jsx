import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider, theme, App as AntdApp } from 'antd';
import { Toaster } from 'react-hot-toast';
import ruRU from 'antd/locale/ru_RU';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import './App.css';

// Stores
import useAuthStore from './store/useAuthStore.js';

// Components
import ProtectedRoute from './components/ProtectedRoute.jsx';
import Layout from './components/Layout.jsx';
import LoginPage from './pages/auth/LoginPage.jsx';
import LoadingSpinner from './components/LoadingSpinner.jsx';

// Admin Pages
import DashboardPage from './pages/admin/DashboardPage.jsx';
import UsersPage from './pages/admin/UsersPage.jsx';
import ApartmentsPage from './pages/admin/ApartmentsPage.jsx';
import BookingsPage from './pages/admin/BookingsPage.jsx';
import SendNotificationsPage from './pages/admin/SendNotificationsPage.jsx';
import ConciergesPage from './pages/admin/ConciergesPage.jsx';
import CleanersPage from './pages/admin/CleanersPage.jsx';
import LocksPage from './pages/admin/LocksPage.jsx';
import ChatPage from './pages/admin/ChatPage.jsx';

// Owner Pages
import OwnerDashboardPage from './pages/owner/OwnerDashboardPage.jsx';
import OwnerApartmentsPage from './pages/owner/OwnerApartmentsPage.jsx';
import OwnerBookingsPage from './pages/owner/OwnerBookingsPage.jsx';
import OwnerStatisticsPage from './pages/owner/OwnerStatisticsPage.jsx';

// Concierge Pages
import ConciergeDashboardPage from './pages/concierge/ConciergeDashboardPage.jsx';
import ConciergeApartmentsPage from './pages/concierge/ConciergeApartmentsPage.jsx';
import ConciergeBookingsPage from './pages/concierge/ConciergeBookingsPage.jsx';
import ConciergeChatPage from './pages/concierge/ConciergeChatPage.jsx';

// Cleaner Pages
import CleanerDashboardPage from './pages/cleaner/CleanerDashboardPage.jsx';
import CleanerApartmentsPage from './pages/cleaner/CleanerApartmentsPage.jsx';
import CleanerCleaningPage from './pages/cleaner/CleanerCleaningPage.jsx';
import CleanerSchedulePage from './pages/cleaner/CleanerSchedulePage.jsx';

// Universal Notifications Page
import UniversalNotificationsPage from './pages/notifications/UniversalNotificationsPage.jsx';

// Настройка dayjs
dayjs.locale('ru');

// Создание клиента React Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

// Компонент для главной страницы с редиректом
const HomePage = () => {
  const { user } = useAuthStore();
  
  if (user?.role === 'owner') {
    return <Navigate to="/owner/apartments" replace />;
  }
  if (user?.role === 'concierge') {
    return <Navigate to="/concierge/dashboard" replace />;
  }
  if (user?.role === 'cleaner') {
    return <Navigate to="/cleaner/dashboard" replace />;
  }
  // Админы, модераторы идут в админскую панель
  return <Navigate to="/admin/dashboard" replace />;
};

function App() {
  const { isAuthenticated, isLoading, initialize, initialized, user } = useAuthStore();

  useEffect(() => {
    if (!initialized) {
      initialize();
    }
  }, [initialize, initialized]);

  // Показываем загрузку только если не инициализирована
  if (isLoading && !initialized) {
    return <LoadingSpinner />;
  }

  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        locale={ruRU}
        theme={{
          token: {
            colorPrimary: '#3b82f6',
            borderRadius: 8,
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
          },
          algorithm: theme.defaultAlgorithm,
        }}
      >
        <AntdApp>
          <Router>
            <div className="App w-full min-h-screen">
              <Routes>
              {/* Логин */}
              <Route 
                path="/login" 
                element={
                  isAuthenticated ? (
                    <Navigate to={
                      user?.role === 'owner' ? '/owner/apartments' : 
                      user?.role === 'concierge' ? '/concierge/dashboard' :
                      user?.role === 'cleaner' ? '/cleaner/dashboard' :
                      '/admin/dashboard'
                    } replace />
                  ) : (
                    <LoginPage />
                  )
                } 
              />
              <Route 
                path="/admin/login" 
                element={
                  isAuthenticated ? (
                    <Navigate to={
                      user?.role === 'owner' ? '/owner/apartments' : 
                      user?.role === 'concierge' ? '/concierge/dashboard' :
                      user?.role === 'cleaner' ? '/cleaner/dashboard' :
                      '/admin/dashboard'
                    } replace />
                  ) : (
                    <LoginPage />
                  )
                } 
              />

              {/* Главная страница - редирект */}
              <Route 
                path="/" 
                element={<HomePage />} 
              />

              {/* Админские маршруты */}
              <Route 
                path="/admin/dashboard" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <DashboardPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/users" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <UsersPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/apartments" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <ApartmentsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/bookings" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <BookingsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/notifications" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <UniversalNotificationsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/send-notifications" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <SendNotificationsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/concierges" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <ConciergesPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/cleaners" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <CleanersPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/locks" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <LocksPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/admin/chat" 
                element={
                  <ProtectedRoute requiredRoles={['admin', 'moderator']}>
                    <Layout>
                      <ChatPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />

              {/* Маршруты владельца */}
              <Route 
                path="/owner/dashboard" 
                element={
                  <ProtectedRoute requiredRoles={['owner']}>
                    <Layout>
                      <OwnerDashboardPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/owner/apartments" 
                element={
                  <ProtectedRoute requiredRoles={['owner']}>
                    <Layout>
                      <OwnerApartmentsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/owner/bookings" 
                element={
                  <ProtectedRoute requiredRoles={['owner']}>
                    <Layout>
                      <OwnerBookingsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/owner/statistics" 
                element={
                  <ProtectedRoute requiredRoles={['owner']}>
                    <Layout>
                      <OwnerStatisticsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/owner/notifications" 
                element={
                  <ProtectedRoute requiredRoles={['owner']}>
                    <Layout>
                      <UniversalNotificationsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />

              {/* Маршруты консьержа */}
              <Route 
                path="/concierge/dashboard" 
                element={
                  <ProtectedRoute requiredRoles={['concierge']}>
                    <Layout>
                      <ConciergeDashboardPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/concierge/apartments" 
                element={
                  <ProtectedRoute requiredRoles={['concierge']}>
                    <Layout>
                      <ConciergeApartmentsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/concierge/bookings" 
                element={
                  <ProtectedRoute requiredRoles={['concierge']}>
                    <Layout>
                      <ConciergeBookingsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/concierge/chat" 
                element={
                  <ProtectedRoute requiredRoles={['concierge']}>
                    <Layout>
                      <ConciergeChatPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/concierge/notifications" 
                element={
                  <ProtectedRoute requiredRoles={['concierge']}>
                    <Layout>
                      <UniversalNotificationsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />

              {/* Маршруты уборщицы */}
              <Route 
                path="/cleaner/dashboard" 
                element={
                  <ProtectedRoute requiredRoles={['cleaner']}>
                    <Layout>
                      <CleanerDashboardPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/cleaner/apartments" 
                element={
                  <ProtectedRoute requiredRoles={['cleaner']}>
                    <Layout>
                      <CleanerApartmentsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/cleaner/apartments-for-cleaning" 
                element={
                  <ProtectedRoute requiredRoles={['cleaner']}>
                    <Layout>
                      <CleanerCleaningPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/cleaner/schedule" 
                element={
                  <ProtectedRoute requiredRoles={['cleaner']}>
                    <Layout>
                      <CleanerSchedulePage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/cleaner/notifications" 
                element={
                  <ProtectedRoute requiredRoles={['cleaner']}>
                    <Layout>
                      <UniversalNotificationsPage />
                    </Layout>
                  </ProtectedRoute>
                } 
              />

              {/* Fallback - редирект на главную */}
              <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </div>
          </Router>
        </AntdApp>
        
        {/* Toast уведомления */}
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: '#363636',
              color: '#fff',
            },
            success: {
              duration: 3000,
              iconTheme: {
                primary: '#22c55e',
                secondary: '#fff',
              },
            },
            error: {
              duration: 5000,
              iconTheme: {
                primary: '#ef4444',
                secondary: '#fff',
              },
            },
          }}
        />
      </ConfigProvider>
    </QueryClientProvider>
  );
}

export default App;
