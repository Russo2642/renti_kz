import React from 'react';
import { Navigate } from 'react-router-dom';
import useAuthStore from '../store/useAuthStore.js';
import LoadingSpinner from './LoadingSpinner';

const ProtectedRoute = ({ children, requiredRoles = [] }) => {
  const { isAuthenticated, user, isLoading, initialized } = useAuthStore();

  // Показываем загрузку только если не инициализирована и загружается
  if (isLoading && !initialized) {
    return <LoadingSpinner />;
  }

  // Если не авторизован, редиректим на логин
  if (!isAuthenticated) {
    return <Navigate to="/admin/login" replace />;
  }

  // Если требуются определенные роли
  if (requiredRoles.length > 0) {
    if (!user || !requiredRoles.includes(user.role)) {
      // Редиректим на соответствующую роли страницу
      const redirectPath = 
        user?.role === 'owner' ? '/owner/apartments' : 
        user?.role === 'concierge' ? '/concierge/dashboard' : 
        user?.role === 'cleaner' ? '/cleaner/dashboard' :
        '/admin/dashboard';
      return <Navigate to={redirectPath} replace />;
    }
  }

  return children;
};

export default ProtectedRoute; 