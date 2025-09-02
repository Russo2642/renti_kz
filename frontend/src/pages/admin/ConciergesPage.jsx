import {
  CheckCircleOutlined, CloseCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  HomeOutlined,
  PlusOutlined,
  TeamOutlined,
  UserOutlined
} from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Avatar,
  Button, Card,
  Checkbox,
  Col,
  Descriptions,
  Drawer,
  Form,
  message,
  Modal,
  Popconfirm,
  Row,
  Select, Space,
  Statistic,
  Table,
  Tag,
  TimePicker,
  Tooltip,
  Typography
} from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useState } from 'react';
import { apartmentsAPI, conciergesAPI, usersAPI } from '../../lib/api.js';

const { Title, Text } = Typography;
const { Option } = Select;

const ConciergesPage = () => {
  const [filters, setFilters] = useState({});
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [selectedConcierge, setSelectedConcierge] = useState(null);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();
  const [assignForm] = Form.useForm();
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const queryClient = useQueryClient();

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получение консьержей
  const { data: conciergesData, isLoading } = useQuery({
    queryKey: ['concierges', filters],
    queryFn: () => conciergesAPI.getAll(filters)
  });

  // Получение квартир для назначения
  const { data: apartmentsData } = useQuery({
    queryKey: ['apartments-list'],
    queryFn: () => apartmentsAPI.adminGetAllApartments({ status: 'approved' })
  });

  // Получение пользователей для создания консьержа
  const { data: usersData } = useQuery({
    queryKey: ['users-list'],
    queryFn: () => usersAPI.adminGetAllUsers({ roles: ['user', 'concierge'] })
  });

  // Получение существующих консьержей для назначения на квартиры
  const { data: existingConciergesData } = useQuery({
    queryKey: ['existing-concierges'],
    queryFn: () => conciergesAPI.getAll({ is_active: true })
  });

  // Мутация для создания консьержа
  const createMutation = useMutation({
    mutationFn: conciergesAPI.create,
    onSuccess: () => {
      queryClient.invalidateQueries(['concierges']);
      queryClient.invalidateQueries(['existing-concierges']);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('Консьерж создан');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка создания консьержа');
    }
  });

  // Мутация для обновления консьержа
  const updateMutation = useMutation({
    mutationFn: ({ id, ...data }) => conciergesAPI.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['concierges']);
      setEditModalVisible(false);
      editForm.resetFields();
      message.success('Консьерж обновлен');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка обновления консьержа');
    }
  });

  // Мутация для удаления консьержа
  const deleteMutation = useMutation({
    mutationFn: conciergesAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries(['concierges']);
      message.success('Консьерж удален');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка удаления консьержа');
    }
  });

  // Мутация для назначения консьержа
  const assignMutation = useMutation({
    mutationFn: conciergesAPI.assign,
    onSuccess: () => {
      queryClient.invalidateQueries(['concierges']);
      queryClient.invalidateQueries(['existing-concierges']);
      setAssignModalVisible(false);
      assignForm.resetFields();
      message.success('Консьерж назначен');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка назначения консьержа');
    }
  });

  // Мутация для удаления консьержа из квартиры
  const removeFromApartmentMutation = useMutation({
    mutationFn: ({ conciergeId, apartmentId }) => conciergesAPI.removeFromApartment(conciergeId, apartmentId),
    onSuccess: () => {
      queryClient.invalidateQueries(['concierges']);
      message.success('Консьерж удален из квартиры');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка удаления консьержа из квартиры');
    }
  });

  const handleCreate = (values) => {
    // Преобразуем расписание
    const schedule = {};
    const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
    
    days.forEach(day => {
      if (values[`${day}_enabled`] && values[`${day}_start`] && values[`${day}_end`]) {
        schedule[day] = [{
          start: values[`${day}_start`].format('HH:mm'),
          end: values[`${day}_end`].format('HH:mm')
        }];
      } else {
        schedule[day] = null;
      }
    });

    const data = {
      user_id: parseInt(values.user_id),
      apartment_ids: [parseInt(values.apartment_id)], // Передаем массив для новой backend логики
      schedule: schedule
    };

    createMutation.mutate(data);
  };

  const handleUpdate = (values) => {
    const schedule = {};
    const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
    
    days.forEach(day => {
      if (values[`${day}_enabled`] && values[`${day}_start`] && values[`${day}_end`]) {
        schedule[day] = [{
          start: values[`${day}_start`].format('HH:mm'),
          end: values[`${day}_end`].format('HH:mm')
        }];
      } else {
        schedule[day] = null;
      }
    });

    updateMutation.mutate({
      id: selectedConcierge.id,
      is_active: values.is_active,
      schedule: schedule
    });
  };

  const handleAssign = (values) => {
    assignMutation.mutate({
      concierge_id: parseInt(values.concierge_id),
      apartment_id: parseInt(values.apartment_id)
    });
  };

  const handleDelete = (id) => {
    deleteMutation.mutate(id);
  };

  const handleRemoveFromApartment = (conciergeId, apartmentId) => {
    removeFromApartmentMutation.mutate({ conciergeId, apartmentId });
  };

  const getStatusColor = (isActive) => {
    return isActive ? 'green' : 'red';
  };

  const formatSchedule = (schedule) => {
    if (!schedule) return 'Не задано';
    
    const days = {
      monday: 'Пн',
      tuesday: 'Вт', 
      wednesday: 'Ср',
      thursday: 'Чт',
      friday: 'Пт',
      saturday: 'Сб',
      sunday: 'Вс'
    };

    const workingDays = [];
    Object.entries(days).forEach(([key, label]) => {
      if (schedule[key] && schedule[key].length > 0) {
        const time = schedule[key][0];
        workingDays.push(`${label}: ${time.start}-${time.end}`);
      }
    });

    return workingDays.length > 0 ? workingDays.join(', ') : 'Нет рабочих дней';
  };

  const columns = [
    {
      title: 'Консьерж',
      key: 'concierge',
      render: (_, record) => (
        <div className="flex items-center space-x-3">
          <Avatar 
            size="large" 
            icon={<UserOutlined />}
          />
          <div>
            <div className="font-medium">
              {record.user ? `${record.user.first_name} ${record.user.last_name}` : 'Неизвестный пользователь'}
            </div>
            <div className="text-gray-500 text-sm">{record.user?.phone}</div>
            <div className="text-gray-400 text-xs">ID: {record.user_id}</div>
          </div>
        </div>
      ),
    },
    {
      title: 'Email',
      dataIndex: ['user', 'email'],
      key: 'email',
      render: (email) => email || '—',
    },
    {
      title: 'Роль пользователя',
      dataIndex: ['user', 'role'],
      key: 'role',
      render: (role) => {
        const colors = {
          'concierge': 'blue',
          'user': 'default',
          'admin': 'red',
          'owner': 'green'
        };
        const roleTexts = {
          'concierge': 'Консьерж',
          'user': 'Пользователь',
          'admin': 'Администратор',
          'owner': 'Владелец'
        };
        return <Tag color={colors[role] || 'default'}>{roleTexts[role] || role || '—'}</Tag>;
      },
    },
    {
      title: 'Квартиры',
      key: 'apartments',
      render: (_, record) => {
        if (!record.apartments || record.apartments.length === 0) return '—';
        const firstApartment = record.apartments[0];
        return (
          <div>
            <div className="font-medium">
              {firstApartment.street} {firstApartment.building}
              {firstApartment.apartment_number ? `, кв. ${firstApartment.apartment_number}` : ''}
            </div>
            <div className="text-gray-500 text-sm">
              ID: {firstApartment.id}
              {record.apartments.length > 1 && ` (+${record.apartments.length - 1} еще)`}
            </div>
          </div>
        );
      },
    },
    {
      title: 'Расписание',
      dataIndex: 'schedule',
      key: 'schedule',
      render: (schedule) => (
        <Tooltip title={formatSchedule(schedule)}>
          <div className="max-w-32 truncate">
            {formatSchedule(schedule)}
          </div>
        </Tooltip>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive) => (
        <Tag color={getStatusColor(isActive)}>
          {isActive ? 'Активен' : 'Неактивен'}
        </Tag>
      ),
    },
    {
      title: 'Создан',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD.MM.YYYY HH:mm'),
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
                setSelectedConcierge(record);
                setDetailsVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedConcierge(record);
                const schedule = record.schedule || {};
                const formValues = { is_active: record.is_active };
                
                // Заполняем расписание
                const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
                days.forEach(day => {
                  if (schedule[day] && schedule[day].length > 0) {
                    const time = schedule[day][0];
                    formValues[`${day}_enabled`] = true;
                    formValues[`${day}_start`] = dayjs(time.start, 'HH:mm');
                    formValues[`${day}_end`] = dayjs(time.end, 'HH:mm');
                  }
                });

                editForm.setFieldsValue(formValues);
                setEditModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Удалить из квартиры">
            <Popconfirm
              title="Удалить консьержа из квартиры?"
              description={record.apartments?.length > 1 ? 
                `Будет удален из первой квартиры (${record.apartments[0]?.street} ${record.apartments[0]?.building})` :
                "Консьерж будет деактивирован, но запись останется"
              }
              onConfirm={() => {
                const apartmentId = record.apartments?.[0]?.id;
                if (apartmentId) {
                  handleRemoveFromApartment(record.id, apartmentId);
                }
              }}
              okText="Да"
              cancelText="Нет"
              disabled={!record.apartments || record.apartments.length === 0}
            >
              <Button
                type="text"
                icon={<HomeOutlined />}
                style={{ color: 'orange' }}
                disabled={!record.apartments || record.apartments.length === 0}
              />
            </Popconfirm>
          </Tooltip>
          <Tooltip title="Удалить полностью">
            <Popconfirm
              title="Удалить консьержа?"
              description="Это действие нельзя отменить. Консьерж будет полностью удален."
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

  // Подсчет статистики
  const concierges = conciergesData?.data || [];
  const stats = {
    total: concierges.length,
    active: concierges.filter(c => c.is_active).length,
    inactive: concierges.filter(c => !c.is_active).length,
    withApartments: concierges.filter(c => c.apartments && c.apartments.length > 0).length,
  };

  return (
    <div className={`${isMobile ? 'p-4' : 'p-6'}`}>
      <div className={`mb-6 ${isMobile ? 'space-y-4' : 'flex justify-between items-center'}`}>
        <div>
          <Title level={2} className={isMobile ? 'text-xl' : ''}>Управление консьержами</Title>
          <Text type="secondary">
            Управление консьержами и их назначение на квартиры
          </Text>
        </div>
        <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
            className={isMobile ? 'w-full' : ''}
          >
            Добавить консьержа
          </Button>
          <Button 
            icon={<TeamOutlined />}
            onClick={() => setAssignModalVisible(true)}
            className={isMobile ? 'w-full' : ''}
          >
            Назначить консьержа
          </Button>
        </Space>
      </div>

      {/* Статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Всего консьержей"
              value={stats.total}
              prefix={<TeamOutlined />}
              valueStyle={{ fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Активных"
              value={stats.active}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Неактивных"
              value={stats.inactive}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#cf1322', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="С квартирами"
              value={stats.withApartments}
              prefix={<HomeOutlined />}
              valueStyle={{ color: '#1890ff', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Фильтры */}
      <Card className="mb-6">
        <div className="flex flex-wrap gap-4 items-end">
          <div className="flex-1 min-w-[120px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
            <Select placeholder="Все" allowClear style={{ width: '100%' }} value={filters.is_active} onChange={(value) => setFilters({...filters, is_active: value})}>
              <Option value={true}>Активен</Option>
              <Option value={false}>Неактивен</Option>
            </Select>
          </div>
          <div className="flex-1 min-w-[200px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Квартира</label>
            <Select 
              placeholder="Выберите квартиру" 
              allowClear 
              style={{ width: '100%' }}
              showSearch
              optionFilterProp="children"
              value={filters.apartment_id}
              onChange={(value) => setFilters({...filters, apartment_id: value})}
            >
              {(apartmentsData?.data?.apartments || []).map(apartment => (
                <Option key={apartment.id} value={apartment.id}>
                  {apartment.street} {apartment.building}
                  {apartment.apartment_number ? `, кв. ${apartment.apartment_number}` : ''}
                </Option>
              ))}
            </Select>
          </div>
          <div className="flex-shrink-0">
            <Button onClick={() => {
              setFilters({});
              form.resetFields();
            }}>
              Сбросить
            </Button>
          </div>
        </div>
      </Card>

      {/* Таблица консьержей */}
      <Card>
        <Table
          columns={columns}
          dataSource={concierges}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 1200 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{
            total: conciergesData?.total,
            pageSize: 10,
            showSizeChanger: !isMobile,
            showQuickJumper: !isMobile,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} консьержей`,
            responsive: true,
            simple: isMobile,
          }}
        />
      </Card>

      {/* Модал создания консьержа */}
      <Modal
        title="Создать консьержа"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={isMobile ? '95%' : 800}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreate}
        >
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="user_id"
                label="Пользователь"
                rules={[{ required: true, message: 'Выберите пользователя' }]}
              >
                <Select 
                  placeholder="Выберите пользователя"
                  showSearch
                  optionFilterProp="children"
                  filterOption={(input, option) =>
                    option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                  }
                >
                  {usersData?.data?.users?.map(user => (
                    <Option key={user.id} value={user.id}>
                      {user.first_name} {user.last_name} ({user.phone})
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="apartment_id"
                label="Квартира"
                rules={[{ required: true, message: 'Выберите квартиру' }]}
              >
                <Select 
                  placeholder="Выберите квартиру"
                  showSearch
                  optionFilterProp="children"
                >
                  {(apartmentsData?.data?.apartments || []).map(apartment => (
                    <Option key={apartment.id} value={apartment.id}>
                      {apartment.street} {apartment.building}
                      {apartment.apartment_number ? `, кв. ${apartment.apartment_number}` : ''}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <div className="mb-4">
            <Title level={5}>Расписание работы</Title>
            <Text type="secondary">Настройте рабочие дни и часы консьержа</Text>
          </div>

          {['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'].map((day, index) => {
            const dayNames = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье'];
            return (
              <Row key={day} gutter={16} align="middle" className="mb-2">
                <Col xs={24} sm={6} md={4}>
                  <Form.Item name={`${day}_enabled`} valuePropName="checked" className="mb-0">
                    <Checkbox className={isMobile ? 'text-sm' : ''}>{dayNames[index]}</Checkbox>
                  </Form.Item>
                </Col>
                <Col xs={12} sm={9} md={6}>
                  <Form.Item 
                    name={`${day}_start`}
                    className="mb-0"
                    dependencies={[`${day}_enabled`]}
                  >
                    <TimePicker 
                      placeholder="Начало" 
                      format="HH:mm"
                      disabled={!Form.useWatch(`${day}_enabled`, form)}
                      className="w-full"
                    />
                  </Form.Item>
                </Col>
                <Col xs={12} sm={9} md={6}>
                  <Form.Item 
                    name={`${day}_end`}
                    className="mb-0"
                    dependencies={[`${day}_enabled`]}
                  >
                    <TimePicker 
                      placeholder="Конец" 
                      format="HH:mm"
                      disabled={!Form.useWatch(`${day}_enabled`, form)}
                      className="w-full"
                    />
                  </Form.Item>
                </Col>
              </Row>
            );
          })}

          <Form.Item className="mt-6">
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={createMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Создать
              </Button>
              <Button onClick={() => {
                setCreateModalVisible(false);
                form.resetFields();
              }} className={isMobile ? 'w-full' : ''}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Модал редактирования консьержа */}
      <Modal
        title="Редактировать консьержа"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          editForm.resetFields();
        }}
        footer={null}
        width={isMobile ? '95%' : 800}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdate}
        >
          <Form.Item
            name="is_active"
            label="Статус"
            valuePropName="checked"
          >
            <Checkbox>Активен</Checkbox>
          </Form.Item>

          <div className="mb-4">
            <Title level={5}>Расписание работы</Title>
          </div>

          {['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'].map((day, index) => {
            const dayNames = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье'];
            return (
              <Row key={day} gutter={16} align="middle" className="mb-2">
                <Col xs={24} sm={6} md={4}>
                  <Form.Item name={`${day}_enabled`} valuePropName="checked" className="mb-0">
                    <Checkbox className={isMobile ? 'text-sm' : ''}>{dayNames[index]}</Checkbox>
                  </Form.Item>
                </Col>
                <Col xs={12} sm={9} md={6}>
                  <Form.Item 
                    name={`${day}_start`}
                    className="mb-0"
                  >
                    <TimePicker 
                      placeholder="Начало" 
                      format="HH:mm"
                      className="w-full"
                    />
                  </Form.Item>
                </Col>
                <Col xs={12} sm={9} md={6}>
                  <Form.Item 
                    name={`${day}_end`}
                    className="mb-0"
                  >
                    <TimePicker 
                      placeholder="Конец" 
                      format="HH:mm"
                      className="w-full"
                    />
                  </Form.Item>
                </Col>
              </Row>
            );
          })}

          <Form.Item className="mt-6">
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={updateMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Обновить
              </Button>
              <Button onClick={() => {
                setEditModalVisible(false);
                editForm.resetFields();
              }} className={isMobile ? 'w-full' : ''}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Модал назначения консьержа */}
      <Modal
        title="Назначить консьержа"
        open={assignModalVisible}
        onCancel={() => {
          setAssignModalVisible(false);
          assignForm.resetFields();
        }}
        footer={null}
        width={isMobile ? '95%' : 600}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={assignForm}
          layout="vertical"
          onFinish={handleAssign}
        >
          <Form.Item
            name="concierge_id"
            label="Консьерж"
            rules={[{ required: true, message: 'Выберите консьержа' }]}
          >
            <Select 
              placeholder="Выберите консьержа"
              showSearch
              optionFilterProp="children"
            >
              {existingConciergesData?.data?.map(concierge => (
                <Option key={concierge.id} value={concierge.id}>
                  {concierge.user ? `${concierge.user.first_name} ${concierge.user.last_name} (${concierge.user.phone})` : `ID: ${concierge.id}`}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="apartment_id"
            label="Квартира"
            rules={[{ required: true, message: 'Выберите квартиру' }]}
          >
            <Select 
              placeholder="Выберите квартиру"
              showSearch
              optionFilterProp="children"
            >
              {(apartmentsData?.data?.apartments || []).map(apartment => (
                <Option key={apartment.id} value={apartment.id}>
                  {apartment.street} {apartment.building}
                  {apartment.apartment_number ? `, кв. ${apartment.apartment_number}` : ''}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item>
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={assignMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Назначить
              </Button>
              <Button onClick={() => {
                setAssignModalVisible(false);
                assignForm.resetFields();
              }} className={isMobile ? 'w-full' : ''}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Drawer с деталями консьержа */}
      <Drawer
        title="Детали консьержа"
        placement={isMobile ? 'bottom' : 'right'}
        size={isMobile ? 'default' : 'large'}
        onClose={() => setDetailsVisible(false)}
        open={detailsVisible}
        width={isMobile ? '100%' : 720}
        height={isMobile ? '90%' : undefined}
      >
        {selectedConcierge && (
          <div>
            <Descriptions column={isMobile ? 1 : 2} bordered size={isMobile ? 'small' : 'default'}>
              <Descriptions.Item label="ID">{selectedConcierge.id}</Descriptions.Item>
              <Descriptions.Item label="Пользователь">
                {selectedConcierge.user ? 
                  `${selectedConcierge.user.first_name} ${selectedConcierge.user.last_name}` : 
                  'Неизвестный пользователь'
                }
              </Descriptions.Item>
              <Descriptions.Item label="Телефон">{selectedConcierge.user?.phone || '—'}</Descriptions.Item>
              <Descriptions.Item label="Email">{selectedConcierge.user?.email || '—'}</Descriptions.Item>
              <Descriptions.Item label="Роль пользователя">
                <Tag color={selectedConcierge.user?.role === 'concierge' ? 'blue' : 'default'}>
                  {selectedConcierge.user?.role || '—'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Квартира" span={isMobile ? 1 : 2}>
                {selectedConcierge.apartment ? 
                  `${selectedConcierge.apartment.street} ${selectedConcierge.apartment.building}` +
                  (selectedConcierge.apartment.apartment_number ? `, кв. ${selectedConcierge.apartment.apartment_number}` : '') :
                  '—'
                }
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                <Tag color={getStatusColor(selectedConcierge.is_active)}>
                  {selectedConcierge.is_active ? 'Активен' : 'Неактивен'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Расписание" span={isMobile ? 1 : 2}>
                <div style={{ whiteSpace: 'pre-line' }}>
                  {formatSchedule(selectedConcierge.schedule)}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Создан" span={isMobile ? 1 : 2}>
                <div className="font-mono text-sm">
                  {dayjs(selectedConcierge.created_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedConcierge.created_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Обновлен" span={isMobile ? 1 : 2}>
                <div className="font-mono text-sm">
                  {dayjs(selectedConcierge.updated_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedConcierge.updated_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default ConciergesPage; 