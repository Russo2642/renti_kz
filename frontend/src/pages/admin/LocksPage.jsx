import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Table, Button, Card, Tag, Modal, Form, Input, Select, Space,
  Row, Col, Statistic, Typography, Badge, Tooltip, message,
  Popconfirm, Drawer, Descriptions, Timeline, Switch, InputNumber,
  DatePicker
} from 'antd';
import {
  LockOutlined, UnlockOutlined, EditOutlined, DeleteOutlined, EyeOutlined,
  PlusOutlined, WifiOutlined, PoweroffOutlined, HistoryOutlined,
  KeyOutlined, WarningOutlined, CheckCircleOutlined, ThunderboltOutlined, StopOutlined
} from '@ant-design/icons';
import { locksAPI } from '../../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;

const LocksPage = () => {
  const [filters, setFilters] = useState({
    status: null,
    is_online: null,
    apartment_id: null,
    unbound: null,
    min_battery_level: null,
    max_battery_level: null
  });
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [historyVisible, setHistoryVisible] = useState(false);
  const [passwordsVisible, setPasswordsVisible] = useState(false);
  const [generatePasswordVisible, setGeneratePasswordVisible] = useState(false);
  const [selectedLock, setSelectedLock] = useState(null);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const queryClient = useQueryClient();

  // Отслеживание изменения размера экрана
  React.useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получение замков (админская версия)
  const { data: locksData, isLoading } = useQuery({
    queryKey: ['admin-locks', filters, currentPage, pageSize],
    queryFn: () => {
      const params = {
        page: currentPage,
        page_size: pageSize,
      };
      
      // Добавляем фильтры только если они не null
      if (filters.status) params.status = filters.status;
      if (filters.is_online !== null) params.is_online = filters.is_online;
      if (filters.apartment_id) params.apartment_id = filters.apartment_id;
      if (filters.unbound !== null) params.unbound = filters.unbound;
      if (filters.min_battery_level !== null) params.min_battery_level = filters.min_battery_level;
      if (filters.max_battery_level !== null) params.max_battery_level = filters.max_battery_level;
      
      return locksAPI.adminGetAllLocks(params);
    }
  });

  // Получение статистики замков
  const { data: locksStats, isLoading: isLoadingStats } = useQuery({
    queryKey: ['admin-locks-stats'],
    queryFn: locksAPI.adminGetLocksStatistics,
    staleTime: 5 * 60 * 1000, // 5 минут
  });

  // Получение истории для выбранного замка
  const { data: historyData, isLoading: historyLoading } = useQuery({
    queryKey: ['lock-history', selectedLock?.unique_id],
    queryFn: () => selectedLock ? locksAPI.getHistory(selectedLock.unique_id) : null,
    enabled: !!selectedLock && historyVisible
  });

  // Получение паролей для выбранного замка
  const { data: passwordsData, isLoading: passwordsLoading } = useQuery({
    queryKey: ['lock-passwords', selectedLock?.unique_id],
    queryFn: () => selectedLock ? locksAPI.adminGetAllLockPasswords(selectedLock.unique_id) : null,
    enabled: !!selectedLock && passwordsVisible
  });



  // Мутация для создания замка
  const createMutation = useMutation({
    mutationFn: locksAPI.create,
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-locks']);
      queryClient.invalidateQueries(['admin-locks-stats']);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('Замок добавлен');
    }
  });

  // Мутация для обновления замка
  const updateMutation = useMutation({
    mutationFn: ({ id, ...data }) => locksAPI.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-locks']);
      queryClient.invalidateQueries(['admin-locks-stats']);
      setEditModalVisible(false);
      editForm.resetFields();
      message.success('Замок обновлен');
    }
  });

  // Мутация для удаления замка
  const deleteMutation = useMutation({
    mutationFn: locksAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-locks']);
      queryClient.invalidateQueries(['admin-locks-stats']);
      message.success('Замок удален');
    }
  });

  // Админские мутации
  const bindToApartmentMutation = useMutation({
    mutationFn: ({ lockId, apartmentId }) => locksAPI.adminBindLockToApartment(lockId, apartmentId),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-locks']);
      queryClient.invalidateQueries(['admin-locks-stats']);
      message.success('Замок привязан к квартире');
    },
    onError: () => {
      message.error('Ошибка при привязке замка');
    }
  });

  const unbindFromApartmentMutation = useMutation({
    mutationFn: locksAPI.adminUnbindLockFromApartment,
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-locks']);
      queryClient.invalidateQueries(['admin-locks-stats']);
      message.success('Замок отвязан от квартиры');
    },
    onError: () => {
      message.error('Ошибка при отвязке замка');
    }
  });

  const emergencyResetMutation = useMutation({
    mutationFn: locksAPI.adminEmergencyResetLock,
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-locks']);
      message.success('Замок сброшен');
    },
    onError: () => {
      message.error('Ошибка при сбросе замка');
    }
  });



  const syncWithTuyaMutation = useMutation({
    mutationFn: locksAPI.syncWithTuya,
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-locks']);
      message.success('Синхронизация выполнена');
    },
    onError: () => {
      message.error('Ошибка синхронизации');
    }
  });

  // Мутации для управления паролями
  const generatePasswordMutation = useMutation({
    mutationFn: ({ uniqueId, data }) => locksAPI.adminGeneratePassword(uniqueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['lock-passwords', selectedLock?.unique_id]);
      queryClient.invalidateQueries(['admin-locks']);
      setPasswordsVisible(false);
      passwordForm.resetFields();
      message.success('Пароль создан успешно');
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Ошибка создания пароля');
    }
  });

  const deactivatePasswordMutation = useMutation({
    mutationFn: locksAPI.adminDeactivatePassword,
    onSuccess: () => {
      queryClient.invalidateQueries(['lock-passwords', selectedLock?.unique_id]);
      queryClient.invalidateQueries(['admin-locks']);
      message.success('Пароль деактивирован');
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Ошибка деактивации пароля');
    }
  });

  const handleCreate = (values) => {
    // Преобразуем apartment_id в число
    const formattedValues = {
      ...values,
      apartment_id: values.apartment_id ? parseInt(values.apartment_id) : null
    };
    createMutation.mutate(formattedValues);
  };

  const handleUpdate = (values) => {
    updateMutation.mutate({
      id: selectedLock.id,
      ...values
    });
  };

  const handleDelete = (id) => {
    deleteMutation.mutate(id);
  };

  const handleBindToApartment = (lockId, apartmentId) => {
    bindToApartmentMutation.mutate({ lockId, apartmentId });
  };

  const handleUnbindFromApartment = (lockId) => {
    unbindFromApartmentMutation.mutate(lockId);
  };

  const handleEmergencyReset = (lockId) => {
    emergencyResetMutation.mutate(lockId);
  };



  const handleSyncWithTuya = (uniqueId) => {
    syncWithTuyaMutation.mutate(uniqueId);
  };

  // Обработчики для управления паролями
  const handleGeneratePassword = (values) => {
    const data = {
      name: values.name,
      description: values.description,
      valid_from: values.valid_from.toISOString(),
      valid_until: values.valid_until.toISOString(),
      user_id: values.user_id ? parseInt(values.user_id) : null,
    };
    generatePasswordMutation.mutate({
      uniqueId: selectedLock.unique_id,
      data
    });
  };

  const handleDeactivatePassword = (passwordId) => {
    deactivatePasswordMutation.mutate(passwordId);
  };

  const getOnlineStatusColor = (isOnline) => {
    return isOnline ? 'green' : 'red';
  };

  const getOnlineStatusText = (isOnline) => {
    return isOnline ? 'Онлайн' : 'Офлайн';
  };

  const getLockStatusColor = (status) => {
    const colors = {
      'closed': 'blue',
      'open': 'orange'
    };
    return colors[status] || 'gray';
  };

  const getLockStatusText = (status) => {
    const texts = {
      'closed': 'Закрыт',
      'open': 'Открыт'
    };
    return texts[status] || 'Неизвестно';
  };

  const getBatteryIcon = (level) => {
    if (level > 50) return <PoweroffOutlined style={{ color: '#52c41a' }} />;
    if (level > 20) return <PoweroffOutlined style={{ color: '#faad14' }} />;
    return <PoweroffOutlined style={{ color: '#ff4d4f' }} />;
  };

  // Функция для обновления фильтров
  const updateFilter = (key, value) => {
    setFilters(prev => ({
      ...prev,
      [key]: value
    }));
    setCurrentPage(1); // Сбрасываем страницу при изменении фильтра
  };

  // Функция для сброса всех фильтров
  const resetFilters = () => {
    setFilters({
      status: null,
      is_online: null,
      apartment_id: null,
      unbound: null,
      min_battery_level: null,
      max_battery_level: null
    });
    setCurrentPage(1);
  };

  const columns = [
    {
      title: 'Уникальный ID',
      dataIndex: 'unique_id',
      key: 'unique_id',
      render: (id) => (
        <Text strong>{id}</Text>
      ),
    },
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      render: (name) => (
        <Text strong>{name || 'Без названия'}</Text>
      ),
    },
    {
      title: 'Квартира',
      key: 'apartment',
      render: (_, record) => (
        <div>
          {record.apartment_info ? (
            <>
              <div className="font-medium">
                {record.apartment_info.street}, д. {record.apartment_info.building}
              </div>
              <div className="text-gray-500 text-sm">
                кв. {record.apartment_info.apartment_number}
              </div>
            </>
          ) : (
            <Text type="secondary">Не привязан</Text>
          )}
        </div>
      ),
    },
    {
      title: 'Статус',
      key: 'status',
      render: (_, record) => (
        <div className="space-y-1">
          <div>
            <Tag color={getLockStatusColor(record.status)} icon={
              record.status === 'closed' ? <LockOutlined /> : 
              record.status === 'open' ? <UnlockOutlined /> : <LockOutlined />
            }>
              {getLockStatusText(record.status)}
            </Tag>
          </div>
          <div>
            <Tag color={getOnlineStatusColor(record.is_online)} icon={
              record.is_online ? <WifiOutlined /> : <WarningOutlined />
            }>
              {getOnlineStatusText(record.is_online)}
            </Tag>
          </div>
        </div>
      ),
    },
    {
      title: 'Заряд батареи',
      dataIndex: 'battery_level',
      key: 'battery_level',
      render: (level) => (
        <div className="flex items-center space-x-2">
          {getBatteryIcon(level)}
          <span>{level}%</span>
        </div>
      ),
    },
    {
      title: 'Активные пароли',
      dataIndex: 'active_passwords_count',
      key: 'active_passwords_count',
      render: (count) => (
        <Badge count={count || 0} showZero color="orange" />
      ),
    },
    {
      title: 'Последний сигнал',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat', 
      render: (date) => date ? dayjs(date).format('DD.MM HH:mm') : '—',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="Просмотр">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedLock(record);
                setDetailsVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Пароли">
            <Button
              type="text"
              icon={<KeyOutlined />}
              onClick={() => {
                setSelectedLock(record);
                setPasswordsVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="История">
            <Button
              type="text"
              icon={<HistoryOutlined />}
              onClick={() => {
                setSelectedLock(record);
                setHistoryVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedLock(record);
                editForm.setFieldsValue({
                  name: record.name,
                  description: record.description
                });
                setEditModalVisible(true);
              }}
            />
          </Tooltip>
          {!record.apartment_id ? (
            <Tooltip title="Привязать к квартире">
              <Button
                type="text"
                icon={<LockOutlined />}
                onClick={() => {
                  // Здесь можно открыть модальное окно для выбора квартиры
                  const apartmentId = prompt('Введите ID квартиры для привязки:');
                  if (apartmentId) {
                    handleBindToApartment(record.id, parseInt(apartmentId));
                  }
                }}
              />
            </Tooltip>
          ) : (
            <Tooltip title="Отвязать от квартиры">
              <Button
                type="text"
                icon={<UnlockOutlined />}
                onClick={() => handleUnbindFromApartment(record.id)}
              />
            </Tooltip>
          )}
          <Tooltip title="Экстренный сброс">
            <Popconfirm
              title="Выполнить экстренный сброс замка?"
              description="Все пароли будут удалены"
              onConfirm={() => handleEmergencyReset(record.id)}
              okText="Да"
              cancelText="Нет"
            >
              <Button
                type="text"
                danger
                icon={<WarningOutlined />}
              />
            </Popconfirm>
          </Tooltip>
          <Tooltip title="Удалить">
            <Popconfirm
              title="Удалить замок?"
              description="Это действие нельзя отменить"
              onConfirm={() => handleDelete(record.id)}
              okText="Да"
              cancelText="Нет"
            >
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];



  return (
    <div className="space-y-6">
      <div className="mb-6 flex flex-col lg:flex-row lg:justify-between lg:items-center space-y-4 lg:space-y-0">
        <div>
          <Title level={2} className="!mb-2">Управление умными замками</Title>
          <Text type="secondary">
            Мониторинг и управление системой умных замков
          </Text>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
          className="w-full sm:w-auto"
          size="large"
        >
          Добавить замок
        </Button>
      </div>

      {/* Статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Всего замков"
              value={locksStats?.data?.total_locks || 0}
              loading={isLoadingStats}
              prefix={<LockOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Онлайн"
              value={locksStats?.data?.online_locks || 0}
              loading={isLoadingStats}
              prefix={<Badge status="success" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Офлайн"
              value={locksStats?.data?.offline_locks || 0}
              loading={isLoadingStats}
              prefix={<Badge status="error" />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="С ошибками"
              value={locksStats?.data?.error_locks || 0}
              loading={isLoadingStats}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Дополнительная статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Низкий заряд"
              value={locksStats?.data?.low_battery_locks || 0}
              loading={isLoadingStats}
              prefix={<ThunderboltOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Привязанные"
              value={locksStats?.data?.bound_locks || 0}
              loading={isLoadingStats}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#13c2c2' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Непривязанные"
              value={locksStats?.data?.unbound_locks || 0}
              loading={isLoadingStats}
              prefix={<StopOutlined />}
              valueStyle={{ color: '#8c8c8c' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Активные пароли"
              value={locksStats?.data?.total_active_passwords || 0}
              loading={isLoadingStats}
              prefix={<KeyOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Фильтры */}
      <Card className="mb-6">
        <div className="flex flex-wrap gap-4 items-end">
          <div className="flex-1 min-w-[160px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Статус замка</label>
            <Select
              placeholder="Выберите статус"
              className="w-full"
              allowClear
              value={filters.status}
              onChange={(value) => updateFilter('status', value)}
            >
              <Option value="closed">Закрыт</Option>
              <Option value="open">Открыт</Option>
            </Select>
          </div>
          
          <div className="flex-1 min-w-[160px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Подключение</label>
            <Select
              placeholder="Онлайн статус"
              className="w-full"
              allowClear
              value={filters.is_online}
              onChange={(value) => updateFilter('is_online', value)}
            >
              <Option value={true}>Онлайн</Option>
              <Option value={false}>Офлайн</Option>
            </Select>
          </div>
          
          <div className="flex-1 min-w-[150px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">ID квартиры</label>
            <InputNumber
              placeholder="ID квартиры"
              className="w-full"
              value={filters.apartment_id}
              onChange={(value) => updateFilter('apartment_id', value)}
              min={1}
            />
          </div>
          
          <div className="flex-1 min-w-[160px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Привязка</label>
            <Select
              placeholder="Привязка к квартире"
              className="w-full"
              allowClear
              value={filters.unbound}
              onChange={(value) => updateFilter('unbound', value)}
            >
              <Option value={false}>Привязанные</Option>
              <Option value={true}>Непривязанные</Option>
            </Select>
          </div>
          
          <div className="flex-1 min-w-[120px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Мин. заряд %</label>
            <InputNumber
              placeholder="Минимум"
              className="w-full"
              value={filters.min_battery_level}
              onChange={(value) => updateFilter('min_battery_level', value)}
              min={0}
              max={100}
            />
          </div>
          
          <div className="flex-1 min-w-[120px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Макс. заряд %</label>
            <InputNumber
              placeholder="Максимум"
              className="w-full"
              value={filters.max_battery_level}
              onChange={(value) => updateFilter('max_battery_level', value)}
              min={0}
              max={100}
            />
          </div>
          
          <div className="flex-shrink-0">
            <Button 
              onClick={resetFilters}
              className="px-8"
            >
              Сбросить
            </Button>
          </div>
        </div>
      </Card>

      {/* Таблица замков */}
      <Card>
        <Table
          columns={columns}
          dataSource={locksData?.data?.locks || []}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 1400 }}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: locksData?.data?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: !isMobile,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} замков`,
            responsive: true,
            size: isMobile ? 'small' : 'default',
            simple: isMobile,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            },
          }}
          size={isMobile ? 'small' : 'default'}
        />
      </Card>

      {/* Модал создания замка */}
      <Modal
        title="Добавить умный замок"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 600}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreate}
        >
          <Form.Item
            name="unique_id"
            label="Уникальный ID"
            rules={[{ required: true, message: 'Введите уникальный ID' }]}
          >
            <Input placeholder="Уникальный идентификатор замка" />
          </Form.Item>
          
          <Form.Item
            name="apartment_id"
            label="ID квартиры (необязательно)"
          >
            <InputNumber 
              placeholder="ID квартиры для установки замка (необязательно)" 
              style={{ width: '100%' }}
              min={1}
            />
          </Form.Item>

          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true, message: 'Введите название' }]}
          >
            <Input placeholder="Название замка" />
          </Form.Item>

          <Form.Item name="description" label="Описание">
            <Input placeholder="Описание замка" />
          </Form.Item>

          <Form.Item name="firmware_version" label="Версия прошивки">
            <Input placeholder="Версия прошивки замка" />
          </Form.Item>

          <Form.Item 
            name="tuya_device_id" 
            label="Tuya Device ID"
            rules={[{ required: true, message: 'Введите Tuya Device ID' }]}
          >
            <Input placeholder="Идентификатор устройства в системе Tuya" />
          </Form.Item>

          <Form.Item 
            name="owner_password" 
            label="Пароль владельца"
            rules={[{ required: true, message: 'Введите пароль владельца' }]}
          >
            <Input.Password placeholder="Пароль владельца замка" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={createMutation.isPending}
              >
                Добавить
              </Button>
              <Button onClick={() => setCreateModalVisible(false)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Модал редактирования замка */}
      <Modal
        title="Редактировать замок"
        open={editModalVisible}
        onCancel={() => setEditModalVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 600}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdate}
        >
          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true, message: 'Введите название' }]}
          >
            <Input placeholder="Название замка" />
          </Form.Item>

          <Form.Item name="description" label="Описание">
            <Input placeholder="Описание замка" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={updateMutation.isPending}
              >
                Сохранить
              </Button>
              <Button onClick={() => setEditModalVisible(false)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Drawer с деталями замка */}
      <Drawer
        title="Детали замка"
        width={isMobile ? '100%' : 720}
        open={detailsVisible}
        onClose={() => setDetailsVisible(false)}
        placement={isMobile ? 'bottom' : 'right'}
        height={isMobile ? '90%' : undefined}
      >
        {selectedLock && (
          <div>
            <Descriptions 
              column={isMobile ? 1 : 2} 
              bordered
              size={isMobile ? 'small' : 'default'}
            >
              <Descriptions.Item label="Уникальный ID" span={isMobile ? 1 : 2}>
                <Text strong className="font-mono">{selectedLock.unique_id}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Название" span={isMobile ? 1 : 2}>
                <div className="break-words">{selectedLock.name}</div>
              </Descriptions.Item>
              <Descriptions.Item label="Квартира" span={isMobile ? 1 : 2}>
                <div className="break-words">
                  {selectedLock.apartment_info ? 
                    `${selectedLock.apartment_info.street}, д. ${selectedLock.apartment_info.building}, кв. ${selectedLock.apartment_info.apartment_number}` 
                    : 'Не привязан к квартире'
                  }
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Описание">
                <div className="break-words">{selectedLock.description || '—'}</div>
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                <Tag color={getOnlineStatusColor(selectedLock.is_online)}>
                  {getOnlineStatusText(selectedLock.is_online)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Заряд батареи">
                <div className="flex items-center space-x-2">
                  {getBatteryIcon(selectedLock.battery_level)}
                  <span className="font-mono font-bold">
                    {selectedLock.battery_level || 0}%
                  </span>
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Онлайн статус">
                <Tag color={selectedLock.is_online ? 'green' : 'red'}>
                  {selectedLock.is_online ? 'Онлайн' : 'Офлайн'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Активные пароли">
                <Text className="font-mono font-bold">
                  {selectedLock.active_passwords_count || 0}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="Последний сигнал">
                <div className="font-mono text-sm">
                  {selectedLock.last_heartbeat ? (
                    <>
                      {dayjs(selectedLock.last_heartbeat).format('DD.MM.YYYY')}
                      <br />
                      {dayjs(selectedLock.last_heartbeat).format('HH:mm')}
                    </>
                  ) : '—'}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Создан">
                <div className="font-mono text-sm">
                  {dayjs(selectedLock.created_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedLock.created_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Обновлен">
                <div className="font-mono text-sm">
                  {dayjs(selectedLock.updated_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedLock.updated_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="ID квартиры" span={isMobile ? 1 : 2}>
                <Text className="font-mono">
                  {selectedLock.apartment_id || 'Не привязан'}
                </Text>
              </Descriptions.Item>
            </Descriptions>

            {/* Управление замком */}
            <div className="mt-6">
              <Title level={5}>Управление</Title>
              <div className={`${isMobile ? 'space-y-3' : 'space-x-3'} ${isMobile ? '' : 'flex'}`}>
                <Button
                  icon={<ThunderboltOutlined />}
                  onClick={() => handleSyncWithTuya(selectedLock.unique_id)}
                  loading={syncWithTuyaMutation.isPending}
                  className={isMobile ? 'w-full' : ''}
                  type="primary"
                  size="large"
                >
                  Синхронизировать с Tuya
                </Button>
                <Button
                  icon={<KeyOutlined />}
                  onClick={() => {
                    setPasswordsVisible(true);
                  }}
                  className={isMobile ? 'w-full' : ''}
                  type="default"
                  size="large"
                >
                  Управление паролями
                </Button>
              </div>
            </div>
          </div>
        )}
      </Drawer>

      {/* Drawer с историей замка */}
      <Drawer
        title="История активности замка"
        width={isMobile ? '100%' : 720}
        open={historyVisible}
        onClose={() => setHistoryVisible(false)}
        placement={isMobile ? 'bottom' : 'right'}
        height={isMobile ? '90%' : undefined}
      >
        {selectedLock && (
          <div>
            <div className="mb-4">
              <Title level={4}>История замка: {selectedLock.unique_id}</Title>
              <Text type="secondary" className="break-words">{selectedLock.apartment?.title}</Text>
            </div>
            
            <Timeline 
              loading={historyLoading} 
              mode={isMobile ? 'left' : 'alternate'}
              items={historyData?.events?.map((event, index) => ({
                key: index,
                color: event.event_type === 'unlock' ? 'green' : 
                       event.event_type === 'lock' ? 'blue' : 
                       event.event_type === 'error' ? 'red' : 'gray',
                dot: event.event_type === 'unlock' ? <UnlockOutlined /> :
                     event.event_type === 'lock' ? <LockOutlined /> :
                     event.event_type === 'password_generated' ? <KeyOutlined /> :
                     <WarningOutlined />,
                children: (
                  <div>
                    <Text strong className={isMobile ? 'text-sm' : ''}>{event.description}</Text>
                    <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
                      {dayjs(event.timestamp).format('DD.MM.YYYY HH:mm:ss')}
                    </div>
                    {event.user_name && (
                      <div className={`${isMobile ? 'text-xs' : 'text-sm'}`}>
                        Пользователь: {event.user_name}
                      </div>
                    )}
                    {event.details && (
                      <div className={`text-gray-600 mt-1 break-words ${isMobile ? 'text-xs' : 'text-sm'}`}>
                        {event.details}
                      </div>
                    )}
                  </div>
                )
              })) || [{
                children: <Text type="secondary">История отсутствует</Text>
              }]}
            />
          </div>
        )}
      </Drawer>

      {/* Модал управления паролями */}
      <Modal
        title="Управление паролями"
        open={passwordsVisible}
        onCancel={() => setPasswordsVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 1200}
        style={isMobile ? { top: 20 } : {}}
      >
        <div className="space-y-4">
          <div className={`${isMobile ? 'space-y-2' : 'flex justify-between items-center'}`}>
            <Title level={5} className={isMobile ? 'mb-2' : 'mb-0'}>Все пароли замка</Title>
            <div className={`text-sm ${isMobile ? '' : 'text-right'}`}>
              <Text type="secondary">
                Всего: {passwordsData?.data?.total_count || 0}
              </Text>
              {passwordsData?.data?.passwords && (
                <div className="mt-1">
                  <Text type="success">
                    Активных: {passwordsData.data.passwords.filter(p => p.is_active).length}
                  </Text>
                  {' • '}
                  <Text type="secondary">
                    Неактивных: {passwordsData.data.passwords.filter(p => !p.is_active).length}
                  </Text>
                </div>
              )}
            </div>
          </div>
          {passwordsLoading ? (
            <div>Загрузка...</div>
          ) : passwordsData?.data?.passwords ? (
            <Table
              columns={[
                {
                  title: 'ID',
                  dataIndex: 'id',
                  key: 'id',
                  width: 50,
                  responsive: ['lg'],
                },
                {
                  title: 'Название',
                  dataIndex: 'name',
                  key: 'name',
                  width: isMobile ? 80 : 120,
                  ellipsis: true,
                },
                {
                  title: 'Пароль',
                  dataIndex: 'password',
                  key: 'password',
                  width: isMobile ? 80 : 90,
                  render: (password) => (
                    <Text code copyable={{ text: password }} className={isMobile ? 'text-xs' : 'text-sm'}>
                      {password}
                    </Text>
                  ),
                },
                {
                  title: 'Статус',
                  dataIndex: 'is_active',
                  key: 'is_active',
                  width: isMobile ? 70 : 80,
                  render: (isActive) => (
                    <Tag color={isActive ? 'green' : 'red'} size="small">
                      {isMobile ? (isActive ? 'Акт' : 'Неакт') : (isActive ? 'Активен' : 'Неактивен')}
                    </Tag>
                  ),
                },
                {
                  title: 'Тип',
                  dataIndex: 'binding_type',
                  key: 'binding_type',
                  width: 90,
                  responsive: ['md'],
                  render: (type) => (
                    <Tag color={type === 'user' ? 'blue' : 'orange'} size="small">
                      {type === 'user' ? 'Пользователь' : 'Бронирование'}
                    </Tag>
                  ),
                },
                {
                  title: 'Привязка',
                  dataIndex: 'binding_info',
                  key: 'binding_info',
                  width: 180,
                  responsive: ['lg'],
                  render: (info, record) => {
                    if (!info) return '—';
                    
                    if (record.binding_type === 'user') {
                      return (
                        <div className="text-xs">
                          <div className="font-medium truncate">{info.user_name}</div>
                          <div className="text-gray-500 truncate">{info.user_email}</div>
                          <div className="text-gray-400">ID: {info.user_id}</div>
                        </div>
                      );
                    } else if (record.binding_type === 'booking') {
                      return (
                        <div className="text-xs">
                          <div className="font-medium">#{info.booking_id}</div>
                          <div className="text-gray-500">{info.booking_dates}</div>
                          <div className="text-gray-400">Арендатор: {info.renter_id}</div>
                        </div>
                      );
                    }
                    return '—';
                  },
                },
                {
                  title: 'Действует с',
                  dataIndex: 'valid_from',
                  key: 'valid_from',
                  width: 110,
                  render: (date) => date ? (
                    <div className="text-xs">
                      <div>{dayjs(date).format('DD.MM.YYYY')}</div>
                      <div className="font-mono">{dayjs(date).format('HH:mm')}</div>
                    </div>
                  ) : '—',
                  responsive: ['md'],
                },
                {
                  title: 'Действует до',
                  dataIndex: 'valid_until',
                  key: 'valid_until',
                  width: 110,
                  render: (date) => date ? (
                    <div className="text-xs">
                      <div>{dayjs(date).format('DD.MM.YYYY')}</div>
                      <div className="font-mono">{dayjs(date).format('HH:mm')}</div>
                    </div>
                  ) : '—',
                  responsive: ['md'],
                },
                {
                  title: 'Период действия',
                  key: 'dates_mobile',
                  responsive: ['xs', 'sm'],
                  render: (_, record) => (
                    <div className="text-xs">
                      <div className="font-medium">
                        С: {dayjs(record.valid_from).format('DD.MM HH:mm')}
                      </div>
                      <div className="font-medium">
                        До: {dayjs(record.valid_until).format('DD.MM HH:mm')}
                      </div>
                      <div className="text-gray-500 mt-1">
                        {record.binding_type === 'user' && record.binding_info?.user_name && (
                          <span className="truncate">{record.binding_info.user_name}</span>
                        )}
                        {record.binding_type === 'booking' && record.binding_info?.booking_id && (
                          <span>#{record.binding_info.booking_id}</span>
                        )}
                      </div>
                    </div>
                  ),
                },
                {
                  title: 'Создан',
                  dataIndex: 'created_at',
                  key: 'created_at',
                  width: 110,
                  render: (date) => (
                    <div className="text-xs">
                      <div>{dayjs(date).format('DD.MM.YYYY')}</div>
                      <div className="text-gray-500">{dayjs(date).format('HH:mm')}</div>
                    </div>
                  ),
                  responsive: ['xl'],
                },
                {
                  title: 'Действия',
                  key: 'actions',
                  width: isMobile ? 80 : 100,
                  fixed: 'right',
                  render: (_, record) => (
                    <div className="text-center">
                      {record.is_active ? (
                        <Tooltip title="Деактивировать пароль">
                          <Button
                            type="text"
                            icon={<StopOutlined />}
                            onClick={() => handleDeactivatePassword(record.id)}
                            danger
                            size="small"
                          />
                        </Tooltip>
                      ) : (
                        <Tag color="red" size="small">
                          {isMobile ? 'Неакт' : 'Неактивен'}
                        </Tag>
                      )}
                    </div>
                  ),
                },
              ]}
              dataSource={passwordsData.data.passwords}
              loading={passwordsLoading}
              rowKey="id"
              size="small"
              scroll={{ x: isMobile ? 600 : 1000, y: isMobile ? 300 : 400 }}
              pagination={{
                size: 'small',
                simple: isMobile,
                showSizeChanger: !isMobile,
                pageSize: 8,
                showQuickJumper: !isMobile,
                showTotal: (total, range) => 
                  `${range[0]}-${range[1]} из ${total} паролей`,
              }}
            />
          ) : (
            <Text type="secondary">Нет паролей для этого замка.</Text>
          )}

          <Title level={5} className="mt-6">Генерация нового пароля</Title>
          <Form
            form={passwordForm}
            layout="vertical"
            onFinish={handleGeneratePassword}
          >
            <Row gutter={[16, 0]}>
              <Col xs={24} sm={12}>
                <Form.Item
                  name="name"
                  label="Название пароля"
                  rules={[{ required: true, message: 'Введите название пароля' }]}
                >
                  <Input placeholder="Например, 'Доступ к замку'" />
                </Form.Item>
              </Col>
              <Col xs={24} sm={12}>
                <Form.Item
                  name="user_id"
                  label="ID пользователя (опционально)"
                >
                  <Input placeholder="ID пользователя" type="number" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={[16, 0]}>
              <Col span={24}>
                <Form.Item
                  name="description"
                  label="Описание"
                  rules={[{ required: true, message: 'Введите описание пароля' }]}
                >
                  <Input.TextArea 
                    placeholder="Описание назначения пароля"
                    rows={isMobile ? 3 : 2}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={[16, 0]}>
              <Col xs={24} sm={12}>
                <Form.Item
                  name="valid_from"
                  label="Действителен с (дата и время)"
                  rules={[{ required: true, message: 'Выберите дату и время начала' }]}
                >
                  <DatePicker 
                    showTime={{ 
                      format: 'HH:mm',
                      minuteStep: 15 
                    }}
                    format="DD.MM.YYYY HH:mm"
                    placeholder="Выберите дату и время начала" 
                    className="w-full"
                    size={isMobile ? 'large' : 'middle'}
                  />
                </Form.Item>
              </Col>
              <Col xs={24} sm={12}>
                <Form.Item
                  name="valid_until"
                  label="Действителен до (дата и время)"
                  rules={[{ required: true, message: 'Выберите дату и время окончания' }]}
                >
                  <DatePicker 
                    showTime={{ 
                      format: 'HH:mm',
                      minuteStep: 15 
                    }}
                    format="DD.MM.YYYY HH:mm"
                    placeholder="Выберите дату и время окончания" 
                    className="w-full"
                    size={isMobile ? 'large' : 'middle'}
                  />
                </Form.Item>
              </Col>
            </Row>
            <div className={`bg-blue-50 p-4 rounded mb-4 ${isMobile ? 'text-sm' : ''}`}>
              <Text type="secondary" className={isMobile ? 'text-xs' : 'text-sm'}>
                <strong>Важно:</strong> Обязательно указывайте точное время! Пароль будет активен строго в указанный период.
                {isMobile ? (
                  <br />
                ) : ' '}
                Рекомендуется давать пароли с запасом времени (например, за 15-30 минут до заселения).
              </Text>
            </div>
            <Form.Item>
              <Space className="w-full" direction={isMobile ? 'vertical' : 'horizontal'}>
                <Button 
                  type="primary" 
                  htmlType="submit"
                  loading={generatePasswordMutation.isPending}
                  className={isMobile ? 'w-full' : ''}
                  icon={<KeyOutlined />}
                  size={isMobile ? 'large' : 'middle'}
                >
                  Сгенерировать пароль
                </Button>
                <Button 
                  onClick={() => {
                    setPasswordsVisible(false);
                    passwordForm.resetFields();
                  }}
                  className={isMobile ? 'w-full' : ''}
                  size={isMobile ? 'large' : 'middle'}
                >
                  Отмена
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </div>
      </Modal>
    </div>
  );
};

export default LocksPage; 