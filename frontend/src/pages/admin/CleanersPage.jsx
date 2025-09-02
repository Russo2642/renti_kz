import {
  CheckCircleOutlined, CloseCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  HomeOutlined,
  PlusOutlined,
  TeamOutlined,
  UserOutlined,
  ToolOutlined
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
import { apartmentsAPI, cleanersAPI, usersAPI } from '../../lib/api.js';

const { Title, Text } = Typography;
const { Option } = Select;

const CleanersPage = () => {
  const [filters, setFilters] = useState({});
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [selectedCleaner, setSelectedCleaner] = useState(null);
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

  // Получение уборщиц
  const { data: cleanersData, isLoading } = useQuery({
    queryKey: ['cleaners', filters],
    queryFn: () => cleanersAPI.getAll(filters)
  });

  // Получение квартир для назначения
  const { data: apartmentsData, isLoading: apartmentsLoading, error: apartmentsError } = useQuery({
    queryKey: ['apartments-list'],
    queryFn: () => apartmentsAPI.adminGetAllApartments({ status: 'approved', page: 1, page_size: 100 }),
    onSuccess: (data) => {
    },
    onError: (error) => {
      console.error('Error loading apartments:', error);
    }
  });

  // Получение пользователей для создания уборщицы
  const { data: usersData, isLoading: usersLoading, error: usersError } = useQuery({
    queryKey: ['users-list'],
    queryFn: () => usersAPI.adminGetAllUsers({ page: 1, page_size: 100 }),
    onSuccess: (data) => {
    },
    onError: (error) => {
      console.error('Error loading users:', error);
    }
  });

  // Получение существующих уборщиц для назначения на квартиры
  const { data: existingCleanersData } = useQuery({
    queryKey: ['existing-cleaners'],
    queryFn: () => cleanersAPI.getAll({ is_active: true })
  });

  // Мутация для создания уборщицы
  const createMutation = useMutation({
    mutationFn: cleanersAPI.create,
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaners']);
      queryClient.invalidateQueries(['existing-cleaners']);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('Уборщица создана');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка создания уборщицы');
    }
  });

  // Мутация для обновления уборщицы
  const updateMutation = useMutation({
    mutationFn: ({ id, ...data }) => cleanersAPI.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaners']);
      setEditModalVisible(false);
      editForm.resetFields();
      message.success('Уборщица обновлена');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка обновления уборщицы');
    }
  });

  // Мутация для удаления уборщицы
  const deleteMutation = useMutation({
    mutationFn: cleanersAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaners']);
      queryClient.invalidateQueries(['existing-cleaners']);
      message.success('Уборщица удалена');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка удаления уборщицы');
    }
  });

  // Мутация для назначения уборщицы на квартиру
  const assignMutation = useMutation({
    mutationFn: cleanersAPI.assignToApartment,
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaners']);
      setAssignModalVisible(false);
      assignForm.resetFields();
      message.success('Уборщица назначена на квартиру');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка назначения уборщицы');
    }
  });

  // Мутация для удаления назначения уборщицы с квартиры
  const removeAssignmentMutation = useMutation({
    mutationFn: cleanersAPI.removeFromApartment,
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaners']);
      message.success('Назначение уборщицы удалено');
    },
    onError: (error) => {
      message.error(error.response?.data?.message || 'Ошибка удаления назначения');
    }
  });

  const handleCreate = async (values) => {
    try {
      const { apartment_ids, ...formData } = values;
      
      // Преобразуем расписание в правильный формат
      const cleanerData = { ...formData };
      
      // Если есть данные расписания, преобразуем их
      const scheduleData = {};
      let hasSchedule = false;
      
      const daysOfWeek = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
      daysOfWeek.forEach(day => {
        if (formData[day] && formData[day].start_time && formData[day].end_time) {
          scheduleData[day] = [{
            start: formData[day].start_time.format('HH:mm'),
            end: formData[day].end_time.format('HH:mm'),
          }];
          hasSchedule = true;
        } else {
          scheduleData[day] = [];
        }
        // Удаляем данные дня из основных данных
        delete cleanerData[day];
      });
      
      // Добавляем расписание, если оно есть
      if (hasSchedule) {
        cleanerData.schedule = scheduleData;
      }
      
      // Сначала создаем уборщицу
      const result = await createMutation.mutateAsync(cleanerData);
      
      // Если указаны квартиры, назначаем их
      if (apartment_ids && apartment_ids.length > 0) {
        try {
          await cleanersAPI.updateApartments(result.data.id, apartment_ids);
        } catch (apartmentError) {
          console.error('Error assigning apartments:', apartmentError);
          message.warning('Уборщица создана, но возникла ошибка при назначении квартир');
        }
      }

      queryClient.invalidateQueries(['cleaners']);
      queryClient.invalidateQueries(['existing-cleaners']);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('Уборщица создана');
    } catch (error) {
      message.error(error.response?.data?.message || 'Ошибка создания уборщицы');
    }
  };

  const handleEdit = async (values) => {
    try {
      const { apartment_ids, ...formData } = values;
      
      // Обновляем основную информацию (без расписания)
      const cleanerData = { ...formData };
      
      // Отделяем расписание от основных данных
      const daysOfWeek = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
      const scheduleData = {};
      let hasScheduleChanges = false;
      
      daysOfWeek.forEach(day => {
        // Удаляем данные дня из основных данных
        delete cleanerData[day];
        
        // Проверяем, есть ли изменения в расписании для этого дня
        if (formData[day]) {
          if (formData[day].start_time && formData[day].end_time) {
            // День с рабочими часами
            scheduleData[day] = [{
              start: formData[day].start_time.format('HH:mm'),
              end: formData[day].end_time.format('HH:mm'),
            }];
            hasScheduleChanges = true;
          } else if (formData[day].start_time === null && formData[day].end_time === null) {
            // День явно очищен (выходной)
            scheduleData[day] = [];
            hasScheduleChanges = true;
          }
          // Если только одно поле заполнено, игнорируем этот день
        }
      });
      
      // Сначала обновляем основную информацию
      await updateMutation.mutateAsync({
        id: selectedCleaner.id,
        ...cleanerData
      });
      
      // Затем обновляем расписание через PATCH, если есть изменения
      if (hasScheduleChanges) {
        await cleanersAPI.updateSchedulePatch(selectedCleaner.id, scheduleData);
      }

      queryClient.invalidateQueries(['cleaners']);
      setEditModalVisible(false);
      editForm.resetFields();
      message.success('Уборщица успешно обновлена');
    } catch (error) {
      message.error(error.response?.data?.message || 'Ошибка обновления уборщицы');
    }
  };

  const handleDelete = (id) => {
    deleteMutation.mutate(id);
  };

  const handleAssign = (values) => {
    assignMutation.mutate(values);
  };

  const handleRemoveAssignment = (cleanerId, apartmentId) => {
    removeAssignmentMutation.mutate({
      cleaner_id: cleanerId,
      apartment_id: apartmentId
    });
  };

  const showDetails = (cleaner) => {
    setSelectedCleaner(cleaner);
    setDetailsVisible(true);
  };

  const showEditModal = (cleaner) => {
    setSelectedCleaner(cleaner);
    
    // Подготавливаем данные для формы
    // Пробуем получить квартиры из разных полей
    const apartmentIds = cleaner.apartments ? cleaner.apartments.map(apt => apt.id) : 
                        cleaner.assigned_apartments ? cleaner.assigned_apartments.map(apt => apt.id) : 
                        cleaner.apartment_ids ? cleaner.apartment_ids : [];
    
    const formData = {
      user_id: cleaner.user_id,
      is_active: cleaner.is_active,
      apartment_ids: apartmentIds,
    };

    // Добавляем расписание
    if (cleaner.schedule) {
      Object.entries(cleaner.schedule).forEach(([day, hoursArray]) => {
        if (hoursArray && Array.isArray(hoursArray) && hoursArray.length > 0) {
          const firstHours = hoursArray[0];
          if (firstHours && firstHours.start && firstHours.end) {
            formData[day] = {
              start_time: dayjs(firstHours.start, 'HH:mm'),
              end_time: dayjs(firstHours.end, 'HH:mm'),
            };
          }
        }
      });
    }

    editForm.setFieldsValue(formData);
    setEditModalVisible(true);
  };

  // Колонки таблицы
  const columns = [
    {
      title: 'Уборщица',
      key: 'user',
      render: (_, record) => (
        <div className="flex items-center space-x-3">
          <Avatar 
            size="large" 
            icon={<UserOutlined />}
            className="bg-green-500"
          >
            {record.user?.first_name?.[0]}{record.user?.last_name?.[0]}
          </Avatar>
          <div>
            <div className="font-medium text-gray-900">
              {record.user?.first_name} {record.user?.last_name}
            </div>
            <div className="text-sm text-gray-500">
              {record.user?.phone}
            </div>
            <div className="text-xs text-gray-400">
              ID: {record.id}
            </div>
          </div>
        </div>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive) => (
        <Tag 
          icon={isActive ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
          color={isActive ? 'success' : 'error'}
        >
          {isActive ? 'Активна' : 'Неактивна'}
        </Tag>
      ),
    },
    {
      title: 'Квартиры',
      key: 'apartments',
      render: (_, record) => (
        <div>
          {record.apartments && record.apartments.length > 0 ? (
            <div className="space-y-1">
              {Array.isArray(record.apartments) && record.apartments.map((apt, index) => (
                <div key={index} className="text-sm">
                  <span className="text-blue-600">
                    {apt.street}, д. {apt.building}
                  </span>
                  {apt.apartment_number && (
                    <span className="text-gray-500 ml-1">
                      кв. {apt.apartment_number}
                    </span>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <Text type="secondary">Без назначений</Text>
          )}
        </div>
      ),
    },
    {
      title: 'Расписание',
      key: 'schedule',
      render: (_, record) => (
        <div className="text-xs">
          {record.schedule ? (
            <div className="space-y-1">
              {Object.entries(record.schedule).map(([day, hoursArray]) => {
                if (!hoursArray || !Array.isArray(hoursArray) || hoursArray.length === 0) return null;
                const firstHours = hoursArray[0];
                if (!firstHours || !firstHours.start || !firstHours.end) return null;
                
                const dayNames = {
                  monday: 'Пн',
                  tuesday: 'Вт',
                  wednesday: 'Ср',
                  thursday: 'Чт',
                  friday: 'Пт',
                  saturday: 'Сб',
                  sunday: 'Вс'
                };
                return (
                  <div key={day}>
                    <span className="font-medium">{dayNames[day]}:</span>
                    <span className="ml-1">
                      {firstHours.start} - {firstHours.end}
                    </span>
                  </div>
                );
              })}
            </div>
          ) : (
            <Text type="secondary">Не задано</Text>
          )}
        </div>
      ),
    },
    {
      title: 'Статистика',
      key: 'stats',
      render: (_, record) => (
        <div className="text-xs space-y-1">
          <div className="flex items-center space-x-1">
            <ToolOutlined className="text-blue-500" />
            <span>{record.stats?.total_apartments || 0} квартир</span>
          </div>
          <div className="flex items-center space-x-1">
            <CheckCircleOutlined className="text-green-500" />
            <span>{record.stats?.cleaned_today || 0} сегодня</span>
          </div>
          <div className="flex items-center space-x-1">
            <CloseCircleOutlined className="text-orange-500" />
            <span>{record.stats?.apartments_needing_cleaning || 0} нужна уборка</span>
          </div>
        </div>
      ),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="Просмотр">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => showDetails(record)}
              size="small"
            />
          </Tooltip>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => showEditModal(record)}
              size="small"
            />
          </Tooltip>
          <Popconfirm
            title="Удалить уборщицу?"
            description="Это действие нельзя отменить"
            onConfirm={() => handleDelete(record.id)}
            okText="Да"
            cancelText="Нет"
          >
            <Tooltip title="Удалить">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                size="small"
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // Подсчет статистики
  const cleaners = cleanersData?.data || [];
  const stats = {
    total: cleaners.length,
    active: cleaners.filter(c => c.is_active).length,
    inactive: cleaners.filter(c => !c.is_active).length,
    withApartments: cleaners.filter(c => c.apartments && c.apartments.length > 0).length,
  };

  return (
    <div className={`${isMobile ? 'p-4' : 'p-6'}`}>
      <div className={`mb-6 ${isMobile ? 'space-y-4' : 'flex justify-between items-center'}`}>
        <div>
          <Title level={2} className={isMobile ? 'text-xl' : ''}>Управление уборщицами</Title>
          <Text type="secondary">
            Управление уборщицами и их назначение на квартиры
          </Text>
        </div>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
          className={isMobile ? 'w-full' : ''}
        >
          Добавить уборщицу
        </Button>
      </div>

      {/* Статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="Всего уборщиц"
              value={stats.total}
              prefix={<ToolOutlined className="text-blue-500" />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="Активных"
              value={stats.active}
              prefix={<CheckCircleOutlined className="text-green-500" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="Неактивных"
              value={stats.inactive}
              prefix={<CloseCircleOutlined className="text-red-500" />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="С назначениями"
              value={stats.withApartments}
              prefix={<HomeOutlined className="text-orange-500" />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Дополнительные действия */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={24} sm={12}>
          <Card>
            <div className="flex items-center justify-between">
              <div>
                <Title level={4} className="!mb-1">Квартиры для уборки</Title>
                <Text type="secondary">Просмотр всех квартир, нуждающихся в уборке</Text>
              </div>
              <Button 
                type="primary" 
                icon={<ToolOutlined />}
                onClick={() => {
                  // TODO: открыть модальное окно со списком квартир для уборки
                  message.info('Функция в разработке');
                }}
              >
                Просмотреть
              </Button>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12}>
          <Card>
            <div className="flex items-center justify-between">
              <div>
                <Title level={4} className="!mb-1">Назначить уборщицу</Title>
                <Text type="secondary">Назначить уборщицу на квартиру</Text>
              </div>
              <Button 
                type="default" 
                icon={<HomeOutlined />}
                onClick={() => setAssignModalVisible(true)}
              >
                Назначить
              </Button>
            </div>
          </Card>
        </Col>
      </Row>

      {/* Таблица уборщиц */}
      <Card>
        <Table
          columns={columns}
          dataSource={cleaners}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 1200 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{
            total: cleanersData?.total,
            pageSize: 10,
            showSizeChanger: !isMobile,
            showQuickJumper: !isMobile,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} уборщиц`,
            responsive: true,
            simple: isMobile,
          }}
        />
      </Card>

      {/* Модал создания уборщицы */}
      <Modal
        title="Создать уборщицу"
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
          className="mt-4"
        >
          <Form.Item
            name="user_id"
            label="Выберите пользователя"
            rules={[{ required: true, message: 'Выберите пользователя' }]}
          >
            <Select
              placeholder="Выберите пользователя"
              showSearch
              optionFilterProp="children"
              loading={usersLoading}
              notFoundContent={
                usersLoading ? 'Загрузка...' : 
                usersError ? `Ошибка: ${usersError.message}` : 
                usersData?.data?.users?.length === 0 ? 'Нет пользователей' :
                'Нет данных'
              }
            >
              {Array.isArray(usersData?.data?.users) && usersData.data.users.map(user => (
                <Option key={user.id} value={user.id}>
                  {user.first_name} {user.last_name} ({user.phone})
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="is_active"
            label="Статус"
            valuePropName="checked"
            initialValue={true}
          >
            <Checkbox>Активна</Checkbox>
          </Form.Item>

          <Form.Item
            name="apartment_ids"
            label="Назначить на квартиры"
          >
            <Select
              mode="multiple"
              placeholder="Выберите квартиры (необязательно)"
              showSearch
              optionFilterProp="children"
              loading={apartmentsLoading}
              notFoundContent={
                apartmentsLoading ? 'Загрузка...' : 
                apartmentsError ? `Ошибка: ${apartmentsError.message}` : 
                apartmentsData?.data?.apartments?.length === 0 ? 'Нет одобренных квартир' :
                'Нет данных'
              }
            >
              {Array.isArray(apartmentsData?.data?.apartments) && apartmentsData.data.apartments.map(apartment => (
                <Option key={apartment.id} value={apartment.id}>
                  {apartment.street}, д. {apartment.building}
                  {apartment.apartment_number && `, кв. ${apartment.apartment_number}`}
                </Option>
              ))}
            </Select>
          </Form.Item>

          {/* Расписание */}
          <div className="mb-4">
            <Text strong>Расписание работы</Text>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
              {[
                { key: 'monday', label: 'Понедельник' },
                { key: 'tuesday', label: 'Вторник' },
                { key: 'wednesday', label: 'Среда' },
                { key: 'thursday', label: 'Четверг' },
                { key: 'friday', label: 'Пятница' },
                { key: 'saturday', label: 'Суббота' },
                { key: 'sunday', label: 'Воскресенье' }
              ].map(day => (
                <div key={day.key} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Text>{day.label}</Text>
                    <Form.Item
                      noStyle
                      shouldUpdate={(prevValues, currentValues) => 
                        prevValues[day.key]?.start_time !== currentValues[day.key]?.start_time ||
                        prevValues[day.key]?.end_time !== currentValues[day.key]?.end_time
                      }
                    >
                      {({ getFieldValue }) => {
                        const startTime = getFieldValue([day.key, 'start_time']);
                        const endTime = getFieldValue([day.key, 'end_time']);
                        const hasTime = startTime && endTime;
                        return (
                          <Tag color={hasTime ? 'green' : 'red'} size="small">
                            {hasTime ? 'Работаю' : 'Выходной'}
                          </Tag>
                        );
                      }}
                    </Form.Item>
                  </div>
                  <div className="flex space-x-2">
                    <Form.Item
                      name={[day.key, 'start_time']}
                      className="mb-0 flex-1"
                    >
                      <TimePicker
                        placeholder="Начало"
                        format="HH:mm"
                        size="small"
                        allowClear
                      />
                    </Form.Item>
                    <Form.Item
                      name={[day.key, 'end_time']}
                      className="mb-0 flex-1"
                    >
                      <TimePicker
                        placeholder="Конец"
                        format="HH:mm"
                        size="small"
                        allowClear
                      />
                    </Form.Item>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end space-x-2 pt-4 border-t">
            <Button onClick={() => setCreateModalVisible(false)}>
              Отмена
            </Button>
            <Button 
              type="primary" 
              htmlType="submit"
              loading={createMutation.isPending}
            >
              Создать
            </Button>
          </div>
        </Form>
      </Modal>

      {/* Модал редактирования уборщицы */}
      <Modal
        title="Редактировать уборщицу"
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
          onFinish={handleEdit}
          className="mt-4"
        >
          <Form.Item
            name="user_id"
            label="Пользователь"
            rules={[{ required: true, message: 'Выберите пользователя' }]}
          >
            <Select
              placeholder="Выберите пользователя"
              showSearch
              optionFilterProp="children"
              loading={usersLoading}
              notFoundContent={
                usersLoading ? 'Загрузка...' : 
                usersError ? `Ошибка: ${usersError.message}` : 
                usersData?.data?.users?.length === 0 ? 'Нет пользователей' :
                'Нет данных'
              }
            >
              {Array.isArray(usersData?.data?.users) && usersData.data.users.map(user => (
                <Option key={user.id} value={user.id}>
                  {user.first_name} {user.last_name} ({user.phone})
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="is_active"
            label="Статус"
            valuePropName="checked"
          >
            <Checkbox>Активна</Checkbox>
          </Form.Item>



          {/* Расписание */}
          <div className="mb-4">
            <Text strong>Расписание работы</Text>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
              {[
                { key: 'monday', label: 'Понедельник' },
                { key: 'tuesday', label: 'Вторник' },
                { key: 'wednesday', label: 'Среда' },
                { key: 'thursday', label: 'Четверг' },
                { key: 'friday', label: 'Пятница' },
                { key: 'saturday', label: 'Суббота' },
                { key: 'sunday', label: 'Воскресенье' }
              ].map(day => (
                <div key={day.key} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Text>{day.label}</Text>
                    <Form.Item
                      noStyle
                      shouldUpdate={(prevValues, currentValues) => 
                        prevValues[day.key]?.start_time !== currentValues[day.key]?.start_time ||
                        prevValues[day.key]?.end_time !== currentValues[day.key]?.end_time
                      }
                    >
                      {({ getFieldValue }) => {
                        const startTime = getFieldValue([day.key, 'start_time']);
                        const endTime = getFieldValue([day.key, 'end_time']);
                        const hasTime = startTime && endTime;
                        return (
                          <Tag color={hasTime ? 'green' : 'red'} size="small">
                            {hasTime ? 'Работаю' : 'Выходной'}
                          </Tag>
                        );
                      }}
                    </Form.Item>
                  </div>
                  <div className="flex space-x-2">
                    <Form.Item
                      name={[day.key, 'start_time']}
                      className="mb-0 flex-1"
                    >
                      <TimePicker
                        placeholder="Начало"
                        format="HH:mm"
                        size="small"
                        allowClear
                      />
                    </Form.Item>
                    <Form.Item
                      name={[day.key, 'end_time']}
                      className="mb-0 flex-1"
                    >
                      <TimePicker
                        placeholder="Конец"
                        format="HH:mm"
                        size="small"
                        allowClear
                      />
                    </Form.Item>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end space-x-2 pt-4 border-t">
            <Button onClick={() => setEditModalVisible(false)}>
              Отмена
            </Button>
            <Button 
              type="primary" 
              htmlType="submit"
              loading={updateMutation.isPending}
            >
              Сохранить
            </Button>
          </div>
        </Form>
      </Modal>

      {/* Модал назначения уборщицы на квартиру */}
      <Modal
        title="Назначить уборщицу на квартиру"
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
          className="mt-4"
        >
          <Form.Item
            name="cleaner_id"
            label="Выберите уборщицу"
            rules={[{ required: true, message: 'Выберите уборщицу' }]}
          >
            <Select
              placeholder="Выберите уборщицу"
              showSearch
              optionFilterProp="children"
              loading={!existingCleanersData}
            >
              {Array.isArray(existingCleanersData?.data) && existingCleanersData.data.map(cleaner => (
                <Option key={cleaner.id} value={cleaner.id}>
                  {cleaner.user?.first_name} {cleaner.user?.last_name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="apartment_id"
            label="Выберите квартиру"
            rules={[{ required: true, message: 'Выберите квартиру' }]}
          >
            <Select
              placeholder="Выберите квартиру"
              showSearch
              optionFilterProp="children"
              loading={apartmentsLoading}
              notFoundContent={
                apartmentsLoading ? 'Загрузка...' : 
                apartmentsError ? `Ошибка: ${apartmentsError.message}` : 
                apartmentsData?.data?.apartments?.length === 0 ? 'Нет одобренных квартир' :
                'Нет данных'
              }
            >
              {Array.isArray(apartmentsData?.data?.apartments) && apartmentsData.data.apartments.map(apartment => (
                <Option key={apartment.id} value={apartment.id}>
                  {apartment.street}, д. {apartment.building}
                  {apartment.apartment_number && `, кв. ${apartment.apartment_number}`}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <div className="flex justify-end space-x-2 pt-4 border-t">
            <Button onClick={() => setAssignModalVisible(false)}>
              Отмена
            </Button>
            <Button 
              type="primary" 
              htmlType="submit"
              loading={assignMutation.isPending}
            >
              Назначить
            </Button>
          </div>
        </Form>
      </Modal>

      {/* Drawer с деталями уборщицы */}
      <Drawer
        title="Информация об уборщице"
        placement="right"
        onClose={() => setDetailsVisible(false)}
        open={detailsVisible}
        width={isMobile ? '100%' : 600}
      >
        {selectedCleaner && (
          <div className="space-y-6">
            {/* Основная информация */}
            <Card title="Основная информация">
              <Descriptions column={1} size="small">
                <Descriptions.Item label="ID">
                  {selectedCleaner.id}
                </Descriptions.Item>
                <Descriptions.Item label="Имя">
                  {selectedCleaner.user?.first_name} {selectedCleaner.user?.last_name}
                </Descriptions.Item>
                <Descriptions.Item label="Телефон">
                  {selectedCleaner.user?.phone}
                </Descriptions.Item>
                <Descriptions.Item label="Email">
                  {selectedCleaner.user?.email || 'Не указан'}
                </Descriptions.Item>
                <Descriptions.Item label="Статус">
                  <Tag 
                    icon={selectedCleaner.is_active ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
                    color={selectedCleaner.is_active ? 'success' : 'error'}
                  >
                    {selectedCleaner.is_active ? 'Активна' : 'Неактивна'}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Создана">
                  {dayjs(selectedCleaner.created_at).format('DD.MM.YYYY HH:mm')}
                </Descriptions.Item>
                <Descriptions.Item label="Обновлена">
                  {dayjs(selectedCleaner.updated_at).format('DD.MM.YYYY HH:mm')}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {/* Назначенные квартиры */}
            <Card 
              title="Назначенные квартиры"
              extra={
                <Button 
                  type="primary"
                  size="small"
                  icon={<PlusOutlined />}
                  onClick={() => {
                    assignForm.setFieldsValue({
                      cleaner_id: selectedCleaner.id
                    });
                    setAssignModalVisible(true);
                  }}
                >
                  Назначить квартиру
                </Button>
              }
            >
              {selectedCleaner.apartments && selectedCleaner.apartments.length > 0 ? (
                <div className="space-y-3">
                  {Array.isArray(selectedCleaner.apartments) && selectedCleaner.apartments.map((apartment, index) => (
                    <div key={index} className="flex items-center justify-between p-3 border rounded">
                      <div>
                        <div className="font-medium">
                          {apartment.street}, д. {apartment.building}
                        </div>
                        {apartment.apartment_number && (
                          <div className="text-sm text-gray-500">
                            Квартира {apartment.apartment_number}
                          </div>
                        )}
                      </div>
                      <Popconfirm
                        title="Убрать назначение?"
                        description="Уборщица больше не будет назначена на эту квартиру"
                        onConfirm={() => handleRemoveAssignment(selectedCleaner.id, apartment.id)}
                        okText="Да"
                        cancelText="Нет"
                      >
                        <Button 
                          type="text" 
                          danger 
                          icon={<DeleteOutlined />}
                          size="small"
                        >
                          Убрать
                        </Button>
                      </Popconfirm>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-4">
                  <Text type="secondary">Квартиры не назначены</Text>
                  <br />
                  <Button 
                    type="dashed"
                    icon={<PlusOutlined />}
                    onClick={() => {
                      assignForm.setFieldsValue({
                        cleaner_id: selectedCleaner.id
                      });
                      setAssignModalVisible(true);
                    }}
                    className="mt-2"
                  >
                    Назначить первую квартиру
                  </Button>
                </div>
              )}
            </Card>

            {/* Расписание */}
            <Card title="Расписание работы">
              {selectedCleaner.schedule ? (
                <div className="space-y-2">
                  {selectedCleaner.schedule && Object.entries(selectedCleaner.schedule).map(([day, hoursArray]) => {
                    if (!hoursArray || !Array.isArray(hoursArray) || hoursArray.length === 0) return null;
                    const firstHours = hoursArray[0];
                    if (!firstHours || !firstHours.start || !firstHours.end) return null;
                    
                    const dayNames = {
                      monday: 'Понедельник',
                      tuesday: 'Вторник',
                      wednesday: 'Среда',
                      thursday: 'Четверг',
                      friday: 'Пятница',
                      saturday: 'Суббота',
                      sunday: 'Воскресенье'
                    };
                    return (
                      <div key={day} className="flex justify-between p-2 border-b">
                        <span className="font-medium">{dayNames[day]}</span>
                        <span>{firstHours.start} - {firstHours.end}</span>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <Text type="secondary">Расписание не задано</Text>
              )}
            </Card>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default CleanersPage;
