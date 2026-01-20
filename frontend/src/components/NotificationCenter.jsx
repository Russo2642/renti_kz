import React, { useState, useEffect } from 'react';
import { 
  Dropdown, 
  Badge, 
  Button, 
  List, 
  Typography, 
  Card, 
  Space, 
  Spin, 
  Empty,
  Divider,
  Tag,
  message
} from 'antd';
import {
  BellOutlined,
  CheckOutlined,
  DeleteOutlined,
  SettingOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { notificationsAPI } from '../lib/api.js';
import { getNotificationTypeColor, getNotificationTypeText, getPriorityColor } from '../utils/notificationTypes.js';
import useAuthStore from '../store/useAuthStore.js';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/ru';

dayjs.extend(relativeTime);
dayjs.locale('ru');

const { Text, Title } = Typography;

const NotificationCenter = ({ maxItems = 5, showAllLink = true, placement = "bottomRight" }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIds, setSelectedIds] = useState([]);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { user } = useAuthStore();

  // Отслеживание размера экрана
  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth < 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получение непрочитанных уведомлений
  const { data: unreadNotifications, isLoading: unreadLoading } = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: () => notificationsAPI.getUnread(),
    refetchInterval: 30000, // Обновляем каждые 30 секунд
    enabled: isOpen, // Загружаем только когда открыто
  });

  // Получение количества непрочитанных
  const { data: unreadCount, isLoading: countLoading } = useQuery({
    queryKey: ['notifications-count'],
    queryFn: () => notificationsAPI.getCount(),
    refetchInterval: 15000, // Обновляем каждые 15 секунд
  });

  // Мутация для отметки как прочитанное
  const markAsReadMutation = useMutation({
    mutationFn: (id) => notificationsAPI.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      message.success('Уведомление отмечено как прочитанное');
    },
    onError: () => {
      message.error('Ошибка при отметке уведомления');
    }
  });

  // Мутация для отметки нескольких как прочитанные
  const markMultipleAsReadMutation = useMutation({
    mutationFn: (ids) => notificationsAPI.markMultipleAsRead(ids),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      setSelectedIds([]);
      message.success('Уведомления отмечены как прочитанные');
    },
    onError: () => {
      message.error('Ошибка при отметке уведомлений');
    }
  });

  // Мутация для отметки всех как прочитанные
  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationsAPI.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      message.success('Все уведомления отмечены как прочитанные');
    },
    onError: () => {
      message.error('Ошибка при отметке всех уведомлений');
    }
  });

  // Мутация для удаления уведомления
  const deleteNotificationMutation = useMutation({
    mutationFn: (id) => notificationsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications-unread']);
      queryClient.invalidateQueries(['notifications-count']);
      message.success('Уведомление удалено');
    },
    onError: () => {
      message.error('Ошибка при удалении уведомления');
    }
  });

  // Функции для работы с типами уведомлений импортированы из utils

  const notifications = unreadNotifications?.data || [];
  const count = unreadCount?.data?.count || 0;
  const displayNotifications = maxItems ? notifications.slice(0, maxItems) : notifications;

  // Определяем правильный путь для страницы уведомлений в зависимости от роли
  const getNotificationsPath = () => {
    if (!user?.role) return '/admin/notifications';
    
    switch (user.role) {
      case 'admin':
        return '/admin/notifications';
      case 'owner':
        return '/owner/notifications';
      case 'concierge':
        return '/concierge/notifications';
      case 'cleaner':
        return '/cleaner/notifications';
      default:
        return '/admin/notifications';
    }
  };

  // Содержимое дропдауна
  const dropdownContent = (
    <Card 
      className={`${isMobile ? 'w-80 max-w-80' : 'w-96 max-w-96'} shadow-lg`}
      bodyStyle={{ padding: 0 }}
      title={
        <div className={`flex items-center justify-between ${isMobile ? 'px-3 py-2' : 'px-4 py-2'}`}>
          <Title level={isMobile ? 6 : 5} className="!mb-0">
            {isMobile ? `Уведомления ${count > 0 ? `(${count})` : ''}` : `Уведомления ${count > 0 ? `(${count})` : ''}`}
          </Title>
          <Space size="small">
            <Button
              type="text"
              size={isMobile ? "small" : "small"}
              icon={<ReloadOutlined />}
              loading={unreadLoading}
              onClick={() => {
                queryClient.invalidateQueries(['notifications-unread']);
                queryClient.invalidateQueries(['notifications-count']);
              }}
              className={isMobile ? "!p-1" : ""}
            />
            {count > 0 && (
              <Button
                type="text"
                size={isMobile ? "small" : "small"}
                icon={<CheckOutlined />}
                onClick={() => markAllAsReadMutation.mutate()}
                loading={markAllAsReadMutation.isLoading}
                title="Отметить все как прочитанные"
                className={isMobile ? "!p-1" : ""}
              />
            )}
          </Space>
        </div>
      }
    >
      <div className={`${isMobile ? 'max-h-64' : 'max-h-80'} overflow-y-auto`}>
        {unreadLoading ? (
          <div className="flex justify-center py-8">
            <Spin />
          </div>
        ) : displayNotifications.length === 0 ? (
          <div className="p-4">
            <Empty 
              description="Нет новых уведомлений"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          </div>
        ) : (
                      <List
              dataSource={displayNotifications}
              renderItem={(notification) => (
                <List.Item
                  className={`${isMobile ? '!px-3 !py-2' : '!px-4 !py-3'} hover:bg-gray-50 cursor-pointer border-l-4`}
                  style={{ borderLeftColor: getPriorityColor(notification.priority) }}
                  actions={[
                    <Button
                      type="text"
                      size={isMobile ? "small" : "small"}
                      icon={<CheckOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        markAsReadMutation.mutate(notification.id);
                      }}
                      loading={markAsReadMutation.isLoading}
                      title="Отметить как прочитанное"
                      className={isMobile ? "!p-1" : ""}
                    />,
                    <Button
                      type="text"
                      size={isMobile ? "small" : "small"}
                      icon={<DeleteOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        deleteNotificationMutation.mutate(notification.id);
                      }}
                      loading={deleteNotificationMutation.isLoading}
                      danger
                      title="Удалить"
                      className={isMobile ? "!p-1" : ""}
                    />
                  ]}
                >
                  <List.Item.Meta
                    title={
                      <div className="flex items-start justify-between gap-2">
                        <Text strong className={`${isMobile ? 'text-xs' : 'text-sm'} flex-1 leading-tight`}>
                          {isMobile && notification.title.length > 30 
                            ? `${notification.title.substring(0, 30)}...`
                            : notification.title
                          }
                        </Text>
                        <Tag 
                          color={getNotificationTypeColor(notification.type)}
                          className={`${isMobile ? 'text-xs scale-90' : 'text-xs'} shrink-0`}
                        >
                          {isMobile 
                            ? getNotificationTypeText(notification.type).substring(0, 8)
                            : getNotificationTypeText(notification.type)
                          }
                        </Tag>
                      </div>
                    }
                    description={
                      <div className="space-y-1 mt-1">
                        <Text className={`${isMobile ? 'text-xs' : 'text-sm'} text-gray-600 leading-tight block`}>
                          {notification.body && notification.body.length > (isMobile ? 60 : 100)
                            ? `${notification.body.substring(0, isMobile ? 60 : 100)}...` 
                            : notification.body
                          }
                        </Text>
                        <div className={`${isMobile ? 'flex-col items-start gap-1' : 'flex items-center justify-between'}`}>
                          <Text className="text-xs text-gray-400">
                            {dayjs(notification.created_at).fromNow()}
                          </Text>
                          {(notification.booking_id || notification.apartment_id) && (
                            <Text className="text-xs text-blue-500">
                              {notification.booking_id && `#${notification.booking_id}`}
                              {notification.apartment_id && ` Кв.${notification.apartment_id}`}
                            </Text>
                          )}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
        )}
      </div>
      
      {showAllLink && count > maxItems && (
        <>
          <Divider className="!my-0" />
          <div className={`${isMobile ? 'p-3' : 'p-4'} text-center bg-gray-50`}>
            <Button 
              type="link" 
              onClick={() => {
                setIsOpen(false);
                // Навигация к полной странице уведомлений
                navigate(getNotificationsPath());
              }}
              className={`text-blue-600 font-medium ${isMobile ? 'text-sm' : ''}`}
              size={isMobile ? "small" : "default"}
            >
              {isMobile ? `Все (${count})` : `Посмотреть все уведомления (${count})`}
            </Button>
          </div>
        </>
      )}
    </Card>
  );

  return (
    <Dropdown
      dropdownRender={() => dropdownContent}
      trigger={['click']}
      placement={isMobile ? "bottomRight" : placement}
      open={isOpen}
      onOpenChange={setIsOpen}
    >
      <Badge count={count} size="small" className="cursor-pointer">
        <Button
          type="text"
          icon={<BellOutlined />}
          className="!w-10 !h-10 flex items-center justify-center hover:bg-gray-100"
          loading={countLoading}
        />
      </Badge>
    </Dropdown>
  );
};

export default NotificationCenter;
