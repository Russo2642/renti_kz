import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { authAPI } from '../lib/api.js';

const useAuthStore = create(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isLoading: true,
      isAuthenticated: false,
      initialized: false,

      // Авторизация
      login: async (phone, password) => {
        set({ isLoading: true });
        try {
          const response = await authAPI.login(phone, password);
          const { access_token, refresh_token, user } = response.data;
          
          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', refresh_token);
          
          set({
            user,
            accessToken: access_token,
            refreshToken: refresh_token,
            isAuthenticated: true,
            isLoading: false,
            initialized: true,
          });
          
          return { success: true, user };
        } catch (error) {
          set({ isLoading: false });
          return { 
            success: false, 
            error: error.response?.data?.error || 'Ошибка авторизации' 
          };
        }
      },

      // Регистрация
      register: async (userData) => {
        set({ isLoading: true });
        try {
          const response = await authAPI.register(userData);
          set({ isLoading: false });
          return { success: true, data: response.data };
        } catch (error) {
          set({ isLoading: false });
          return { 
            success: false, 
            error: error.response?.data?.error || 'Ошибка регистрации' 
          };
        }
      },

      // Выход
      logout: async () => {
        const { accessToken } = get();
        try {
          if (accessToken) {
            await authAPI.logout(accessToken);
          }
        } catch (error) {
          console.error('Ошибка при выходе:', error);
        } finally {
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            initialized: true,
          });
        }
      },

      // Получение текущего пользователя
      getCurrentUser: async () => {
        const token = localStorage.getItem('access_token');
        if (!token) {
          set({ isLoading: false, initialized: true });
          return false;
        }

        set({ isLoading: true });
        try {
          const response = await authAPI.getCurrentUser();
          set({
            user: response.data.user || response.data,
            accessToken: token,
            refreshToken: localStorage.getItem('refresh_token'),
            isAuthenticated: true,
            isLoading: false,
            initialized: true,
          });
          return true;
        } catch (error) {
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            initialized: true,
          });
          return false;
        }
      },

      // Обновление профиля пользователя
      updateUser: (userData) => {
        set((state) => ({
          user: { ...state.user, ...userData }
        }));
      },

      // Проверка роли пользователя
      hasRole: (roles) => {
        const { user } = get();
        if (!user || !user.role) return false;
        
        if (Array.isArray(roles)) {
          return roles.includes(user.role);
        }
        return user.role === roles;
      },

      // Проверка на администратора
      isAdmin: () => {
        const { user } = get();
        return user?.role === 'admin';
      },

      // Проверка на модератора
      isModerator: () => {
        const { user } = get();
        return user?.role === 'moderator' || user?.role === 'admin';
      },

      // Проверка на владельца
      isOwner: () => {
        const { user } = get();
        return user?.role === 'owner';
      },

      // Проверка на пользователя
      isUser: () => {
        const { user } = get();
        return user?.role === 'user';
      },

      // Проверка на консьержа
      isConcierge: () => {
        const { user } = get();
        return user?.role === 'concierge';
      },

      // Инициализация при загрузке приложения
      initialize: async () => {
        const { initialized } = get();
        if (initialized) return;
        
        const token = localStorage.getItem('access_token');
        if (token) {
          await get().getCurrentUser();
        } else {
          set({ isLoading: false, initialized: true });
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ 
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

export default useAuthStore; 