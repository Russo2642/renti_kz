import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  HomeOutlined,
  PlusOutlined
} from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Card, Checkbox, Col, Descriptions, Drawer, Form, Image, Input, InputNumber, Modal, Popconfirm, Row, Select, Space, Table, Tag, Tooltip, Typography, Upload, message } from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useState } from 'react';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import { apartmentsAPI, contractsAPI, dictionariesAPI, locationsAPI } from '../../lib/api.js';

const { Option } = Select;
const { TextArea } = Input;
const { Title, Text } = Typography;

const OwnerApartmentsPage = () => {
  const [modalVisible, setModalVisible] = useState(false);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [selectedApartment, setSelectedApartment] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [form] = Form.useForm();
  const [fileList, setFileList] = useState([]);
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

  // Получаем квартиры владельца
  const { data: apartments, isLoading } = useQuery({
    queryKey: ['my-apartments', currentPage, pageSize],
    queryFn: () => {
      const params = {
        page: currentPage,
        page_size: pageSize,
      };
      return apartmentsAPI.getMyApartments(params);
    }
  });

  // Получаем справочники
  const { data: conditions } = useQuery({
    queryKey: ['conditions'],
    queryFn: () => dictionariesAPI.getConditions(),
  });

  const { data: amenities } = useQuery({
    queryKey: ['amenities'],
    queryFn: () => dictionariesAPI.getAmenities(),
  });

  const { data: houseRules } = useQuery({
    queryKey: ['house-rules'],
    queryFn: () => dictionariesAPI.getHouseRules(),
  });

  const { data: cities } = useQuery({
    queryKey: ['cities'],
    queryFn: () => locationsAPI.getCities(),
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

  // Получаем реальные данные квартир из API
  const apartmentsList = apartments?.data?.apartments || [];

  // Мутация для создания/обновления квартиры
  const saveApartmentMutation = useMutation({
    mutationFn: (data) => {
      if (selectedApartment) {
        return apartmentsAPI.update(selectedApartment.id, data);
      } else {
        return apartmentsAPI.create(data);
      }
    },
    onSuccess: () => {
      message.success(selectedApartment ? 'Квартира обновлена' : 'Квартира создана');
      setModalVisible(false);
      queryClient.invalidateQueries(['my-apartments']);
      form.resetFields();
      setFileList([]);
      setShowHourlyPrice(false);
      setShowDailyPrice(false);
      setExistingPhotos([]);
      setSelectedCityId(null);
      setSelectedDistrictId(null);
    },
    onError: () => {
      message.error('Ошибка при сохранении квартиры');
    },
  });

  // Мутация для удаления квартиры
  const deleteApartmentMutation = useMutation({
    mutationFn: (id) => apartmentsAPI.delete(id),
    onSuccess: () => {
      message.success('Квартира удалена');
      queryClient.invalidateQueries(['my-apartments']);
    },
    onError: () => {
      message.error('Ошибка при удалении квартиры');
    },
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

  const handleCreate = () => {
    setSelectedApartment(null);
    form.resetFields();
    setFileList([]);
    setModalVisible(true);
  };

  const handleView = (apartment) => {
    setSelectedApartment(apartment);
    setDetailsVisible(true);
  };

  const handleEdit = (apartment) => {
    setSelectedApartment(apartment);
    
    // Устанавливаем городе и районы для каскадных селектов
    if (apartment.city_id) {
      setSelectedCityId(apartment.city_id);
    }
    if (apartment.district_id) {
      setSelectedDistrictId(apartment.district_id);
    }
    
    // Устанавливаем состояния для отображения полей цен
    setShowHourlyPrice(apartment.rental_type_hourly || false);
    setShowDailyPrice(apartment.rental_type_daily || false);
    
    // Устанавливаем существующие фотографии
    setExistingPhotos(apartment.photos || []);
    
    form.setFieldsValue({
      ...apartment,
      amenity_ids: apartment.amenities?.map(amenity => amenity.id) || [],
      house_rule_ids: apartment.house_rules?.map(rule => rule.id) || [],
      city_id: apartment.city_id,
      district_id: apartment.district_id,
      microdistrict_id: apartment.microdistrict_id,
      condition_id: apartment.condition_id,
      rental_type_hourly: apartment.rental_type_hourly || false,
      rental_type_daily: apartment.rental_type_daily || false,
      latitude: apartment.location?.latitude || apartment.latitude,
      longitude: apartment.location?.longitude || apartment.longitude,
      listing_type: apartment.listing_type,
    });
    
    setFileList([]);
    setModalVisible(true);
  };

  const handleDelete = (id) => {
    deleteApartmentMutation.mutate(id);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      saveApartmentMutation.mutate(values);
    } catch (error) {
      console.error('Ошибка валидации:', error);
    }
  };

  const handleModalCancel = () => {
    setModalVisible(false);
    setSelectedApartment(null);
    form.resetFields();
    setFileList([]);
    setShowHourlyPrice(false);
    setShowDailyPrice(false);
    setExistingPhotos([]);
    setSelectedCityId(null);
    setSelectedDistrictId(null);
  };

  // Обработчики изменения типов аренды
  const handleRentalTypeChange = (type, checked) => {
    if (type === 'hourly') {
      setShowHourlyPrice(checked);
    } else if (type === 'daily') {
      setShowDailyPrice(checked);
    }
  };

  // Обработчик удаления существующей фотографии
  const handleDeleteExistingPhoto = (photoId) => {
    deletePhotoMutation.mutate(photoId);
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

  // Обработчик просмотра контракта
  const handleViewContract = async (apartmentId) => {
    try {
      // Сначала получаем ID договора через apartment
      const contractResponse = await contractsAPI.getByApartmentId(apartmentId);
      const contractId = contractResponse.data.id;
      
      // Затем получаем HTML договора
      const htmlResponse = await contractsAPI.getContractHTML(contractId);
      const htmlContent = htmlResponse.data.html;
      
      const newWindow = window.open('', '_blank');
      newWindow.document.write(htmlContent);
      newWindow.document.close();
    } catch (error) {
      console.error('Ошибка при получении контракта:', error);
      message.error('Ошибка при загрузке контракта');
    }
  };

  const uploadProps = {
    fileList,
    onChange: ({ fileList: newFileList }) => setFileList(newFileList),
    beforeUpload: () => false, // Предотвращаем автоматическую загрузку
    listType: 'picture-card',
  };

  const columns = [
    {
      title: 'Квартира',
      key: 'apartment',
      render: (record) => (
        <div className="flex items-center space-x-3">
          <div className={`rounded-lg overflow-hidden bg-gray-100 ${isMobile ? 'w-12 h-12' : 'w-16 h-16'}`}>
            {record.photos?.[0]?.url ? (
              <Image
                src={record.photos[0].url}
                alt="Квартира"
                className="w-full h-full object-cover"
                preview={false}
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <HomeOutlined className="text-gray-400 text-xl" />
              </div>
            )}
          </div>
          <div>
            <div className={`font-medium text-gray-900 ${isMobile ? 'text-sm' : ''}`}>
              {record.street}, {record.building}, кв. {record.apartment_number}
            </div>
            <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              {record.room_count}-комн., {record.total_area} м², {record.floor}/{record.total_floors} эт.
            </div>
            {!isMobile && (
              <div className="text-xs text-gray-400">
                ID: #{record.id}
              </div>
            )}
          </div>
        </div>
      ),
    },
    {
      title: 'Цена',
      key: 'price',
      render: (record) => (
        <div className="text-right">
          {record.price && (
            <div className={`font-semibold ${isMobile ? 'text-sm' : 'text-lg'}`}>
              {record.price.toLocaleString()} ₸/ч
            </div>
          )}
          {record.daily_price && (
            <div className={`font-medium ${isMobile ? 'text-xs' : 'text-sm'} text-blue-600`}>
              {record.daily_price.toLocaleString()} ₸/сут
            </div>
          )}
          {!record.price && !record.daily_price && (
            <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>Не указана</div>
          )}
        </div>
      ),
      sorter: (a, b) => (a.price || 0) - (b.price || 0),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const statusConfig = {
          approved: { color: 'green', text: 'Одобрена' },
          pending: { color: 'orange', text: 'На модерации' },
          rejected: { color: 'red', text: 'Отклонена' },
          blocked: { color: 'red', text: 'Заблокирована' },
          inactive: { color: 'gray', text: 'Неактивна' },
        };
        const config = statusConfig[status] || { color: 'default', text: status };
        return <Tag color={config.color} className={isMobile ? 'text-xs' : ''}>{config.text}</Tag>;
      },
      filters: [
        { text: 'Одобрена', value: 'approved' },
        { text: 'На модерации', value: 'pending' },
        { text: 'Отклонена', value: 'rejected' },
        { text: 'Заблокирована', value: 'blocked' },
        { text: 'Неактивна', value: 'inactive' },
      ],
      onFilter: (value, record) => record.status === value,
    },
    ...(isMobile ? [] : [{
      title: 'Удобства',
      key: 'amenities',
      render: (record) => (
        <div className="max-w-32">
          {record.amenities?.slice(0, 3).map((amenity, index) => (
            <Tag key={index} size="small" className="mb-1">
              {amenity.name}
            </Tag>
          ))}
          {record.amenities?.length > 3 && (
            <Tag size="small" color="blue">
              +{record.amenities.length - 3}
            </Tag>
          )}
        </div>
      ),
    }]),
    ...(isMobile ? [] : [{
      title: 'Создана',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD.MM.YYYY'),
      sorter: (a, b) => new Date(a.created_at) - new Date(b.created_at),
    }]),
    {
      title: 'Действия',
      key: 'actions',
      render: (record) => (
        <Space size={isMobile ? 'small' : 'middle'}>
          <Tooltip title="Просмотр">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleView(record)}
              size={isMobile ? 'small' : 'default'}
            />
          </Tooltip>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              size={isMobile ? 'small' : 'default'}
            />
          </Tooltip>
          <Popconfirm
            title="Удалить квартиру?"
            description="Это действие нельзя будет отменить"
            onConfirm={() => handleDelete(record.id)}
            okText="Удалить"
            cancelText="Отмена"
            okType="danger"
          >
            <Tooltip title="Удалить">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                size={isMobile ? 'small' : 'default'}
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
    <div className={`space-y-6 ${isMobile ? 'p-4' : 'p-6'}`}>
      {/* Заголовок и действия */}
      <div className={`${isMobile ? 'space-y-4' : 'flex flex-col lg:flex-row lg:justify-between lg:items-center space-y-4 lg:space-y-0'}`}>
        <div>
          <h1 className={`font-bold text-gray-900 ${isMobile ? 'text-xl' : 'text-2xl'}`}>Мои квартиры</h1>
          <p className={`text-gray-600 ${isMobile ? 'text-sm' : ''}`}>Управление объявлениями и статистика</p>
        </div>
        
        <Button 
          type="primary" 
          icon={<PlusOutlined />} 
          onClick={handleCreate}
          className={isMobile ? 'w-full' : ''}
        >
          Добавить квартиру
        </Button>
      </div>

      {/* Статистика */}
      <div className={`grid gap-4 ${isMobile ? 'grid-cols-2' : 'grid-cols-1 md:grid-cols-4'}`}>
        <Card>
          <div className="text-center">
            <div className={`font-bold text-blue-600 ${isMobile ? 'text-lg' : 'text-2xl'}`}>
              {apartmentsList.length}
            </div>
            <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>Всего квартир</div>
          </div>
        </Card>
        
        <Card>
          <div className="text-center">
            <div className={`font-bold text-green-600 ${isMobile ? 'text-lg' : 'text-2xl'}`}>
              {apartmentsList.filter(apt => apt.status === 'approved').length}
            </div>
            <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>Одобренных</div>
          </div>
        </Card>
        
        <Card>
          <div className="text-center">
            <div className={`font-bold text-orange-600 ${isMobile ? 'text-lg' : 'text-2xl'}`}>
              {apartmentsList.filter(apt => apt.status === 'pending').length}
            </div>
            <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>На модерации</div>
          </div>
        </Card>
        
        <Card>
          <div className="text-center">
            <div className={`font-bold text-red-600 ${isMobile ? 'text-lg' : 'text-2xl'}`}>
              {apartmentsList.filter(apt => ['rejected', 'blocked', 'inactive'].includes(apt.status)).length}
            </div>
            <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>Неактивных</div>
          </div>
        </Card>
      </div>

      {/* Таблица квартир */}
      <Card>
        <Table
          columns={columns}
          dataSource={apartmentsList}
          rowKey="id"
          scroll={{ x: isMobile ? 600 : 1200 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: apartments?.data?.pagination?.total || 0,
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

      {/* Модальное окно создания/редактирования */}
      <Modal
        title={selectedApartment ? 'Редактирование квартиры' : 'Добавление квартиры'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={handleModalCancel}
        width={isMobile ? '95%' : 1000}
        style={isMobile ? { top: 20 } : {}}
        okText="Сохранить"
        cancelText="Отмена"
        confirmLoading={saveApartmentMutation.isPending}
      >
        <Form
          form={form}
          layout="vertical"
          className="mt-4"
          onValuesChange={(changedValues) => {
            if (changedValues.city_id) {
              setSelectedCityId(changedValues.city_id);
              form.setFieldsValue({ district_id: undefined, microdistrict_id: undefined });
            }
            if (changedValues.district_id) {
              setSelectedDistrictId(changedValues.district_id);
              form.setFieldsValue({ microdistrict_id: undefined });
            }
          }}
        >
          {/* Локация */}
          <Row gutter={16}>
            <Col xs={24} sm={8}>
              <Form.Item
                name="city_id"
                label="Город"
                rules={[{ required: true, message: 'Выберите город' }]}
              >
                <Select placeholder="Выберите город" showSearch optionFilterProp="children">
                  {(cities?.data || cities || []).map(city => (
                    <Option key={city.id} value={city.id}>{city.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="district_id" label="Район">
                <Select placeholder="Выберите район" disabled={!selectedCityId} showSearch optionFilterProp="children">
                  {(districtsData?.data || districtsData || []).map(district => (
                    <Option key={district.id} value={district.id}>{district.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="microdistrict_id" label="Микрорайон">
                <Select placeholder="Выберите микрорайон" disabled={!selectedDistrictId} showSearch optionFilterProp="children">
                  {(microdistrictsData?.data || microdistrictsData || []).map(microdistrict => (
                    <Option key={microdistrict.id} value={microdistrict.id}>{microdistrict.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          {/* Адрес */}
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="street" label="Улица" rules={[{ required: true, message: 'Введите название улицы' }]}>
                <Input placeholder="Название улицы" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="building" label="Дом" rules={[{ required: true, message: 'Введите номер дома' }]}>
                <Input placeholder="Номер дома" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} sm={8}>
              <Form.Item name="apartment_number" label="Номер квартиры" rules={[{ required: true, message: 'Введите номер квартиры' }]}>
                <InputNumber min={1} className="w-full" placeholder="Номер квартиры" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="room_count" label="Количество комнат" rules={[{ required: true, message: 'Введите количество комнат' }]}>
                <InputNumber min={1} className="w-full" placeholder="Количество комнат" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="floor" label="Этаж" rules={[{ required: true, message: 'Введите этаж' }]}>
                <InputNumber min={1} className="w-full" placeholder="Этаж" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} sm={8}>
              <Form.Item name="total_floors" label="Всего этажей" rules={[{ required: true, message: 'Введите общее количество этажей' }]}>
                <InputNumber min={1} className="w-full" placeholder="Всего этажей" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="total_area" label="Общая площадь (м²)" rules={[{ required: true, message: 'Введите общую площадь' }]}>
                <InputNumber min={1} step={0.1} className="w-full" placeholder="Общая площадь" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="kitchen_area" label="Площадь кухни (м²)" rules={[{ required: true, message: 'Введите площадь кухни' }]}>
                <InputNumber min={1} step={0.1} className="w-full" placeholder="Площадь кухни" />
              </Form.Item>
            </Col>
          </Row>

          {/* Состояние */}
          <Row gutter={16}>
            <Col xs={24} sm={8}>
              <Form.Item name="condition_id" label="Состояние квартиры" rules={[{ required: true, message: 'Выберите состояние' }]}>
                <Select placeholder="Выберите состояние">
                  {(conditions?.data || conditions || []).map(condition => (
                    <Option key={condition.id} value={condition.id}>{condition.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="latitude" label="Широта">
                <InputNumber step={0.000001} className="w-full" placeholder="43.238949" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="longitude" label="Долгота">
                <InputNumber step={0.000001} className="w-full" placeholder="76.889709" />
              </Form.Item>
            </Col>
          </Row>

          {/* Типы аренды */}
          <Row gutter={16}>
            <Col xs={24} sm={8}>
              <Form.Item label="Типы аренды" rules={[{ validator: (_, value) => {
                const hourly = form.getFieldValue('rental_type_hourly');
                const daily = form.getFieldValue('rental_type_daily');
                if (!hourly && !daily) {
                  return Promise.reject('Выберите хотя бы один тип аренды');
                }
                return Promise.resolve();
              }}]}>
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
                </div>
              </Form.Item>
            </Col>
            <Col xs={24} sm={16}>
              <Form.Item name="listing_type" label="Тип листинга" rules={[{ required: true, message: 'Выберите тип листинга' }]}>
                <Select placeholder="Выберите тип листинга">
                  <Option value="owner">Владелец</Option>
                  <Option value="realtor">Риелтор</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          {/* Условное отображение полей цен */}
          {(showHourlyPrice || showDailyPrice) && (
            <Row gutter={16}>
              {showHourlyPrice && (
                <Col xs={24} sm={12}>
                  <Form.Item name="price" label="Цена за час (₸)" rules={[{ required: showHourlyPrice, message: 'Введите цену за час' }]}>
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                </Col>
              )}
              {showDailyPrice && (
                <Col xs={24} sm={12}>
                  <Form.Item name="daily_price" label="Цена за сутки (₸)" rules={[{ required: showDailyPrice, message: 'Введите цену за сутки' }]}>
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                </Col>
              )}
            </Row>
          )}

          <Form.Item name="residential_complex" label="Жилой комплекс">
            <Input placeholder="Название жилого комплекса" />
          </Form.Item>

          {/* Удобства */}
          <Form.Item name="amenity_ids" label="Удобства">
            <Select mode="multiple" placeholder="Выберите удобства">
              {amenities?.data?.map(amenity => (
                <Option key={amenity.id} value={amenity.id}>
                  {amenity.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          {/* Правила дома */}
          <Form.Item name="house_rule_ids" label="Правила дома">
            <Select mode="multiple" placeholder="Выберите правила">
              {houseRules?.data?.map(rule => (
                <Option key={rule.id} value={rule.id}>
                  {rule.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="description" label="Описание">
            <TextArea rows={4} placeholder="Опишите квартиру" />
          </Form.Item>

          {/* Управление фотографиями */}
          <Form.Item label="Фотографии">
            {/* Существующие фотографии */}
            {existingPhotos.length > 0 && (
              <div className="mb-4">
                <div className="text-sm font-medium mb-2">Текущие фотографии:</div>
                <div className={`grid gap-4 ${isMobile ? 'grid-cols-2' : 'grid-cols-2 md:grid-cols-4'}`}>
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
            
            <div className={`text-gray-500 mt-2 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              Поддерживаются форматы: JPG, PNG, GIF. Максимум 10 фотографий всего.
            </div>
          </Form.Item>
        </Form>
      </Modal>

      {/* Drawer деталей квартиры */}
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
                <Tag color={
                  selectedApartment.status === 'approved' ? 'green' :
                  selectedApartment.status === 'pending' ? 'orange' :
                  'red'
                }>
                  {selectedApartment.status === 'approved' ? 'Одобрена' :
                   selectedApartment.status === 'pending' ? 'На модерации' :
                   selectedApartment.status === 'rejected' ? 'Отклонена' :
                   selectedApartment.status === 'blocked' ? 'Заблокирована' :
                   'Неактивна'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Состояние">
                {selectedApartment.condition?.name || '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Доступна">
                <Tag color={selectedApartment.is_free ? 'green' : 'red'}>
                  {selectedApartment.is_free ? 'Да' : 'Нет'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Создано" span={isMobile ? 1 : 2}>
                <div className="font-mono text-sm">
                  {dayjs(selectedApartment.created_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedApartment.created_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Обновлено" span={isMobile ? 1 : 2}>
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
              {selectedApartment.id && (
                <Descriptions.Item label="Договор" span={isMobile ? 1 : 2}>
                  <Button 
                    type="primary"
                    size="large"
                    onClick={() => handleViewContract(selectedApartment.id)}
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

            {selectedApartment.house_rules && selectedApartment.house_rules.length > 0 && (
              <div className="mt-6">
                <Title level={4}>Правила дома</Title>
                <div className="flex flex-wrap gap-2">
                  {selectedApartment.house_rules.map((rule, index) => (
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
    </div>
  );
};

export default OwnerApartmentsPage; 