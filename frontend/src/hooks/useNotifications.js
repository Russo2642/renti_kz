import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { notificationsAPI } from '../lib/api.js';
import useAuthStore from '../store/useAuthStore.js';

/**
 * Хук для работы с уведомлениями
 * Предоставляет функции для получения, отметки и удаления уведомлений
 */
const useNotifications = (options = {}) => {
  const {
    autoRefresh = true,
    refreshInterval = 30000, // 30 секунд
    enableRealtime = false,
    onNewNotification = null,
    onNotificationRead = null,
    onNotificationDeleted = null
  } = options;

  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const [isConnected, setIsConnected] = useState(false);

  // Получение непрочитанных уведомлений
  const {
    data: unreadNotifications,
    isLoading: unreadLoading,
    error: unreadError,
    refetch: refetchUnread
  } = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: () => notificationsAPI.getUnread(),
    enabled: !!user,
    refetchInterval: autoRefresh ? refreshInterval : false,
    onError: (error) => {
      console.error('Ошибка загрузки непрочитанных уведомлений:', error);
    }
  });

  // Получение количества непрочитанных
  const {
    data: unreadCount,
    isLoading: countLoading,
    error: countError,
    refetch: refetchCount
  } = useQuery({
    queryKey: ['notifications-count'],
    queryFn: () => notificationsAPI.getCount(),
    enabled: !!user,
    refetchInterval: autoRefresh ? refreshInterval / 2 : false, // Обновляем чаще
    onError: (error) => {
      console.error('Ошибка загрузки количества уведомлений:', error);
    }
  });

  // Мутация для отметки как прочитанное
  const markAsReadMutation = useMutation({
    mutationFn: (id) => notificationsAPI.markAsRead(id),
    onSuccess: (data, id) => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      
      if (onNotificationRead) {
        onNotificationRead(id);
      }
    },
    onError: (error) => {
      message.error('Ошибка при отметке уведомления как прочитанное');
      console.error('Ошибка отметки уведомления:', error);
    }
  });

  // Мутация для отметки нескольких как прочитанные
  const markMultipleAsReadMutation = useMutation({
    mutationFn: (ids) => notificationsAPI.markMultipleAsRead(ids),
    onSuccess: (data, ids) => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      
      if (onNotificationRead) {
        ids.forEach(id => onNotificationRead(id));
      }
    },
    onError: (error) => {
      message.error('Ошибка при отметке уведомлений как прочитанные');
      console.error('Ошибка отметки уведомлений:', error);
    }
  });

  // Мутация для отметки всех как прочитанные
  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationsAPI.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      
      if (onNotificationRead) {
        onNotificationRead('all');
      }
    },
    onError: (error) => {
      message.error('Ошибка при отметке всех уведомлений как прочитанные');
      console.error('Ошибка отметки всех уведомлений:', error);
    }
  });

  // Мутация для удаления уведомления
  const deleteNotificationMutation = useMutation({
    mutationFn: (id) => notificationsAPI.delete(id),
    onSuccess: (data, id) => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      
      if (onNotificationDeleted) {
        onNotificationDeleted(id);
      }
    },
    onError: (error) => {
      message.error('Ошибка при удалении уведомления');
      console.error('Ошибка удаления уведомления:', error);
    }
  });

  // Мутация для удаления прочитанных уведомлений
  const deleteReadMutation = useMutation({
    mutationFn: () => notificationsAPI.deleteRead(),
    onSuccess: (data) => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      
      message.success(`Удалено ${data.data.deleted_count} прочитанных уведомлений`);
      
      if (onNotificationDeleted) {
        onNotificationDeleted('read');
      }
    },
    onError: (error) => {
      message.error('Ошибка при удалении прочитанных уведомлений');
      console.error('Ошибка удаления прочитанных уведомлений:', error);
    }
  });

  // Мутация для удаления всех уведомлений
  const deleteAllMutation = useMutation({
    mutationFn: () => notificationsAPI.deleteAll(),
    onSuccess: (data) => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      
      message.success(`Удалено ${data.data.deleted_count} уведомлений`);
      
      if (onNotificationDeleted) {
        onNotificationDeleted('all');
      }
    },
    onError: (error) => {
      message.error('Ошибка при удалении всех уведомлений');
      console.error('Ошибка удаления всех уведомлений:', error);
    }
  });

  // WebSocket подключение для real-time уведомлений
  useEffect(() => {
    if (!enableRealtime || !user) return;

    let ws = null;
    
    const connectWebSocket = () => {
      try {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const token = localStorage.getItem('access_token');
        
        if (!token) return;
        
        ws = new WebSocket(`${protocol}//${host}/api/ws/notifications?token=${encodeURIComponent(token)}`);
        
        ws.onopen = () => {
          setIsConnected(true);
          console.log('WebSocket подключен для уведомлений');
        };
        
        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            
            switch (data.type) {
              case 'new_notification':
                // Обновляем кэш уведомлений
                queryClient.invalidateQueries(['notifications-unread']);
                queryClient.invalidateQueries(['notifications-count']);
                
                if (onNewNotification) {
                  onNewNotification(data.notification);
                }
                break;
                
              case 'notification_read':
                queryClient.invalidateQueries(['notifications-unread']);
                queryClient.invalidateQueries(['notifications-count']);
                break;
                
              case 'notification_deleted':
                queryClient.invalidateQueries(['notifications']);
                queryClient.invalidateQueries(['notifications-unread']);
                queryClient.invalidateQueries(['notifications-count']);
                break;
                
              default:
                console.log('Неизвестный тип WebSocket сообщения:', data.type);
            }
          } catch (error) {
            console.error('Ошибка парсинга WebSocket сообщения:', error);
          }
        };
        
        ws.onclose = (event) => {
          setIsConnected(false);
          
          // Автоматическое переподключение через 3 секунды
          if (event.code !== 1000) {
            setTimeout(connectWebSocket, 3000);
          }
        };
        
        ws.onerror = (error) => {
          console.error('WebSocket ошибка:', error);
          setIsConnected(false);
        };
        
      } catch (error) {
        console.error('Ошибка создания WebSocket подключения:', error);
      }
    };
    
    connectWebSocket();
    
    return () => {
      if (ws) {
        ws.close(1000, 'Компонент размонтирован');
      }
    };
  }, [enableRealtime, user, queryClient, onNewNotification]);

  // Функция для регистрации устройства для push-уведомлений
  const registerDevice = useCallback(async (deviceToken, deviceType = 'web') => {
    try {
      await notificationsAPI.registerDevice({
        device_token: deviceToken,
        device_type: deviceType,
        app_version: '1.0.0',
        os_version: navigator.userAgent
      });
      message.success('Устройство зарегистрировано для получения уведомлений');
    } catch (error) {
      console.error('Ошибка регистрации устройства:', error);
      message.error('Ошибка регистрации устройства для уведомлений');
    }
  }, []);

  // Функция для обновления heartbeat устройства
  const updateDeviceHeartbeat = useCallback(async (deviceToken) => {
    try {
      await notificationsAPI.updateDeviceHeartbeat(deviceToken);
    } catch (error) {
      console.error('Ошибка обновления heartbeat:', error);
    }
  }, []);

  // Функция для удаления устройства
  const removeDevice = useCallback(async (deviceToken) => {
    try {
      await notificationsAPI.removeDevice(deviceToken);
      message.success('Устройство удалено из списка для уведомлений');
    } catch (error) {
      console.error('Ошибка удаления устройства:', error);
      message.error('Ошибка удаления устройства');
    }
  }, []);

  // Функция для ручного обновления
  const refresh = useCallback(() => {
    refetchUnread();
    refetchCount();
  }, [refetchUnread, refetchCount]);

  return {
    // Данные
    notifications: unreadNotifications?.data || [],
    count: unreadCount?.data?.count || 0,
    
    // Состояния загрузки
    isLoading: unreadLoading || countLoading,
    unreadLoading,
    countLoading,
    
    // Ошибки
    error: unreadError || countError,
    unreadError,
    countError,
    
    // WebSocket состояние
    isConnected,
    
    // Функции для отметки как прочитанное
    markAsRead: markAsReadMutation.mutate,
    markMultipleAsRead: markMultipleAsReadMutation.mutate,
    markAllAsRead: markAllAsReadMutation.mutate,
    
    // Функции для удаления
    deleteNotification: deleteNotificationMutation.mutate,
    deleteRead: deleteReadMutation.mutate,
    deleteAll: deleteAllMutation.mutate,
    
    // Состояния мутаций
    isMarkingAsRead: markAsReadMutation.isLoading,
    isMarkingMultipleAsRead: markMultipleAsReadMutation.isLoading,
    isMarkingAllAsRead: markAllAsReadMutation.isLoading,
    isDeleting: deleteNotificationMutation.isLoading,
    isDeletingRead: deleteReadMutation.isLoading,
    isDeletingAll: deleteAllMutation.isLoading,
    
    // Функции для устройств
    registerDevice,
    updateDeviceHeartbeat,
    removeDevice,
    
    // Утилиты
    refresh
  };
};

export default useNotifications;
