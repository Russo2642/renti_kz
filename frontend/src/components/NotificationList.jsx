import React, { useState } from 'react';
import { 
  List, 
  Card, 
  Button, 
  Tag, 
  Space, 
  Typography, 
  Pagination,
  Empty,
  Spin,
  Checkbox,
  Divider,
  message,
  Popconfirm
} from 'antd';
import {
  CheckOutlined,
  DeleteOutlined,
  BellOutlined,
  EyeOutlined,
  ClockCircleOutlined
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notificationsAPI } from '../lib/api.js';
import { getNotificationTypeColor, getNotificationTypeText, getPriorityColor } from '../utils/notificationTypes.js';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/ru';

dayjs.extend(relativeTime);
dayjs.locale('ru');

const { Text, Title } = Typography;

const NotificationList = ({ 
  title = "Уведомления",
  showPagination = true,
  pageSize = 10,
  showBulkActions = true,
  showStats = true,
  cardProps = {},
  listProps = {}
}) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedIds, setSelectedIds] = useState([]);
  const queryClient = useQueryClient();

  // Получение уведомлений с пагинацией
  const { data: notificationsData, isLoading } = useQuery({
    queryKey: ['notifications', { page: currentPage, page_size: pageSize }],
    queryFn: () => notificationsAPI.getAll({ 
      page: currentPage, 
      page_size: pageSize 
    }),
    keepPreviousData: true,
  });

  // Получение количества непрочитанных
  const { data: unreadCount } = useQuery({
    queryKey: ['notifications-count'],
    queryFn: () => notificationsAPI.getCount(),
    refetchInterval: 30000,
  });

  // Мутация для отметки как прочитанное
  const markAsReadMutation = useMutation({
    mutationFn: (id) => notificationsAPI.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-count']);
      message.success('Уведомление отмечено как прочитанное');
    }
  });

  // Мутация для отметки нескольких как прочитанные
  const markMultipleAsReadMutation = useMutation({
    mutationFn: (ids) => notificationsAPI.markMultipleAsRead(ids),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-count']);
      setSelectedIds([]);
      message.success('Выбранные уведомления отмечены как прочитанные');
    }
  });

  // Мутация для отметки всех как прочитанные
  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationsAPI.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-count']);
      setSelectedIds([]);
      message.success('Все уведомления отмечены как прочитанные');
    }
  });

  // Мутация для удаления уведомления
  const deleteNotificationMutation = useMutation({
    mutationFn: (id) => notificationsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-count']);
      message.success('Уведомление удалено');
    }
  });

  // Мутация для удаления прочитанных уведомлений
  const deleteReadMutation = useMutation({
    mutationFn: () => notificationsAPI.deleteRead(),
    onSuccess: (data) => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-count']);
      message.success(`Удалено ${data.data.deleted_count} прочитанных уведомлений`);
    }
  });

  // Функции для работы с типами уведомлений импортированы из utils

  const notifications = notificationsData?.data?.notifications || [];
  const pagination = notificationsData?.data?.pagination || {};
  const count = unreadCount?.data?.count || 0;

  // Обработчики
  const handleSelectAll = (checked) => {
    if (checked) {
      setSelectedIds(notifications.map(n => n.id));
    } else {
      setSelectedIds([]);
    }
  };

  const handleSelectNotification = (id, checked) => {
    if (checked) {
      setSelectedIds(prev => [...prev, id]);
    } else {
      setSelectedIds(prev => prev.filter(selectedId => selectedId !== id));
    }
  };

  const handleMarkSelectedAsRead = () => {
    if (selectedIds.length > 0) {
      markMultipleAsReadMutation.mutate(selectedIds);
    }
  };

  return (
    <Card 
      title={
        <div className="flex items-center justify-between">
          <Title level={4} className="!mb-0">
            {title} {count > 0 && `(${count} непрочитанных)`}
          </Title>
          <Space>
            {showBulkActions && count > 0 && (
              <>
                <Button
                  type="primary"
                  icon={<CheckOutlined />}
                  onClick={() => markAllAsReadMutation.mutate()}
                  loading={markAllAsReadMutation.isLoading}
                  size="small"
                >
                  Отметить все как прочитанные
                </Button>
                <Popconfirm
                  title="Удалить все прочитанные уведомления?"
                  onConfirm={() => deleteReadMutation.mutate()}
                  okText="Да"
                  cancelText="Нет"
                >
                  <Button
                    danger
                    icon={<DeleteOutlined />}
                    loading={deleteReadMutation.isLoading}
                    size="small"
                  >
                    Удалить прочитанные
                  </Button>
                </Popconfirm>
              </>
            )}
          </Space>
        </div>
      }
      {...cardProps}
    >
      {/* Статистика */}
      {showStats && count > 0 && (
        <div className="mb-3 md:mb-4 p-2 md:p-3 bg-blue-50 rounded-lg">
          <div className="flex flex-col sm:flex-row sm:items-center space-y-1 sm:space-y-0 sm:space-x-4">
            <Text className="text-sm md:text-base">
              <BellOutlined className="text-blue-500 mr-1" />
              Всего: {pagination.total || 0}
            </Text>
            <Text className="text-sm md:text-base">
              <ClockCircleOutlined className="text-orange-500 mr-1" />
              Непрочитанных: {count}
            </Text>
          </div>
        </div>
      )}

      {/* Массовые действия */}
      {showBulkActions && notifications.length > 0 && (
        <div className="mb-3 md:mb-4">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-2 sm:space-y-0">
            <Space size="small">
              <Checkbox
                checked={selectedIds.length === notifications.length}
                indeterminate={selectedIds.length > 0 && selectedIds.length < notifications.length}
                onChange={(e) => handleSelectAll(e.target.checked)}
                className="text-sm md:text-base"
              >
                <span className="hidden sm:inline">Выбрать все</span>
                <span className="sm:hidden">Все</span>
              </Checkbox>
              {selectedIds.length > 0 && (
                <Text type="secondary" className="text-xs md:text-sm">
                  Выбрано: {selectedIds.length}
                </Text>
              )}
            </Space>
            
            {selectedIds.length > 0 && (
              <Space size="small">
                <Button
                  size="small"
                  icon={<CheckOutlined />}
                  onClick={handleMarkSelectedAsRead}
                  loading={markMultipleAsReadMutation.isLoading}
                  className="text-xs md:text-sm"
                >
                  <span className="hidden sm:inline">Отметить как прочитанные ({selectedIds.length})</span>
                  <span className="sm:hidden">Прочитать ({selectedIds.length})</span>
                </Button>
              </Space>
            )}
          </div>
        </div>
      )}

      {/* Список уведомлений */}
      <Spin spinning={isLoading}>
        {notifications.length === 0 ? (
          <Empty 
            description="Нет уведомлений"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        ) : (
          <List
            dataSource={notifications}
            renderItem={(notification) => (
              <List.Item
                className={`border-l-4 px-2 md:px-4 py-2 md:py-3 ${!notification.is_read ? 'bg-blue-50' : ''}`}
                style={{ borderLeftColor: getPriorityColor(notification.priority) }}
                actions={[
                  !notification.is_read && (
                    <Button
                      type="text"
                      size="small"
                      icon={<CheckOutlined />}
                      onClick={() => markAsReadMutation.mutate(notification.id)}
                      loading={markAsReadMutation.isLoading}
                      title="Отметить как прочитанное"
                      className="hidden sm:flex"
                    />
                  ),
                  <Popconfirm
                    title="Удалить это уведомление?"
                    onConfirm={() => deleteNotificationMutation.mutate(notification.id)}
                    okText="Да"
                    cancelText="Нет"
                  >
                    <Button
                      type="text"
                      size="small"
                      icon={<DeleteOutlined />}
                      loading={deleteNotificationMutation.isLoading}
                      danger
                      title="Удалить"
                      className="hidden sm:flex"
                    />
                  </Popconfirm>
                ].filter(Boolean)}
                extra={
                  <div className="flex items-center space-x-2">
                    {showBulkActions && (
                      <Checkbox
                        checked={selectedIds.includes(notification.id)}
                        onChange={(e) => handleSelectNotification(notification.id, e.target.checked)}
                        className="sm:hidden"
                      />
                    )}
                    {/* Мобильные действия */}
                    <div className="flex sm:hidden space-x-1">
                      {!notification.is_read && (
                        <Button
                          type="text"
                          size="small"
                          icon={<CheckOutlined />}
                          onClick={() => markAsReadMutation.mutate(notification.id)}
                          loading={markAsReadMutation.isLoading}
                          title="Отметить как прочитанное"
                        />
                      )}
                      <Popconfirm
                        title="Удалить?"
                        onConfirm={() => deleteNotificationMutation.mutate(notification.id)}
                        okText="Да"
                        cancelText="Нет"
                      >
                        <Button
                          type="text"
                          size="small"
                          icon={<DeleteOutlined />}
                          loading={deleteNotificationMutation.isLoading}
                          danger
                          title="Удалить"
                        />
                      </Popconfirm>
                    </div>
                    {showBulkActions && (
                      <Checkbox
                        checked={selectedIds.includes(notification.id)}
                        onChange={(e) => handleSelectNotification(notification.id, e.target.checked)}
                        className="hidden sm:block"
                      />
                    )}
                  </div>
                }
              >
                <List.Item.Meta
                  avatar={
                    <div className={`w-3 h-3 rounded-full ${!notification.is_read ? 'bg-blue-500' : 'bg-gray-300'}`} />
                  }
                  title={
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-1 sm:space-y-0">
                      <Text 
                        strong={!notification.is_read}
                        className="text-sm md:text-base pr-2"
                        style={{ 
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden'
                        }}
                      >
                        {notification.title}
                      </Text>
                      <div className="flex items-center space-x-1 sm:space-x-2 flex-shrink-0">
                        <Tag 
                          color={getNotificationTypeColor(notification.type)}
                          className="text-xs"
                        >
                          {getNotificationTypeText(notification.type)}
                        </Tag>
                        <Text className="text-xs text-gray-400 whitespace-nowrap">
                          {dayjs(notification.created_at).fromNow()}
                        </Text>
                      </div>
                    </div>
                  }
                  description={
                    <div className="space-y-1">
                      <Text 
                        className="text-gray-600 text-sm md:text-base"
                        style={{ 
                          display: '-webkit-box',
                          WebkitLineClamp: 3,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden'
                        }}
                      >
                        {notification.body}
                      </Text>
                      {notification.data && Object.keys(notification.data).length > 0 && (
                        <div className="text-xs text-gray-400 break-all">
                          <span className="hidden sm:inline">Дополнительные данные: </span>
                          <span className="sm:hidden">Данные: </span>
                          {JSON.stringify(notification.data)}
                        </div>
                      )}
                      {(notification.booking_id || notification.apartment_id) && (
                        <div className="text-xs text-gray-400">
                          {notification.booking_id && `Бронь #${notification.booking_id}`}
                          {notification.booking_id && notification.apartment_id && ' • '}
                          {notification.apartment_id && `Квартира #${notification.apartment_id}`}
                        </div>
                      )}
                    </div>
                  }
                />
              </List.Item>
            )}
            {...listProps}
          />
        )}
      </Spin>

      {/* Пагинация */}
      {showPagination && pagination.total > pageSize && (
        <div className="mt-3 md:mt-4 text-center">
          <Pagination
            current={currentPage}
            total={pagination.total}
            pageSize={pageSize}
            onChange={setCurrentPage}
            showSizeChanger={false}
            showQuickJumper={false}
            responsive={true}
            size="small"
            showTotal={(total, range) => (
              <span className="text-xs md:text-sm">
                <span className="hidden sm:inline">{range[0]}-{range[1]} из {total} уведомлений</span>
                <span className="sm:hidden">{range[0]}-{range[1]} / {total}</span>
              </span>
            )}
            className="pagination-mobile"
          />
        </div>
      )}
    </Card>
  );
};

export default NotificationList;
