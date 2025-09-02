import {
  DeleteOutlined,
  EditOutlined,
  EnvironmentOutlined,
  EyeOutlined,
  HistoryOutlined,
  PlusOutlined,
  SettingOutlined,
  TagOutlined,
  UploadOutlined
} from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Badge,
  Button, Card,
  Checkbox,
  Col,
  Descriptions,
  Drawer,
  Form,
  Image,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Row,
  Select, Space,
  Statistic,
  Table,
  Tag,
  Tooltip,
  Typography,
  Upload
} from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useState } from 'react';
import LocationFilter from '../../components/LocationFilter.jsx';
import { apartmentsAPI, contractsAPI, locationsAPI, dictionariesAPI, apartmentTypesAPI } from '../../lib/api.js';
import ApartmentBookingHistoryModal from '../../components/ApartmentBookingHistoryModal.jsx';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

const ApartmentsPage = () => {
  const [filters, setFilters] = useState({});
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [selectedApartment, setSelectedApartment] = useState(null);
  const [statusModalVisible, setStatusModalVisible] = useState(false);
  const [historyModalVisible, setHistoryModalVisible] = useState(false);
  const [selectedApartmentForHistory, setSelectedApartmentForHistory] = useState(null);
  const [apartmentTypeModalVisible, setApartmentTypeModalVisible] = useState(false);
  const [countersModalVisible, setCountersModalVisible] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [form] = Form.useForm();
  const [statusForm] = Form.useForm();
  const [apartmentTypeForm] = Form.useForm();
  const [countersForm] = Form.useForm();
  const queryClient = useQueryClient();

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получение квартир (админская версия)
  const { data: apartmentsData, isLoading } = useQuery({
    queryKey: ['admin-apartments', filters, currentPage, pageSize],
    queryFn: () => {
      const params = {
        page: currentPage,
        page_size: pageSize,
        ...filters
      };
      return apartmentsAPI.adminGetAllApartments(params);
    }
  });

  // Получение полной статистики дашборда
  const { data: dashboardData } = useQuery({
    queryKey: ['admin-dashboard-stats'],
    queryFn: apartmentsAPI.adminGetFullDashboardStats,
  });

  // Получение детальной статистики квартир
  const { data: apartmentsStatistics, isLoading: isLoadingStatistics } = useQuery({
    queryKey: ['admin-apartments-statistics'],
    queryFn: apartmentsAPI.adminGetApartmentsStatistics,
    staleTime: 5 * 60 * 1000, // 5 минут
  });

  // Получение словарей для формы редактирования
  const { data: citiesData } = useQuery({
    queryKey: ['cities'],
    queryFn: () => locationsAPI.getCities()
  });

  const { data: conditionsData } = useQuery({
    queryKey: ['conditions'],
    queryFn: () => dictionariesAPI.getConditions()
  });

  const { data: amenitiesData } = useQuery({
    queryKey: ['amenities'],
    queryFn: () => dictionariesAPI.getAmenities()
  });

  const { data: houseRulesData } = useQuery({
    queryKey: ['house-rules'],
    queryFn: () => dictionariesAPI.getHouseRules()
  });

  // Получение типов квартир для админки
  const { data: apartmentTypes } = useQuery({
    queryKey: ['apartmentTypes'],
    queryFn: apartmentTypesAPI.getAll
  });

  // Получение районов для выбранного города
  const [selectedCityId, setSelectedCityId] = useState(null);
  const { data: districtsData } = useQuery({
    queryKey: ['districts', selectedCityId],
    queryFn: () => selectedCityId ? locationsAPI.getDistrictsByCity(selectedCityId) : Promise.resolve([]),
    enabled: !!selectedCityId
  });

  // Получение микрорайонов для выбранного района
  const [selectedDistrictId, setSelectedDistrictId] = useState(null);
  const { data: microdistrictsData } = useQuery({
    queryKey: ['microdistricts', selectedDistrictId],
    queryFn: () => selectedDistrictId ? locationsAPI.getMicrodistrictsByDistrict(selectedDistrictId) : Promise.resolve([]),
    enabled: !!selectedDistrictId
  });

  // Состояния для условного отображения полей
  const [showHourlyPrice, setShowHourlyPrice] = useState(false);
  const [showDailyPrice, setShowDailyPrice] = useState(false);
  const [existingPhotos, setExistingPhotos] = useState([]);





  // Мутация для обновления статуса
  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status, comment }) => apartmentsAPI.updateStatus(id, status, comment),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-apartments']);
      queryClient.invalidateQueries(['admin-dashboard-stats']);
      queryClient.invalidateQueries(['admin-apartments-statistics']);
      setStatusModalVisible(false);
      statusForm.resetFields();
      message.success('Статус квартиры обновлен');
    }
  });

  // Мутация для обновления типа квартиры
  const updateApartmentTypeMutation = useMutation({
    mutationFn: ({ id, apartmentTypeId }) => apartmentsAPI.updateApartmentType(id, apartmentTypeId),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-apartments']);
      queryClient.invalidateQueries(['admin-dashboard-stats']);
      queryClient.invalidateQueries(['admin-apartments-statistics']);
      setApartmentTypeModalVisible(false);
      apartmentTypeForm.resetFields();
      message.success('Тип квартиры обновлен');
    },
    onError: () => {
      message.error('Ошибка при обновлении типа квартиры');
    }
  });

  // Мутация для удаления квартиры (админская версия)
  const deleteMutation = useMutation({
    mutationFn: apartmentsAPI.adminDeleteApartment,
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-apartments']);
      queryClient.invalidateQueries(['admin-dashboard-stats']);
      queryClient.invalidateQueries(['admin-apartments-statistics']);
      message.success('Квартира удалена');
    }
  });

  // Мутация для обновления квартиры
  const updateApartmentMutation = useMutation({
    mutationFn: ({ id, data }) => apartmentsAPI.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-apartments']);
      queryClient.invalidateQueries(['admin-dashboard-stats']);
      queryClient.invalidateQueries(['admin-apartments-statistics']);
      setEditModalVisible(false);
      form.resetFields();
      setShowHourlyPrice(false);
      setShowDailyPrice(false);
      setExistingPhotos([]);
      message.success('Квартира обновлена');
    },
    onError: () => {
      message.error('Ошибка при обновлении квартиры');
    }
  });

  // Мутация для удаления фотографии
  const deletePhotoMutation = useMutation({
    mutationFn: (photoId) => apartmentsAPI.deletePhoto(photoId),
    onSuccess: (_, photoId) => {
      setExistingPhotos(prev => prev.filter(photo => photo.id !== photoId));
      message.success('Фотография удалена');
    },
    onError: () => {
      message.error('Ошибка при удалении фотографии');
    }
  });

  const handleStatusChange = (values) => {
    updateStatusMutation.mutate({
      id: selectedApartment.id,
      ...values
    });
  };

  const handleApartmentTypeChange = (values) => {
    updateApartmentTypeMutation.mutate({
      id: selectedApartment.id,
      apartmentTypeId: values.apartment_type_id
    });
  };

  // Мутация для обновления счетчиков
  const updateCountersMutation = useMutation({
    mutationFn: ({ id, data }) => apartmentsAPI.adminUpdateCounters(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-apartments']);
      queryClient.invalidateQueries(['admin-dashboard-stats']);
      queryClient.invalidateQueries(['admin-apartments-statistics']);
      setCountersModalVisible(false);
      countersForm.resetFields();
      message.success('Счетчики обновлены');
    },
    onError: () => {
      message.error('Ошибка при обновлении счетчиков');
    }
  });

  // Мутация для сброса счетчиков
  const resetCountersMutation = useMutation({
    mutationFn: (id) => apartmentsAPI.adminResetCounters(id),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-apartments']);
      queryClient.invalidateQueries(['admin-dashboard-stats']);
      queryClient.invalidateQueries(['admin-apartments-statistics']);
      setCountersModalVisible(false);
      countersForm.resetFields();
      message.success('Счетчики сброшены');
    },
    onError: () => {
      message.error('Ошибка при сбросе счетчиков');
    }
  });

  const handleCountersUpdate = (values) => {
    updateCountersMutation.mutate({
      id: selectedApartment.id,
      data: values
    });
  };

  const handleCountersReset = () => {
    Modal.confirm({
      title: 'Сброс счетчиков',
      content: 'Вы уверены, что хотите сбросить счетчики просмотров и бронирований?',
      onOk: () => {
        resetCountersMutation.mutate(selectedApartment.id);
      }
    });
  };

  const handleDelete = (id) => {
    deleteMutation.mutate(id);
  };

  const handleApartmentUpdate = (values) => {
    updateApartmentMutation.mutate({
      id: selectedApartment.id,
      data: values
    });
  };

  // Обработчики для выпадающих списков локаций
  const handleCityChange = (cityId) => {
    setSelectedCityId(cityId);
    setSelectedDistrictId(null);
    form.setFieldsValue({ district_id: undefined, microdistrict_id: undefined });
  };

  const handleDistrictChange = (districtId) => {
    setSelectedDistrictId(districtId);
    form.setFieldsValue({ microdistrict_id: undefined });
  };

  // Обработчик загрузки фотографий
  const handlePhotoUpload = ({ fileList }) => {
    const photos = fileList.map(file => {
      if (file.originFileObj) {
        return new Promise((resolve) => {
          const reader = new FileReader();
          reader.onload = () => resolve(reader.result.split(',')[1]); // убираем префикс data:image/jpeg;base64,
          reader.readAsDataURL(file.originFileObj);
        });
      }
      return file.response || file.url;
    });
    
    Promise.all(photos).then(base64Photos => {
      form.setFieldsValue({ photos_base64: base64Photos });
    });
  };

  // Обработчик удаления существующей фотографии
  const handleDeleteExistingPhoto = (photoId) => {
    deletePhotoMutation.mutate(photoId);
  };

  // Обработчики изменения типов аренды
  const handleRentalTypeChange = (type, checked) => {
    if (type === 'hourly') {
      setShowHourlyPrice(checked);
    } else if (type === 'daily') {
      setShowDailyPrice(checked);
    }
  };

  const handleViewContract = async (contractId) => {
    try {
      const response = await contractsAPI.getContractHTML(contractId);
      const htmlContent = response.data.html; // Исправлен путь
      
      const newWindow = window.open('', '_blank');
      newWindow.document.write(htmlContent);
      newWindow.document.close();
    } catch (error) {
      message.error('Ошибка при загрузке договора');
      console.error('Contract error:', error);
    }
  };

  const getStatusColor = (status) => {
    const colors = {
      'pending': 'orange',
      'approved': 'green',
      'rejected': 'red',
      'blocked': 'red',
      'inactive': 'gray'
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status) => {
    const texts = {
      'pending': 'На модерации',
      'approved': 'Одобрено',
      'rejected': 'Отклонено',
      'blocked': 'Заблокировано',
      'inactive': 'Неактивно'
    };
    return texts[status] || status;
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: 'Адрес',
      key: 'address',
      render: (_, record) => (
        <div>
          <div className="font-medium">
            {record.street}, д. {record.building}
          </div>
          <div className="text-gray-500 text-sm">
            кв. {record.apartment_number}, {record.city?.name}
          </div>
        </div>
      ),
    },
    {
      title: 'Владелец',
      key: 'owner',
      render: (_, record) => (
        <div>
          <div>
            {record.owner?.user?.first_name} {record.owner?.user?.last_name}
          </div>
          <div className="text-gray-500 text-sm">
            {record.owner?.user?.phone || '—'}
          </div>
        </div>
      ),
    },
    {
      title: 'Детали',
      key: 'details',
      render: (_, record) => (
        <div>
          <div>{record.room_count}-комн., {record.total_area} м²</div>
          <div className="text-gray-500 text-sm">
            {record.floor}/{record.total_floors} этаж
          </div>
        </div>
      ),
    },
    {
      title: 'Тип квартиры',
      key: 'apartment_type',
      render: (_, record) => {
        const apartmentTypeName = apartmentTypes?.data?.find(type => type.id === record.apartment_type_id)?.name;
        return (
          <div>
            {apartmentTypeName ? (
              <Tag color="blue">{apartmentTypeName}</Tag>
            ) : (
              <Text type="secondary">Не указан</Text>
            )}
          </div>
        );
      },
    },
    {
      title: 'Цена',
      key: 'price',
      render: (_, record) => (
        <div>
          <Text strong>{record.price?.toLocaleString() || '—'} ₸/час</Text>
          {record.daily_price && (
            <div className="text-gray-500 text-sm">
              {record.daily_price?.toLocaleString()} ₸/день
            </div>
          )}
        </div>
      ),
    },
    {
      title: 'Статистика',
      key: 'counters',
      render: (_, record) => (
        <div>
          <div className="text-sm">
            <EyeOutlined className="mr-1" />
            {record.view_count || 0} просмотров
          </div>
          <div className="text-sm text-gray-500">
            <TagOutlined className="mr-1" />
            {record.booking_count || 0} бронирований
          </div>
        </div>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: 'Создано',
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
                setSelectedApartment(record);
                setDetailsVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="История бронирований">
            <Button
              type="text"
              icon={<HistoryOutlined />}
              onClick={() => {
                setSelectedApartmentForHistory(record);
                setHistoryModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Управление счетчиками">
            <Button
              type="text"
              icon={<TagOutlined />}
              onClick={() => {
                setSelectedApartment(record);
                setCountersModalVisible(true);
                countersForm.setFieldsValue({
                  view_count: record.view_count || 0,
                  booking_count: record.booking_count || 0
                });
              }}
            />
          </Tooltip>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<SettingOutlined />}
              onClick={() => {
                setSelectedApartment(record);
                setEditModalVisible(true);
                
                // Устанавливаем выбранные локации для подгрузки зависимых списков
                if (record.city_id) {
                  setSelectedCityId(record.city_id);
                }
                if (record.district_id) {
                  setSelectedDistrictId(record.district_id);
                }
                
                // Устанавливаем состояния для отображения полей цен
                setShowHourlyPrice(record.rental_type_hourly || false);
                setShowDailyPrice(record.rental_type_daily || false);
                
                // Устанавливаем существующие фотографии
                setExistingPhotos(record.photos || []);
                
                // Заполняем форму текущими данными
                form.setFieldsValue({
                  ...record,
                  amenity_ids: record.amenities?.map(amenity => amenity.id) || [],
                  house_rule_ids: record.house_rules?.map(rule => rule.id) || [],
                  city_id: record.city_id,
                  district_id: record.district_id,
                  microdistrict_id: record.microdistrict_id,
                  condition_id: record.condition_id,
                  rental_type_hourly: record.rental_type_hourly || false,
                  rental_type_daily: record.rental_type_daily || false,
                  latitude: record.location?.latitude || record.latitude,
                  longitude: record.location?.longitude || record.longitude,
                  listing_type: record.listing_type,
                });
              }}
            />
          </Tooltip>
          <Tooltip title="Изменить статус">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedApartment(record);
                setStatusModalVisible(true);
                statusForm.setFieldsValue({
                  status: record.status,
                  comment: record.moderator_comment || ''
                });
              }}
            />
          </Tooltip>
          <Tooltip title="Изменить тип квартиры">
            <Button
              type="text"
              icon={<TagOutlined style={{ color: '#1890ff' }} />}
              onClick={() => {
                setSelectedApartment(record);
                setApartmentTypeModalVisible(true);
                apartmentTypeForm.setFieldsValue({
                  apartment_type_id: record.apartment_type_id || undefined
                });
              }}
            />
          </Tooltip>
          <Tooltip title="Удалить">
            <Popconfirm
              title="Удалить квартиру?"
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
    <div className="p-6">
      <div className="mb-6">
        <Title level={2}>Управление квартирами</Title>
        <Text type="secondary">
          Модерация и управление квартирами в системе
        </Text>
      </div>

      {/* Статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Всего квартир"
              value={apartmentsStatistics?.data?.summary?.total_apartments || 0}
              loading={isLoadingStatistics}
              prefix={<EnvironmentOutlined />}
              valueStyle={{ fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setFilters(prev => ({ ...prev, status: 'pending', page: 1 }));
            }}
          >
            <Statistic
              title="На модерации"
              value={apartmentsStatistics?.data?.by_status?.pending || 0}
              loading={isLoadingStatistics}
              prefix={<Badge status="warning" />}
              valueStyle={{ color: '#faad14', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setFilters(prev => ({ ...prev, status: 'approved', page: 1 }));
            }}
          >
            <Statistic
              title="Одобрено"
              value={apartmentsStatistics?.data?.by_status?.approved || 0}
              loading={isLoadingStatistics}
              prefix={<Badge status="success" />}
              valueStyle={{ color: '#52c41a', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setFilters(prev => ({ ...prev, status: 'rejected', page: 1 }));
            }}
          >
            <Statistic
              title="Отклонено"
              value={apartmentsStatistics?.data?.by_status?.rejected || 0}
              loading={isLoadingStatistics}
              prefix={<Badge status="error" />}
              valueStyle={{ color: '#f5222d', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Дополнительная статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Среднее кол-во комнат"
              value={Object.keys(apartmentsStatistics?.data?.by_room_count || {}).reduce((sum, rooms) => 
                sum + parseInt(rooms) * (apartmentsStatistics?.data?.by_room_count[rooms] || 0), 0) / 
                (apartmentsStatistics?.data?.summary?.total_apartments || 1) || 0}
              loading={isLoadingStatistics}
              precision={1}
              valueStyle={{ color: '#1890ff', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Средняя площадь"
              value={apartmentsStatistics?.data?.summary?.avg_area || 0}
              loading={isLoadingStatistics}
              suffix="м²"
              valueStyle={{ color: '#52c41a', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Средняя цена"
              value={apartmentsStatistics?.data?.summary?.avg_price || 0}
              loading={isLoadingStatistics}
              suffix="₸"
              valueStyle={{ color: '#722ed1', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Средняя цена/день"
              value={apartmentsStatistics?.data?.summary?.avg_daily_price || 0}
              loading={isLoadingStatistics}
              suffix="₸"
              valueStyle={{ color: '#fa8c16', fontSize: isMobile ? '20px' : '24px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Фильтры */}
      <Card className="mb-6">
        <div className="flex flex-wrap gap-4 items-end">
          <div className="flex-1 min-w-[180px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
            <Select
              placeholder="Выберите статус"
              className="w-full"
              allowClear
              value={filters.status}
              onChange={(value) => setFilters({ ...filters, status: value })}
            >
              <Option value="pending">На модерации</Option>
              <Option value="approved">Одобрено</Option>
              <Option value="needs_revision">Требует доработки</Option>
              <Option value="rejected">Отклонено</Option>
            </Select>
          </div>
          
          <div className="flex-1 min-w-[180px]">
            <LocationFilter
              showCity={true}
              showDistrict={false}
              showMicrodistrict={false}
              cityId={filters.city_id || null}
              onCityChange={(value) => setFilters({ ...filters, city_id: value })}
              layout="vertical"
              size="default"
            />
          </div>
          
          <div className="flex-1 min-w-[180px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Тип листинга</label>
            <Select
              placeholder="Тип листинга"
              className="w-full"
              allowClear
              value={filters.listing_type}
              onChange={(value) => setFilters({ ...filters, listing_type: value })}
            >
              <Option value="owner">Владелец</Option>
              <Option value="realtor">Риелтор</Option>
            </Select>
          </div>

          <div className="flex-1 min-w-[180px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Тип квартиры</label>
            <Select
              placeholder="Выберите тип"
              className="w-full"
              allowClear
              value={filters.apartment_type_id}
              onChange={(value) => setFilters({ ...filters, apartment_type_id: value })}
            >
              {apartmentTypes?.data?.map(type => (
                <Option key={type.id} value={type.id}>
                  {type.name}
                </Option>
              ))}
            </Select>
          </div>
          
          <div className="flex-1 min-w-[180px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Количество комнат</label>
            <Select
              placeholder="Количество комнат"
              className="w-full"
              allowClear
              value={filters.room_count}
              onChange={(value) => setFilters({ ...filters, room_count: value })}
            >
              <Option value="1">1 комната</Option>
              <Option value="2">2 комнаты</Option>
              <Option value="3">3 комнаты</Option>
              <Option value="4">4+ комнат</Option>
            </Select>
          </div>
          
          <div className="flex-shrink-0">
            <Button 
              onClick={() => setFilters({})}
              className="px-8"
            >
              Сбросить
            </Button>
          </div>
        </div>
      </Card>

      {/* Таблица квартир */}
      <Card>
        <Table
          columns={columns}
          dataSource={apartmentsData?.data?.apartments || []}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 1200 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: apartmentsData?.data?.pagination?.total || 0,
            showSizeChanger: !isMobile,
            showQuickJumper: !isMobile,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} квартир`,
            responsive: true,
            simple: isMobile,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            },
          }}
        />
      </Card>

      {/* Модал деталей квартиры */}
      <Drawer
        title="Детали квартиры"
        width={isMobile ? '100%' : 720}
        open={detailsVisible}
        onClose={() => setDetailsVisible(false)}
        placement={isMobile ? 'bottom' : 'right'}
        height={isMobile ? '90%' : undefined}
      >
        {selectedApartment && (
          <div>
            <Descriptions 
              column={isMobile ? 1 : 2} 
              bordered
              size={isMobile ? 'small' : 'default'}
            >
              <Descriptions.Item label="ID" span={isMobile ? 1 : 2}>
                {selectedApartment.id}
              </Descriptions.Item>
              <Descriptions.Item label="Адрес" span={isMobile ? 1 : 2}>
                <div className="break-words">
                  {selectedApartment.street}, д. {selectedApartment.building}, кв. {selectedApartment.apartment_number}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Город" span={1}>
                {selectedApartment.city?.name || '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Район" span={1}>
                {selectedApartment.district?.name || '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Общая площадь">
                {selectedApartment.total_area} м²
              </Descriptions.Item>
              <Descriptions.Item label="Площадь кухни">
                {selectedApartment.kitchen_area} м²
              </Descriptions.Item>
              <Descriptions.Item label="Этаж">
                {selectedApartment.floor}/{selectedApartment.total_floors}
              </Descriptions.Item>
              <Descriptions.Item label="Комнат">
                {selectedApartment.room_count}
              </Descriptions.Item>
              <Descriptions.Item label="Тип квартиры">
                {(() => {
                  const apartmentTypeName = apartmentTypes?.data?.find(type => type.id === selectedApartment.apartment_type_id)?.name;
                  return apartmentTypeName ? (
                    <Tag color="blue">{apartmentTypeName}</Tag>
                  ) : (
                    <Text type="secondary">Не указан</Text>
                  );
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="Тип аренды">
                {(() => {
                  const types = [];
                  if (selectedApartment.rental_type_hourly) types.push('Почасовая');
                  if (selectedApartment.rental_type_daily) types.push('Посуточная');
                  return types.length > 0 ? types.join(', ') : '—';
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="Цена за час">
                <Text className="font-mono">
                  {selectedApartment.price?.toLocaleString()} ₸
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="Цена за день">
                <Text className="font-mono">
                  {selectedApartment.daily_price?.toLocaleString()} ₸
                </Text>
              </Descriptions.Item>
                          <Descriptions.Item label="Статус">
              <Tag color={getStatusColor(selectedApartment.status)}>
                {getStatusText(selectedApartment.status)}
              </Tag>
            </Descriptions.Item>
            {selectedApartment.apartment_type && (
              <Descriptions.Item label="Тип квартиры">
                <strong>{selectedApartment.apartment_type.name}</strong>
                {selectedApartment.apartment_type.description && (
                  <div style={{ fontSize: '12px', color: '#666', marginTop: '2px' }}>
                    {selectedApartment.apartment_type.description}
                  </div>
                )}
              </Descriptions.Item>
            )}
              <Descriptions.Item label="Состояние">
                {selectedApartment.condition?.name || '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Доступна">
                <Tag color={selectedApartment.is_free ? 'green' : 'red'}>
                  {selectedApartment.is_free ? 'Да' : 'Нет'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Владелец" span={isMobile ? 1 : 2}>
                <div className="break-words">
                  {selectedApartment.owner?.user?.first_name} {selectedApartment.owner?.user?.last_name} 
                  {selectedApartment.owner?.user?.phone && (
                    <div className="text-sm text-gray-500">
                      {selectedApartment.owner?.user?.phone}
                    </div>
                  )}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Создано">
                <div className="font-mono text-sm">
                  {dayjs(selectedApartment.created_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedApartment.created_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Обновлено">
                <div className="font-mono text-sm">
                  {dayjs(selectedApartment.updated_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedApartment.updated_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              {selectedApartment.description && (
                <Descriptions.Item label="Описание" span={isMobile ? 1 : 2}>
                  <div className="break-words">
                    {selectedApartment.description}
                  </div>
                </Descriptions.Item>
              )}
              {selectedApartment.moderator_comment && (
                <Descriptions.Item label="Комментарий модератора" span={isMobile ? 1 : 2}>
                  <div className="break-words p-3 bg-yellow-50 border border-yellow-200 rounded-md">
                    {selectedApartment.moderator_comment}
                  </div>
                </Descriptions.Item>
              )}
              {selectedApartment.contract_id && (
                <Descriptions.Item label="Договор" span={isMobile ? 1 : 2}>
                  <Button 
                    type="primary"
                    size="large"
                    onClick={() => handleViewContract(selectedApartment.contract_id)}
                    className="bg-blue-600 hover:bg-blue-700 border-blue-600 hover:border-blue-700"
                  >
                    📄 Просмотреть договор
                  </Button>
                </Descriptions.Item>
              )}
            </Descriptions>

            {selectedApartment.photos && selectedApartment.photos.length > 0 && (
              <div className="mt-6">
                <Title level={4}>Фотографии</Title>
                <div className={`grid gap-4 ${isMobile ? 'grid-cols-2' : 'grid-cols-3'}`}>
                  {selectedApartment.photos.map((photo, index) => (
                    <Image
                      key={index}
                      src={photo.url}
                      alt={`Фото ${index + 1}`}
                      className="rounded-lg"
                    />
                  ))}
                </div>
              </div>
            )}

            {selectedApartment.amenities && selectedApartment.amenities.length > 0 && (
              <div className="mt-6">
                <Title level={4}>Удобства</Title>
                <div className="flex flex-wrap gap-2">
                  {selectedApartment.amenities.map((amenity, index) => (
                    <Tag key={amenity.id || index} color="blue" className={isMobile ? 'text-xs' : ''}>
                      {amenity.name || amenity}
                    </Tag>
                  ))}
                </div>
              </div>
            )}

            {selectedApartment.rules && selectedApartment.rules.length > 0 && (
              <div className="mt-6">
                <Title level={4}>Правила</Title>
                <div className="flex flex-wrap gap-2">
                  {selectedApartment.rules.map((rule, index) => (
                    <Tag key={rule.id || index} color="orange" className={isMobile ? 'text-xs' : ''}>
                      {rule.name || rule}
                    </Tag>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </Drawer>

      {/* Модал изменения статуса */}
      <Modal
        title="Изменить статус квартиры"
        open={statusModalVisible}
        onCancel={() => setStatusModalVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 520}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={statusForm}
          layout="vertical"
          onFinish={handleStatusChange}
        >
          <Form.Item
            name="status"
            label="Статус"
            rules={[{ required: true, message: 'Выберите статус' }]}
          >
            <Select>
              <Option value="pending">На модерации</Option>
              <Option value="approved">Одобрено</Option>
              <Option value="rejected">Отклонено</Option>
              <Option value="blocked">Заблокировано</Option>
              <Option value="inactive">Неактивно</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="comment"
            label="Комментарий"
          >
            <TextArea 
              rows={4} 
              placeholder="Причина изменения статуса..."
            />
          </Form.Item>
          <Form.Item>
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={updateStatusMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Сохранить
              </Button>
              <Button 
                onClick={() => setStatusModalVisible(false)}
                className={isMobile ? 'w-full' : ''}
              >
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Модал изменения типа квартиры */}
      <Modal
        title="Изменить тип квартиры"
        open={apartmentTypeModalVisible}
        onCancel={() => setApartmentTypeModalVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 420}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={apartmentTypeForm}
          layout="vertical"
          onFinish={handleApartmentTypeChange}
        >
          <Form.Item
            name="apartment_type_id"
            label="Тип квартиры"
            rules={[{ required: true, message: 'Выберите тип квартиры' }]}
          >
            <Select 
              placeholder={
                selectedApartment?.apartment_type_id 
                  ? `Текущий: ${apartmentTypes?.data?.find(type => type.id === selectedApartment.apartment_type_id)?.name || 'Неизвестно'}` 
                  : "Выберите тип квартиры"
              }
              allowClear
            >
              {apartmentTypes?.data?.map(type => (
                <Option key={type.id} value={type.id}>
                  {type.name} - {type.description}
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item>
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={updateApartmentTypeMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Сохранить
              </Button>
              <Button 
                onClick={() => setApartmentTypeModalVisible(false)}
                className={isMobile ? 'w-full' : ''}
              >
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Модальное окно редактирования квартиры */}
      <Modal
        title="Редактировать квартиру"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          setShowHourlyPrice(false);
          setShowDailyPrice(false);
          setExistingPhotos([]);
          form.resetFields();
        }}
        footer={null}
        width={isMobile ? '95%' : 1000}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleApartmentUpdate}
        >
          {/* Локация */}
          <Row gutter={16}>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="city_id"
                label="Город"
                rules={[{ required: true, message: 'Выберите город' }]}
              >
                <Select 
                  placeholder="Выберите город"
                  onChange={handleCityChange}
                  showSearch
                  optionFilterProp="children"
                >
                  {(citiesData?.data || citiesData || []).map(city => (
                    <Option key={city.id} value={city.id}>{city.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="district_id"
                label="Район"
              >
                <Select 
                  placeholder="Выберите район"
                  onChange={handleDistrictChange}
                  disabled={!selectedCityId}
                  showSearch
                  optionFilterProp="children"
                >
                  {(districtsData?.data || districtsData || []).map(district => (
                    <Option key={district.id} value={district.id}>{district.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="microdistrict_id"
                label="Микрорайон"
              >
                <Select 
                  placeholder="Выберите микрорайон"
                  disabled={!selectedDistrictId}
                  showSearch
                  optionFilterProp="children"
                >
                  {(microdistrictsData?.data || microdistrictsData || []).map(microdistrict => (
                    <Option key={microdistrict.id} value={microdistrict.id}>{microdistrict.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          {/* Адрес */}
          <Row gutter={16}>
            <Col span={isMobile ? 24 : 12}>
              <Form.Item
                name="street"
                label="Улица"
                rules={[{ required: true, message: 'Введите название улицы' }]}
              >
                <Input />
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 12}>
              <Form.Item
                name="building"
                label="Дом"
                rules={[{ required: true, message: 'Введите номер дома' }]}
              >
                <Input />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={16}>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="apartment_number"
                label="Номер квартиры"
                rules={[{ required: true, message: 'Введите номер квартиры' }]}
              >
                <InputNumber min={1} className="w-full" />
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="room_count"
                label="Количество комнат"
                rules={[{ required: true, message: 'Введите количество комнат' }]}
              >
                <InputNumber min={1} className="w-full" />
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="floor"
                label="Этаж"
                rules={[{ required: true, message: 'Введите этаж' }]}
              >
                <InputNumber min={1} className="w-full" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="total_floors"
                label="Всего этажей"
                rules={[{ required: true, message: 'Введите общее количество этажей' }]}
              >
                <InputNumber min={1} className="w-full" />
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="total_area"
                label="Общая площадь (м²)"
                rules={[{ required: true, message: 'Введите общую площадь' }]}
              >
                <InputNumber min={1} step={0.1} className="w-full" />
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="kitchen_area"
                label="Площадь кухни (м²)"
                rules={[{ required: true, message: 'Введите площадь кухни' }]}
              >
                <InputNumber min={1} step={0.1} className="w-full" />
              </Form.Item>
            </Col>
          </Row>

          {/* Состояние и координаты */}
          <Row gutter={16}>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="condition_id"
                label="Состояние квартиры"
                rules={[{ required: true, message: 'Выберите состояние' }]}
              >
                <Select placeholder="Выберите состояние">
                  {(conditionsData?.data || conditionsData || []).map(condition => (
                    <Option key={condition.id} value={condition.id}>{condition.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="latitude"
                label="Широта"
                rules={[{ required: false, message: 'Введите широту' }]}
              >
                <InputNumber step={0.000001} className="w-full" placeholder="43.238949" />
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item
                name="longitude"
                label="Долгота"
                rules={[{ required: false, message: 'Введите долготу' }]}
              >
                <InputNumber step={0.000001} className="w-full" placeholder="76.889709" />
              </Form.Item>
            </Col>
          </Row>

          {/* Условное отображение полей цен */}
          {(showHourlyPrice || showDailyPrice) && (
            <Row gutter={16}>
              {showHourlyPrice && (
                <Col span={isMobile ? 24 : 12}>
                  <Form.Item
                    name="price"
                    label="Цена за час (₸)"
                    rules={[{ required: showHourlyPrice, message: 'Введите цену за час' }]}
                  >
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                </Col>
              )}
              {showDailyPrice && (
                <Col span={isMobile ? 24 : 12}>
                  <Form.Item
                    name="daily_price"
                    label="Цена за сутки (₸)"
                    rules={[{ required: showDailyPrice, message: 'Введите цену за сутки' }]}
                  >
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                </Col>
              )}
            </Row>
          )}

          {/* Типы аренды и листинга */}
          <Row gutter={16}>
            <Col span={isMobile ? 24 : 8}>
              <Form.Item 
                label="Типы аренды"
                rules={[
                  {
                    validator: (_, value) => {
                      const hourly = form.getFieldValue('rental_type_hourly');
                      const daily = form.getFieldValue('rental_type_daily');
                      if (!hourly && !daily) {
                        return Promise.reject('Выберите хотя бы один тип аренды');
                      }
                      return Promise.resolve();
                    }
                  }
                ]}
              >
                <div className="space-y-2">
                  <Form.Item name="rental_type_hourly" valuePropName="checked" noStyle>
                    <Checkbox onChange={(e) => handleRentalTypeChange('hourly', e.target.checked)}>
                      Почасовая аренда
                    </Checkbox>
                  </Form.Item>
                  <Form.Item name="rental_type_daily" valuePropName="checked" noStyle>
                    <Checkbox onChange={(e) => handleRentalTypeChange('daily', e.target.checked)}>
                      Посуточная аренда
                    </Checkbox>
                  </Form.Item>
                  <div className="text-xs text-gray-500 mt-2">
                    Выберите типы аренды, чтобы отобразить соответствующие поля цен
                  </div>
                </div>
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 16}>
              <Form.Item
                name="listing_type"
                label="Тип листинга"
                rules={[{ required: true, message: 'Выберите тип листинга' }]}
              >
                <Select placeholder="Выберите тип листинга">
                  <Option value="owner">Владелец</Option>
                  <Option value="realtor">Риелтор</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="residential_complex"
            label="Жилой комплекс"
          >
            <Input />
          </Form.Item>

          {/* Удобства и правила */}
          <Row gutter={16}>
            <Col span={isMobile ? 24 : 12}>
              <Form.Item
                name="amenity_ids"
                label="Удобства"
              >
                <Select
                  mode="multiple"
                  placeholder="Выберите удобства"
                  showSearch
                  optionFilterProp="children"
                >
                  {(amenitiesData?.data || amenitiesData || []).map(amenity => (
                    <Option key={amenity.id} value={amenity.id}>{amenity.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={isMobile ? 24 : 12}>
              <Form.Item
                name="house_rule_ids"
                label="Правила дома"
              >
                <Select
                  mode="multiple"
                  placeholder="Выберите правила"
                  showSearch
                  optionFilterProp="children"
                >
                  {(houseRulesData?.data || houseRulesData || []).map(rule => (
                    <Option key={rule.id} value={rule.id}>{rule.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="description"
            label="Описание"
          >
            <TextArea rows={4} />
          </Form.Item>

          {/* Управление фотографиями */}
          <Form.Item label="Фотографии">
            {/* Существующие фотографии */}
            {existingPhotos.length > 0 && (
              <div className="mb-4">
                <div className="text-sm font-medium mb-2">Текущие фотографии:</div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {existingPhotos.map((photo) => (
                    <div key={photo.id} className="relative group">
                      <Image
                        src={photo.url}
                        alt="Фото квартиры"
                        className="w-full h-24 object-cover rounded-lg"
                      />
                      <Button
                        type="primary"
                        danger
                        size="small"
                        icon={<DeleteOutlined />}
                        className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={() => handleDeleteExistingPhoto(photo.id)}
                        loading={deletePhotoMutation.isPending}
                      />
                    </div>
                  ))}
                </div>
              </div>
            )}
            
            {/* Загрузка новых фотографий */}
            <Form.Item name="photos_base64" noStyle>
              <Upload
                listType="picture-card"
                multiple
                beforeUpload={() => false} // Предотвращаем автоматическую загрузку
                onChange={handlePhotoUpload}
                accept="image/*"
                maxCount={10}
              >
                <div>
                  <PlusOutlined />
                  <div style={{ marginTop: 8 }}>Добавить фото</div>
                </div>
              </Upload>
            </Form.Item>
            
            <div className="text-sm text-gray-500 mt-2">
              Поддерживаются форматы: JPG, PNG, GIF. Максимум 10 фотографий всего.
            </div>
          </Form.Item>

          <Form.Item>
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={updateApartmentMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Сохранить изменения
              </Button>
              <Button 
                onClick={() => setEditModalVisible(false)}
                className={isMobile ? 'w-full' : ''}
              >
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

            {/* Модал истории бронирований квартиры */}
      <ApartmentBookingHistoryModal
        visible={historyModalVisible}
        onClose={() => {
          setHistoryModalVisible(false);
          setSelectedApartmentForHistory(null);
        }}
        apartment={selectedApartmentForHistory}
      />

      {/* Модал управления счетчиками */}
      <Modal
        title={
          <Space>
            <TagOutlined />
            Управление счетчиками квартиры
          </Space>
        }
        open={countersModalVisible}
        onCancel={() => {
          setCountersModalVisible(false);
          countersForm.resetFields();
        }}
        footer={null}
        width={500}
      >
        <Form
          form={countersForm}
          layout="vertical"
          onFinish={handleCountersUpdate}
        >
          <Card size="small" className="mb-4">
            <div className="text-center">
              <div className="text-lg font-medium">
                {selectedApartment?.street}, д. {selectedApartment?.building}, кв. {selectedApartment?.apartment_number}
              </div>
              <div className="text-gray-500">
                {selectedApartment?.city?.name}
              </div>
            </div>
          </Card>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="view_count"
                label={
                  <Space>
                    <EyeOutlined />
                    Просмотры
                  </Space>
                }
                rules={[
                  { required: true, message: 'Укажите количество просмотров' },
                  { type: 'number', min: 0, message: 'Значение не может быть отрицательным' }
                ]}
              >
                <InputNumber
                  className="w-full"
                  placeholder="0"
                  min={0}
                  formatter={value => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  parser={value => value.replace(/\$\s?|(,*)/g, '')}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="booking_count"
                label={
                  <Space>
                    <TagOutlined />
                    Бронирования
                  </Space>
                }
                rules={[
                  { required: true, message: 'Укажите количество бронирований' },
                  { type: 'number', min: 0, message: 'Значение не может быть отрицательным' }
                ]}
              >
                <InputNumber
                  className="w-full"
                  placeholder="0"
                  min={0}
                  formatter={value => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  parser={value => value.replace(/\$\s?|(,*)/g, '')}
                />
              </Form.Item>
            </Col>
          </Row>

          <div className="text-sm text-gray-500 mb-4">
            💡 Эти счетчики влияют на популярность квартиры в поиске и статистике.
          </div>

          <Form.Item className="mb-0">
            <Space direction="vertical" className="w-full">
              <Space className="w-full justify-between">
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={updateCountersMutation.isPending}
                  icon={<SettingOutlined />}
                >
                  Обновить счетчики
                </Button>
                <Button
                  danger
                  onClick={handleCountersReset}
                  loading={resetCountersMutation.isPending}
                  icon={<DeleteOutlined />}
                >
                  Сбросить в ноль
                </Button>
              </Space>
              <Button
                onClick={() => {
                  setCountersModalVisible(false);
                  countersForm.resetFields();
                }}
                className="w-full"
              >
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ApartmentsPage; 