import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  EyeOutlined,
  FileImageOutlined,
  HistoryOutlined,
  HomeOutlined,
  IdcardOutlined,
  UserOutlined
} from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { 
  Avatar, Button, Card, Col, Descriptions, Divider, Form, Image, 
  Input, message, Modal, Popconfirm, Row, Select, Space, Statistic, 
  Table, Tag, Tooltip, Typography 
} from 'antd';
import dayjs from 'dayjs';
import 'dayjs/locale/ru';
import React, { useEffect, useState } from 'react';

// Устанавливаем русскую локаль для dayjs
dayjs.locale('ru');
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import LocationFilter from '../../components/LocationFilter.jsx';
import { usersAPI } from '../../lib/api.js';
import UserBookingHistoryModal from '../../components/UserBookingHistoryModal.jsx';

const { Option } = Select;
const { Title, Text } = Typography;

const UsersPage = () => {
  // Состояния
  const [searchText, setSearchText] = useState('');
  const [selectedRole, setSelectedRole] = useState('all');
  const [verificationFilter, setVerificationFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [cityFilter, setCityFilter] = useState(null);
  const [selectedUser, setSelectedUser] = useState(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailsModalVisible, setDetailsModalVisible] = useState(false);
  const [selectedUserDetails, setSelectedUserDetails] = useState(null);
  const [editingUserDetails, setEditingUserDetails] = useState(null);
  const [historyModalVisible, setHistoryModalVisible] = useState(false);
  const [selectedUserForHistory, setSelectedUserForHistory] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [form] = Form.useForm();
  
  const queryClient = useQueryClient();

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получаем список пользователей
  const { data: users, isLoading } = useQuery({
    queryKey: ['admin-users', searchText, selectedRole, verificationFilter, statusFilter, cityFilter, currentPage, pageSize],
    queryFn: () => {
      const params = {
        page: currentPage,
        page_size: pageSize,
      };
      if (searchText) params.search = searchText;
      if (selectedRole && selectedRole !== 'all') params.role = selectedRole;
      if (verificationFilter && verificationFilter !== 'all') params.verification_status = verificationFilter;
      if (statusFilter && statusFilter !== 'all') params.is_active = statusFilter === 'active';
      if (cityFilter) params.city_id = cityFilter;
      
      return usersAPI.adminGetAllUsers(params);
    },
  });

  // Получаем статистику пользователей
  const { data: userStatistics, isLoading: isLoadingStatistics } = useQuery({
    queryKey: ['admin-users-statistics'],
    queryFn: () => usersAPI.adminGetUsersStatistics(),
  });

  // Получаем детальную информацию о пользователе для просмотра
  const { data: userDetails, isLoading: isLoadingDetails } = useQuery({
    queryKey: ['admin-user-details', selectedUserDetails?.id],
    queryFn: () => usersAPI.adminGetUserById(selectedUserDetails.id),
    enabled: !!selectedUserDetails?.id,
  });

  // Получаем детальную информацию о пользователе для редактирования
  const { data: editUserDetails, isLoading: isLoadingEditDetails } = useQuery({
    queryKey: ['admin-user-edit-details', editingUserDetails?.id],
    queryFn: () => usersAPI.adminGetUserById(editingUserDetails.id),
    enabled: !!editingUserDetails?.id,
  });

  // Обновляем форму когда загрузятся детальные данные пользователя для редактирования
  useEffect(() => {
    if (editUserDetails?.data && modalVisible) {
      const userData = editUserDetails.data.user;
      form.setFieldsValue({
        ...userData,
        city_id: userData.city?.id || userData.city_id,
      });
    }
  }, [editUserDetails, modalVisible, form]);

  // Мутации
  const updateVerificationMutation = useMutation({
    mutationFn: ({ userId, status }) => usersAPI.updateRenterVerificationStatus(userId, status),
    onSuccess: () => {
      message.success('Статус верификации обновлен');
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      queryClient.invalidateQueries({ queryKey: ['admin-user-details'] });
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Ошибка при обновлении статуса');
    },
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }) => usersAPI.adminUpdateUserRole(userId, role),
    onSuccess: () => {
      message.success('Роль пользователя обновлена');
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
    },
    onError: () => {
      message.error('Ошибка при обновлении роли');
    },
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ userId, isActive, reason }) => usersAPI.adminUpdateUserStatus(userId, isActive, reason),
    onSuccess: () => {
      message.success('Статус пользователя обновлен');
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
    },
    onError: () => {
      message.error('Ошибка при обновлении статуса');
    },
  });

  const deleteUserMutation = useMutation({
    mutationFn: (userId) => usersAPI.deleteUserByAdmin(userId),
    onSuccess: () => {
      message.success('Пользователь удален');
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
    },
    onError: () => {
      message.error('Ошибка при удалении пользователя');
    },
  });

  // Обработчики
  const handleVerificationUpdate = (userId, status) => {
    updateVerificationMutation.mutate({ userId, status });
  };

  const handleRoleUpdate = (userId, role) => {
    updateRoleMutation.mutate({ userId, role });
  };

  const handleStatusUpdate = (userId, isActive, reason = '') => {
    updateStatusMutation.mutate({ userId, isActive, reason });
  };

  const handleDeleteUser = (userId) => {
    deleteUserMutation.mutate(userId);
  };

  const handleEditUser = (user) => {
    setSelectedUser(user);
    setEditingUserDetails(user);
    form.setFieldsValue(user);
    setModalVisible(true);
  };

  const handleViewUserDetails = (user) => {
    setSelectedUserDetails(user);
    setDetailsModalVisible(true);
  };

  const handleModalCancel = () => {
    setModalVisible(false);
    setSelectedUser(null);
    setEditingUserDetails(null);
    form.resetFields();
  };

  const handleDetailsModalCancel = () => {
    setDetailsModalVisible(false);
    setSelectedUserDetails(null);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      
      if (values.role !== selectedUser.role) {
        await handleRoleUpdate(selectedUser.id, values.role);
      }
      
      if (values.is_active !== undefined && values.is_active !== selectedUser.is_active) {
        await handleStatusUpdate(selectedUser.id, values.is_active, values.reason);
      }
      
      message.success('Пользователь обновлен');
      setModalVisible(false);
      setSelectedUser(null);
      form.resetFields();
    } catch (error) {
      console.error('Ошибка при обновлении пользователя:', error);
      message.error('Ошибка при обновлении пользователя');
    }
  };

  // Утилиты
  const getRoleColor = (role) => {
    switch (role) {
      case 'admin':
        return 'red';
      case 'moderator':
        return 'purple';
      case 'owner':
        return 'blue';
      case 'concierge':
        return 'cyan';
      case 'user':
        return 'green';
      default:
        return 'default';
    }
  };

  const getRoleText = (role) => {
    switch (role) {
      case 'admin':
        return 'Администратор';
      case 'moderator':
        return 'Модератор';
      case 'owner':
        return 'Владелец недвижимости';
      case 'concierge':
        return 'Консьерж';
      case 'user':
        return 'Пользователь';
      default:
        return role;
    }
  };

  const getVerificationStatusText = (status) => {
    switch (status) {
      case 'pending':
        return 'На проверке';
      case 'approved':
        return 'Верифицирован';
      case 'rejected':
        return 'Отклонен';
      default:
        return status;
    }
  };

  const getVerificationStatusColor = (status) => {
    switch (status) {
      case 'pending':
        return 'orange';
      case 'approved':
        return 'green';
      case 'rejected':
        return 'red';
      default:
        return 'default';
    }
  };

  // Колонки таблицы
  const columns = [
    {
      title: 'Пользователь',
      key: 'user',
      render: (record) => (
        <div className="flex items-center space-x-3">
          <Avatar size="large" icon={<UserOutlined />} />
          <div>
            <div className="font-medium">
              {record.first_name} {record.last_name}
            </div>
            <div className="text-sm text-gray-500">{record.email}</div>
            {record.phone && (
              <div className="text-sm text-gray-500">{record.phone}</div>
            )}
          </div>
        </div>
      ),
    },
    {
      title: 'Город',
      key: 'city',
      render: (record) => (
        <div className="text-sm">
          {record.city?.name || '—'}
        </div>
      ),
    },
    {
      title: 'Роль',
      dataIndex: 'role',
      key: 'role',
      render: (role) => (
        <Tag color={getRoleColor(role)}>
          {getRoleText(role)}
        </Tag>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? 'Активен' : 'Заблокирован'}
        </Tag>
      ),
    },
    {
      title: 'Верификация',
      key: 'verification',
      render: (record) => {
        if (!record.renter_info) {
          return <span className="text-gray-400">—</span>;
        }
        
        const status = record.renter_info.verification_status;
        let color = 'default';
        let text = 'Неизвестно';
        
        if (status === 'approved') {
          color = 'green';
          text = 'Верифицирован';
        } else if (status === 'pending') {
          color = 'orange';
          text = 'На проверке';
        } else if (status === 'rejected') {
          color = 'red';
          text = 'Отклонен';
        }
        
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: 'Дата регистрации',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD.MM.YYYY HH:mm'),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (record) => (
        <Space>
          <Tooltip title="Просмотр">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewUserDetails(record)}
            />
          </Tooltip>
          <Tooltip title="История бронирований">
            <Button
              type="text"
              icon={<HistoryOutlined />}
              onClick={() => {
                setSelectedUserForHistory(record);
                setHistoryModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEditUser(record)}
            />
          </Tooltip>
          <Tooltip title="Изменить статус верификации">
            {record.renter_info?.verification_status === 'pending' && (
              <Space>
                <Button
                  type="text"
                  className="text-green-600"
                  onClick={() => handleVerificationUpdate(record.id, 'approved')}
                >
                  ✓
                </Button>
                <Button
                  type="text"
                  className="text-red-600"
                  onClick={() => handleVerificationUpdate(record.id, 'rejected')}
                >
                  ✗
                </Button>
              </Space>
            )}
          </Tooltip>
          <Popconfirm
            title="Удалить пользователя?"
            description="Это действие нельзя отменить"
            onConfirm={() => handleDeleteUser(record.id)}
          >
            <Tooltip title="Удалить">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="space-y-6">
      <div>
        <Title level={2} className="!mb-2">Управление пользователями</Title>
        <Text type="secondary">
          Просмотр и управление пользователями системы
        </Text>
      </div>

      {/* Основная статистика */}
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={12} md={6} lg={4}>
          <Card>
            <Statistic
              title="Всего пользователей"
              value={userStatistics?.data?.summary?.total_users || 0}
              valueStyle={{ color: '#1890ff', fontSize: isMobile ? '20px' : '24px' }}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={4}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setStatusFilter('active');
              setCurrentPage(1);
            }}
          >
            <Statistic
              title="Активные"
              value={userStatistics?.data?.summary?.active_users || 0}
              valueStyle={{ color: '#52c41a', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={4}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setStatusFilter('inactive');
              setCurrentPage(1);
            }}
          >
            <Statistic
              title="Неактивные"
              value={userStatistics?.data?.summary?.inactive_users || 0}
              valueStyle={{ color: '#f5222d', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={4}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setSelectedRole('user');
              setCurrentPage(1);
            }}
          >
            <Statistic
              title="Арендаторы"
              value={userStatistics?.data?.summary?.total_renters || 0}
              valueStyle={{ color: '#fa8c16', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={4}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setSelectedRole('owner');
              setCurrentPage(1);
            }}
          >
            <Statistic
              title="Владельцы"
              value={userStatistics?.data?.summary?.total_owners || 0}
              valueStyle={{ color: '#722ed1', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Детальная статистика */}
      <Row gutter={[16, 16]}>
        {/* Статистика по городам */}
        <Col xs={24} sm={24} md={12} lg={8}>
          <Card 
            size="small" 
            title="Распределение по городам"
            className="h-full"
          >
            <div className="space-y-2">
              {userStatistics?.data?.by_city && Object.keys(userStatistics.data.by_city).length > 0 ? (
                Object.entries(userStatistics.data.by_city).map(([city, count]) => (
                  <div key={city} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                    <span className="text-sm font-medium">{city}</span>
                    <Tag color="blue">{count}</Tag>
                  </div>
                ))
              ) : (
                <div className="text-center text-gray-500 py-4">
                  Нет данных
                </div>
              )}
            </div>
          </Card>
        </Col>

        {/* Статистика по ролям */}
        <Col xs={24} sm={24} md={12} lg={8}>
          <Card 
            size="small" 
            title="Распределение по ролям"
            className="h-full"
          >
            <div className="space-y-2">
              {userStatistics?.data?.by_role && Object.keys(userStatistics.data.by_role).length > 0 ? (
                Object.entries(userStatistics.data.by_role).map(([role, count]) => (
                  <div key={role} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                    <span className="text-sm font-medium">{getRoleText(role)}</span>
                    <Tag color={getRoleColor(role)}>{count}</Tag>
                  </div>
                ))
              ) : (
                <div className="text-center text-gray-500 py-4">
                  Нет данных
                </div>
              )}
            </div>
          </Card>
        </Col>

        {/* Статистика по верификации */}
        <Col xs={24} sm={24} md={12} lg={8}>
          <Card 
            size="small" 
            title="Верификация арендаторов"
            className="h-full"
          >
            <div className="space-y-2">
              {userStatistics?.data?.renter_verification && Object.keys(userStatistics.data.renter_verification).length > 0 ? (
                Object.entries(userStatistics.data.renter_verification).map(([status, count]) => (
                  <div key={status} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                    <span className="text-sm font-medium">{getVerificationStatusText(status)}</span>
                    <Tag color={getVerificationStatusColor(status)}>{count}</Tag>
                  </div>
                ))
              ) : (
                <div className="text-center text-gray-500 py-4">
                  Нет данных
                </div>
              )}
            </div>
          </Card>
        </Col>

        {/* Статистика по месяцам регистрации */}
        <Col xs={24} sm={24} md={12} lg={12}>
          <Card 
            size="small" 
            title="Регистрации по месяцам"
            className="h-full"
          >
            <div className="space-y-2">
              {userStatistics?.data?.by_month_registration && Object.keys(userStatistics.data.by_month_registration).length > 0 ? (
                Object.entries(userStatistics.data.by_month_registration)
                  .sort(([a], [b]) => b.localeCompare(a)) // Сортировка по убыванию (новые месяцы сверху)
                  .map(([month, count]) => (
                    <div key={month} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span className="text-sm font-medium">
                        {dayjs(month + '-01').format('MMMM YYYY')}
                      </span>
                      <Tag color="purple">{count}</Tag>
                    </div>
                  ))
              ) : (
                <div className="text-center text-gray-500 py-4">
                  Нет данных
                </div>
              )}
            </div>
          </Card>
        </Col>

        {/* Статистика по статусам */}
        <Col xs={24} sm={24} md={12} lg={12}>
          <Card 
            size="small" 
            title="Распределение по статусам"
            className="h-full"
          >
            <div className="space-y-2">
              {userStatistics?.data?.by_status && Object.keys(userStatistics.data.by_status).length > 0 ? (
                Object.entries(userStatistics.data.by_status).map(([status, count]) => (
                  <div key={status} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                    <span className="text-sm font-medium">
                      {status === 'active' ? 'Активные' : 'Неактивные'}
                    </span>
                    <Tag color={status === 'active' ? 'green' : 'red'}>{count}</Tag>
                  </div>
                ))
              ) : (
                <div className="text-center text-gray-500 py-4">
                  Нет данных
                </div>
              )}
            </div>
          </Card>
        </Col>
      </Row>

      {/* Фильтры */}
      <Card>
        <div className="flex flex-wrap gap-4 items-end">
          <div className="flex-1 min-w-[200px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Поиск</label>
            <Input
              placeholder="Поиск по имени или email"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              allowClear
            />
          </div>
          
          <div className="flex-1 min-w-[160px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Роль</label>
            <Select
              value={selectedRole}
              onChange={setSelectedRole}
              className="w-full"
              placeholder="Выберите роль"
            >
              <Option value="all">Все роли</Option>
              <Option value="admin">Администратор</Option>
              <Option value="moderator">Модератор</Option>
              <Option value="owner">Владелец</Option>
              <Option value="concierge">Консьерж</Option>
              <Option value="user">Пользователь</Option>
            </Select>
          </div>
          
          <div className="flex-1 min-w-[160px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Верификация</label>
            <Select
              value={verificationFilter}
              onChange={setVerificationFilter}
              className="w-full"
              placeholder="Статус верификации"
            >
              <Option value="all">Все статусы</Option>
              <Option value="approved">Верифицированы</Option>
              <Option value="pending">На проверке</Option>
              <Option value="rejected">Отклонены</Option>
            </Select>
          </div>
          
          <div className="flex-1 min-w-[160px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Активность</label>
            <Select
              value={statusFilter}
              onChange={setStatusFilter}
              className="w-full"
              placeholder="Статус активности"
            >
              <Option value="all">Все активности</Option>
              <Option value="active">Активные</Option>
              <Option value="inactive">Неактивные</Option>
            </Select>
          </div>
          
          <div className="flex-1 min-w-[160px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Город</label>
            <LocationFilter
              showCity={true}
              showDistrict={false}
              showMicrodistrict={false}
              cityId={cityFilter}
              onCityChange={setCityFilter}
              layout="vertical"
              showLabels={false}
            />
          </div>
          
          <div className="flex-shrink-0">
            <Button 
              onClick={() => {
                setSearchText('');
                setSelectedRole('all');
                setVerificationFilter('all');
                setStatusFilter('all');
                setCityFilter(null);
              }}
              className="px-8"
            >
              Сбросить
            </Button>
          </div>
        </div>
      </Card>

      {/* Таблица */}
      <Card>
        <Table
          columns={columns}
          dataSource={users?.data?.users || []}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 1200 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: users?.data?.pagination?.total || 0,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            },
            showSizeChanger: !isMobile,
            showQuickJumper: !isMobile,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} пользователей`,
            responsive: true,
            simple: isMobile,
          }}
        />
      </Card>

      {/* Модал редактирования */}
      <Modal
        title="Редактировать пользователя"
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={handleModalCancel}
        width={isMobile ? '95%' : 600}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form form={form} layout="vertical">
          <Row gutter={[16, 0]}>
            <Col xs={24} sm={12}>
              <Form.Item label="Имя" name="first_name">
                <Input disabled />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item label="Фамилия" name="last_name">
                <Input disabled />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label="Email" name="email">
            <Input disabled />
          </Form.Item>
          <Form.Item label="Телефон" name="phone">
            <Input disabled />
          </Form.Item>
          <Row gutter={[16, 0]}>
            <Col xs={24} sm={12}>
              <Form.Item label="Роль" name="role">
                <Select>
                                  <Option value="admin">Администратор</Option>
                <Option value="moderator">Модератор</Option>
                <Option value="owner">Владелец недвижимости</Option>
                <Option value="concierge">Консьерж</Option>
                <Option value="user">Пользователь</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item label="Статус" name="is_active">
                <Select>
                  <Option value={true}>Активен</Option>
                  <Option value={false}>Заблокирован</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      {/* Модал детального просмотра */}
      <Modal
        title="Детальная информация о пользователе"
        open={detailsModalVisible}
        onCancel={handleDetailsModalCancel}
        footer={null}
        width={isMobile ? '95%' : 800}
        style={isMobile ? { top: 20 } : {}}
      >
        {isLoadingDetails ? (
          <div className="text-center py-8">
            <LoadingSpinner />
          </div>
        ) : userDetails?.data ? (
          <div className="space-y-4">
            {/* Основная информация */}
            <Card size="small" title="Основная информация">
              <Descriptions 
                bordered 
                column={isMobile ? 1 : 2} 
                size="small"
              >
                <Descriptions.Item label="ID">{userDetails.data.user.id}</Descriptions.Item>
                <Descriptions.Item label="Email">
                  <div className="break-words">{userDetails.data.user.email}</div>
                </Descriptions.Item>
                <Descriptions.Item label="Имя">{userDetails.data.user.first_name}</Descriptions.Item>
                <Descriptions.Item label="Фамилия">{userDetails.data.user.last_name}</Descriptions.Item>
                <Descriptions.Item label="Телефон">{userDetails.data.user.phone || '—'}</Descriptions.Item>
                <Descriptions.Item label="Город">
                  {userDetails.data.user.city?.name || '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Роль">
                  <Tag color={getRoleColor(userDetails.data.user.role)}>
                    {getRoleText(userDetails.data.user.role)}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Статус">
                  <Tag color={userDetails.data.user.is_active ? 'green' : 'red'}>
                    {userDetails.data.user.is_active ? 'Активен' : 'Заблокирован'}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Дата регистрации" span={isMobile ? 1 : 2}>
                  <div className="font-mono text-sm">
                    {dayjs(userDetails.data.user.created_at).format('DD.MM.YYYY')}
                    <br />
                    {dayjs(userDetails.data.user.created_at).format('HH:mm')}
                  </div>
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {/* Информация об арендаторе */}
            {userDetails.data.renter_info && (
              <Card 
                size="small" 
                title={
                  <div className="flex items-center justify-between">
                    <span className="flex items-center space-x-2">
                      <IdcardOutlined />
                      <span>Информация арендатора</span>
                    </span>
                    <Tag 
                      color={getVerificationStatusColor(userDetails.data.renter_info.verification_status)}
                      icon={
                        userDetails.data.renter_info.verification_status === 'approved' ? <CheckCircleOutlined /> :
                        userDetails.data.renter_info.verification_status === 'rejected' ? <CloseCircleOutlined /> : <ExclamationCircleOutlined />
                      }
                    >
                      {getVerificationStatusText(userDetails.data.renter_info.verification_status)}
                    </Tag>
                  </div>
                }
              >
                <div className="space-y-4">
                  <Descriptions 
                    bordered 
                    column={isMobile ? 1 : 2} 
                    size="small"
                  >
                    <Descriptions.Item label="ID арендатора">{userDetails.data.renter_info.id}</Descriptions.Item>
                    <Descriptions.Item label="Тип документа">
                      {userDetails.data.renter_info.document_type === 'udv' ? 'Удостоверение личности' : 'Паспорт'}
                    </Descriptions.Item>
                    <Descriptions.Item label="Дата создания" span={isMobile ? 1 : 2}>
                      <div className="font-mono text-sm">
                        {dayjs(userDetails.data.renter_info.created_at).format('DD.MM.YYYY')}
                        <br />
                        {dayjs(userDetails.data.renter_info.created_at).format('HH:mm')}
                      </div>
                    </Descriptions.Item>
                  </Descriptions>

                  {/* Документы */}
                  <div className="space-y-3">
                    <Divider orientation="left" orientationMargin="0">
                      <span className="flex items-center space-x-1">
                        <FileImageOutlined />
                        <span>Документы</span>
                      </span>
                    </Divider>
                    
                    <Row gutter={[16, 16]}>
                      {userDetails.data.renter_info.document_url && Object.keys(userDetails.data.renter_info.document_url).length > 0 && (
                        <Col xs={24} sm={24} md={userDetails.data.renter_info.photo_with_doc_url ? 12 : 24}>
                          <Card size="small" title="Документы">
                            <div className="space-y-3">
                              {Object.entries(userDetails.data.renter_info.document_url).map(([page, url], index) => (
                                <div key={page}>
                                  <div className="text-sm text-gray-600 mb-2">
                                    {Object.keys(userDetails.data.renter_info.document_url).length > 1 
                                      ? `Страница ${index + 1}` 
                                      : 'Документ'}
                                  </div>
                                  <Image
                                    width="100%"
                                    src={url}
                                    alt={`Документ - ${page}`}
                                    placeholder={<div className="text-center p-4">Загрузка...</div>}
                                    fallback={
                                      <div className="flex flex-col items-center justify-center p-8 bg-gray-100 text-gray-500 rounded border-2 border-dashed border-gray-300">
                                        <FileImageOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
                                        <div className="mt-2 text-sm">Документ недоступен</div>
                                        <div className="text-xs text-gray-400 break-all">
                                          {url}
                                        </div>
                                      </div>
                                    }
                                  />
                                </div>
                              ))}
                            </div>
                          </Card>
                        </Col>
                      )}
                      
                      {userDetails.data.renter_info.photo_with_doc_url && (
                        <Col xs={24} sm={24} md={userDetails.data.renter_info.document_url && Object.keys(userDetails.data.renter_info.document_url).length > 0 ? 12 : 24}>
                          <Card size="small" title="Фото с документом">
                            <Image
                              width="100%"
                              src={userDetails.data.renter_info.photo_with_doc_url}
                              alt="Фото с документом"
                              placeholder={<div className="text-center p-4">Загрузка...</div>}
                              fallback={
                                <div className="flex flex-col items-center justify-center p-8 bg-gray-100 text-gray-500 rounded border-2 border-dashed border-gray-300">
                                  <UserOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
                                  <div className="mt-2 text-sm">Фото недоступно</div>
                                  <div className="text-xs text-gray-400 break-all">
                                    {userDetails.data.renter_info.photo_with_doc_url}
                                  </div>
                                </div>
                              }
                            />
                          </Card>
                        </Col>
                      )}
                    </Row>

                    {/* Действия по верификации */}
                    {userDetails.data.renter_info.verification_status === 'pending' && (
                      <div className="mt-4 p-4 bg-yellow-50 border border-yellow-200 rounded">
                        <div className="text-center space-y-3">
                          <div className="text-sm text-gray-600 mb-3">
                            Пользователь ожидает верификации. Проверьте документы и примите решение:
                          </div>
                          <Space 
                            size="large" 
                            direction={isMobile ? 'vertical' : 'horizontal'}
                            className={isMobile ? 'w-full' : ''}
                          >
                            <Button
                              type="primary"
                              icon={<CheckCircleOutlined />}
                              onClick={() => {
                                handleVerificationUpdate(userDetails.data.user.id, 'approved');
                                handleDetailsModalCancel();
                              }}
                              loading={updateVerificationMutation.isLoading}
                              className={`bg-green-600 hover:bg-green-700 ${isMobile ? 'w-full' : ''}`}
                            >
                              Одобрить верификацию
                            </Button>
                            <Button
                              danger
                              icon={<CloseCircleOutlined />}
                              onClick={() => {
                                handleVerificationUpdate(userDetails.data.user.id, 'rejected');
                                handleDetailsModalCancel();
                              }}
                              loading={updateVerificationMutation.isLoading}
                              className={isMobile ? 'w-full' : ''}
                            >
                              Отклонить верификацию
                            </Button>
                          </Space>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </Card>
            )}

            {/* Статистика */}
            <Card size="small" title="Статистика">
              <Row gutter={[16, 16]}>
                <Col xs={24} sm={8}>
                  <div className="text-center p-4 bg-blue-50 rounded">
                    <div className="text-2xl font-bold text-blue-600">
                      {userDetails.data.apartments_count || 0}
                    </div>
                    <div className="text-sm text-gray-600">Квартир</div>
                  </div>
                </Col>
                <Col xs={24} sm={8}>
                  <div className="text-center p-4 bg-green-50 rounded">
                    <div className="text-2xl font-bold text-green-600">
                      {userDetails.data.bookings_count || 0}
                    </div>
                    <div className="text-sm text-gray-600">Бронирований</div>
                  </div>
                </Col>
                <Col xs={24} sm={8}>
                  <div className="text-center p-4 bg-purple-50 rounded">
                    <div className="text-2xl font-bold text-purple-600">
                      {userDetails.data.property_owner_info ? 'Да' : 'Нет'}
                    </div>
                    <div className="text-sm text-gray-600">Владелец недвижимости</div>
                  </div>
                </Col>
              </Row>
            </Card>
          </div>
        ) : (
          <div className="text-center py-8">
            Данные не найдены
          </div>
        )}
      </Modal>

      {/* Модал истории бронирований */}
      <UserBookingHistoryModal 
        visible={historyModalVisible}
        onClose={() => {
          setHistoryModalVisible(false);
          setSelectedUserForHistory(null);
        }}
        user={selectedUserForHistory}
      />
    </div>
  );
};

export default UsersPage; 