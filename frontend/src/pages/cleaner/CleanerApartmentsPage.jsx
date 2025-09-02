import {
  AreaChartOutlined,
  BuildOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  DollarOutlined,
  EnvironmentOutlined,
  FilterOutlined,
  HomeOutlined,
  InfoCircleOutlined,
  MailOutlined,
  PhoneOutlined,
  SafetyOutlined,
  SearchOutlined,
  UserOutlined
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import {
  Alert,
  Avatar,
  Badge,
  Card,
  Col,
  Descriptions,
  Empty,
  Image,
  Input,
  Row,
  Select,
  Spin,
  Statistic,
  Tag,
  Tooltip,
  Typography
} from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useState } from 'react';
import { cleanerAPI } from '../../lib/api.js';

const { Title, Text } = Typography;
const { Search } = Input;
const { Option } = Select;

const CleanerApartmentsPage = () => {
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [sortBy, setSortBy] = useState('address');
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получение квартир уборщицы
  const { data: apartments, isLoading } = useQuery({
    queryKey: ['cleaner-apartments'],
    queryFn: () => cleanerAPI.getApartments()
  });

  const getStatusIcon = (isFree) => {
    return isFree ? (
      <CheckCircleOutlined style={{ color: '#52c41a' }} />
    ) : (
      <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
    );
  };

  const getStatusText = (isFree) => {
    return isFree ? 'Свободна' : 'Занята';
  };

  const getStatusColor = (isFree) => {
    return isFree ? 'green' : 'red';
  };

  const getModerationStatusColor = (status) => {
    const colors = {
      'approved': 'success',
      'pending': 'processing', 
      'rejected': 'error',
      'draft': 'default'
    };
    return colors[status] || 'default';
  };

  const getModerationStatusText = (status) => {
    const texts = {
      'approved': 'Одобрена',
      'pending': 'На модерации',
      'rejected': 'Отклонена',
      'draft': 'Черновик'
    };
    return texts[status] || status;
  };

  // Фильтрация и сортировка квартир
  const apartmentData = apartments?.data || [];
  
  const filteredAndSortedApartments = React.useMemo(() => {
    let filtered = [...apartmentData];

    // Фильтрация по поиску
    if (searchText) {
      filtered = filtered.filter(apartment => 
        `${apartment.street} ${apartment.building} ${apartment.apartment_number || ''}`
          .toLowerCase()
          .includes(searchText.toLowerCase()) ||
        apartment.city?.name?.toLowerCase().includes(searchText.toLowerCase()) ||
        apartment.district?.name?.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    // Фильтрация по статусу
    if (statusFilter !== 'all') {
      if (statusFilter === 'free') {
        filtered = filtered.filter(apartment => apartment.is_free);
      } else if (statusFilter === 'occupied') {
        filtered = filtered.filter(apartment => !apartment.is_free);
      } else if (statusFilter === 'approved') {
        filtered = filtered.filter(apartment => apartment.status === 'approved');
      } else if (statusFilter === 'pending') {
        filtered = filtered.filter(apartment => apartment.status === 'pending');
      }
    }

    // Сортировка
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'address':
          return `${a.street} ${a.building}`.localeCompare(`${b.street} ${b.building}`);
        case 'price':
          return (b.daily_price || 0) - (a.daily_price || 0);
        case 'area':
          return (b.total_area || 0) - (a.total_area || 0);
        case 'created':
          return new Date(b.created_at) - new Date(a.created_at);
        default:
          return 0;
      }
    });

    return filtered;
  }, [apartmentData, searchText, statusFilter, sortBy]);

  // Статистика
  const stats = React.useMemo(() => {
    const total = apartmentData.length;
    const free = apartmentData.filter(apt => apt.is_free).length;
    const occupied = apartmentData.filter(apt => !apt.is_free).length;
    const approved = apartmentData.filter(apt => apt.status === 'approved').length;
    const pending = apartmentData.filter(apt => apt.status === 'pending').length;
    
    return { total, free, occupied, approved, pending };
  }, [apartmentData]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (!apartmentData || apartmentData.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <Title level={2}>Мои квартиры</Title>
        </div>
        <Card>
          <Empty 
            description="У вас нет назначенных квартир"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </Card>
      </div>
    );
  }

  return (
    <div className={`space-y-6 ${isMobile ? 'p-4' : 'p-6'}`}>
      {/* Заголовок и общая статистика */}
      <div className="flex flex-col space-y-4">
        <div className={`${isMobile ? 'space-y-4' : 'flex justify-between items-center'}`}>
          <div>
            <Title level={2} className={isMobile ? 'text-xl' : ''}>
              Мои квартиры
            </Title>
            <Text type="secondary">
              Управление назначенными квартирами
            </Text>
          </div>
          <div className={`${isMobile ? 'w-full' : ''}`}>
            <Alert
              message={`Всего: ${stats.total} квартир`}
              description={`Свободных: ${stats.free}, Занятых: ${stats.occupied}, Одобренных: ${stats.approved}`}
              type="info"
              showIcon
              icon={<InfoCircleOutlined />}
            />
          </div>
        </div>

        {/* Статистика в карточках */}
        <Row gutter={[8, 8]} className="mb-4">
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card className="text-center h-full" size="small">
              <Statistic
                title={<span className={isMobile ? 'text-xs' : 'text-sm'}>Всего</span>}
                value={stats.total}
                prefix={<HomeOutlined className="text-blue-500" />}
                valueStyle={{ 
                  color: '#1890ff', 
                  fontSize: isMobile ? '18px' : '24px',
                  fontWeight: 'bold'
                }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card className="text-center h-full" size="small">
              <Statistic
                title={<span className={isMobile ? 'text-xs' : 'text-sm'}>Свободных</span>}
                value={stats.free}
                prefix={<CheckCircleOutlined className="text-green-500" />}
                valueStyle={{ 
                  color: '#52c41a', 
                  fontSize: isMobile ? '18px' : '24px',
                  fontWeight: 'bold'
                }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card className="text-center h-full" size="small">
              <Statistic
                title={<span className={isMobile ? 'text-xs' : 'text-sm'}>Занятых</span>}
                value={stats.occupied}
                prefix={<CloseCircleOutlined className="text-red-500" />}
                valueStyle={{ 
                  color: '#ff4d4f', 
                  fontSize: isMobile ? '18px' : '24px',
                  fontWeight: 'bold'
                }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card className="text-center h-full" size="small">
              <Statistic
                title={<span className={isMobile ? 'text-xs' : 'text-sm'}>Одобренных</span>}
                value={stats.approved}
                prefix={<SafetyOutlined className="text-purple-500" />}
                valueStyle={{ 
                  color: '#722ed1', 
                  fontSize: isMobile ? '18px' : '24px',
                  fontWeight: 'bold'
                }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card className="text-center h-full" size="small">
              <Statistic
                title={<span className={isMobile ? 'text-xs' : 'text-sm'}>Посуточных</span>}
                value={apartmentData.filter(apt => apt.rental_type_daily).length}
                prefix={<CalendarOutlined className="text-orange-500" />}
                valueStyle={{ 
                  color: '#fa8c16', 
                  fontSize: isMobile ? '18px' : '24px',
                  fontWeight: 'bold'
                }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card className="text-center h-full" size="small">
              <Statistic
                title={<span className={isMobile ? 'text-xs' : 'text-sm'}>Почасовых</span>}
                value={apartmentData.filter(apt => apt.rental_type_hourly).length}
                prefix={<ClockCircleOutlined className="text-cyan-500" />}
                valueStyle={{ 
                  color: '#13c2c2', 
                  fontSize: isMobile ? '18px' : '24px',
                  fontWeight: 'bold'
                }}
              />
            </Card>
          </Col>
        </Row>
      </div>

      {/* Фильтры и поиск */}
      <Card>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={8} md={6}>
            <Search
              placeholder="Поиск по адресу..."
              allowClear
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              prefix={<SearchOutlined />}
            />
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Select
              value={statusFilter}
              onChange={setStatusFilter}
              className="w-full"
              prefix={<FilterOutlined />}
            >
              <Option value="all">Все статусы</Option>
              <Option value="free">Свободные</Option>
              <Option value="occupied">Занятые</Option>
              <Option value="approved">Одобренные</Option>
              <Option value="pending">На модерации</Option>
            </Select>
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Select
              value={sortBy}
              onChange={setSortBy}
              className="w-full"
            >
              <Option value="address">По адресу</Option>
              <Option value="price">По цене</Option>
              <Option value="area">По площади</Option>
              <Option value="created">По дате</Option>
            </Select>
          </Col>
          <Col xs={24} sm={24} md={10}>
            <Text type="secondary">
              Показано: {filteredAndSortedApartments.length} из {stats.total} квартир
            </Text>
          </Col>
        </Row>
      </Card>

      {/* Список квартир */}
      {filteredAndSortedApartments.length === 0 ? (
        <Card>
          <Empty 
            description="Квартиры не найдены"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </Card>
      ) : (
        <Row gutter={[16, 16]}>
          {filteredAndSortedApartments.map((apartment) => (
            <Col xs={24} sm={12} md={8} lg={6} xl={4} key={apartment.id}>
              <Card
                className="h-full hover:shadow-lg transition-shadow duration-300"
                bodyStyle={{ padding: isMobile ? '12px' : '24px' }}
                cover={
                  apartment.photos && apartment.photos.length > 0 ? (
                    <div className={`${isMobile ? 'h-32' : 'h-48'} bg-gray-200 overflow-hidden relative`}>
                      <Image
                        src={apartment.photos[0]?.url}
                        alt="Квартира"
                        className="w-full h-full object-cover"
                        fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMIAAADDCAYAAADQvc6UAAABRWlDQ1BJQ0MgUHJvZmlsZQAAKJFjYGASSSwoyGFhYGDIzSspCnJ3UoiIjFJgf8LAwSDCIMogwMCcmFxc4BgQ4ANUwgCjUcG3awyMIPqyLsis7PPOq3QdDFcvjV3jOD1boQVTPQrgSkktTgbSf4A4LbmgqISBgTEFyFYuLykAsTuAbJEioKOA7DkgdjqEvQHEToKwj4DVhAQ5A9k3gGyB5IxEoBmML4BsnSQk8XQkNtReEOBxcfXxUQg1Mjc0dyHgXNJBSWpFCYh2zi+oLMpMzyhRcASGUqqCZ16yno6CkYGRAQMDKMwhqj/fAIcloxgHQqxAjIHBEugw5sUIsSQpBobtQPdLciLEVJYzMPBHMDBsayhILEqEO4DxG0txmrERhM29nYGBddr//5/DGRjYNRkY/l7////39v///y4Dmn+LgeHANwDrkl1AuO+pmgAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAwqADAAQAAAABAAAAwwAAAAD9b/HnAAAHlklEQVR4Ae3dP3Ik1RnG4W+FgYxN"
                      />
                      {/* Статус занятости в углу */}
                      <div className="absolute top-1 right-1">
                        <Tag 
                          color={getStatusColor(apartment.is_free)} 
                          className={`border-0 shadow-sm ${isMobile ? 'text-xs px-1' : ''}`}
                        >
                          {getStatusText(apartment.is_free)}
                        </Tag>
                      </div>
                    </div>
                  ) : (
                    <div className={`${isMobile ? 'h-32' : 'h-48'} bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center relative`}>
                      <HomeOutlined className={`${isMobile ? 'text-4xl' : 'text-6xl'} text-blue-300`} />
                      <div className="absolute top-1 right-1">
                        <Tag 
                          color={getStatusColor(apartment.is_free)} 
                          className={`border-0 shadow-sm ${isMobile ? 'text-xs px-1' : ''}`}
                        >
                          {getStatusText(apartment.is_free)}
                        </Tag>
                      </div>
                    </div>
                  )
                }
              >
                <div className={`${isMobile ? 'space-y-2' : 'space-y-4'}`}>
                  {/* Заголовок с адресом */}
                  <div>
                    <Title level={isMobile ? 5 : 4} className="!mb-1 !leading-tight">
                      <EnvironmentOutlined className="text-blue-500 mr-1" />
                      <span className={isMobile ? 'text-sm' : ''}>
                        {apartment.street}, д. {apartment.building}
                      </span>
                    </Title>
                    {apartment.apartment_number && (
                      <Text type="secondary" className={isMobile ? 'text-xs' : 'text-sm'}>
                        Квартира {apartment.apartment_number}
                      </Text>
                    )}
                    <br />
                    <Text type="secondary" className={isMobile ? 'text-xs' : 'text-sm'}>
                      {apartment.city?.name}, {apartment.district?.name}
                    </Text>
                  </div>

                  {/* Основные характеристики */}
                  <Row gutter={[4, 4]}>
                    <Col span={12}>
                      <div className={`text-center ${isMobile ? 'p-1' : 'p-2'} bg-gray-50 rounded`}>
                        <BuildOutlined className="text-orange-500" />
                        <div className={`${isMobile ? 'text-xs' : 'text-xs'} mt-1`}>
                          {apartment.room_count || '?'} комн.
                        </div>
                      </div>
                    </Col>
                    <Col span={12}>
                      <div className={`text-center ${isMobile ? 'p-1' : 'p-2'} bg-gray-50 rounded`}>
                        <AreaChartOutlined className="text-green-500" />
                        <div className={`${isMobile ? 'text-xs' : 'text-xs'} mt-1`}>
                          {apartment.total_area || '?'} м²
                        </div>
                      </div>
                    </Col>
                  </Row>

                  {/* Цена */}
                  <div className={`text-center ${isMobile ? 'p-2' : 'p-3'} bg-blue-50 rounded-lg`}>
                    <DollarOutlined className={`text-blue-600 ${isMobile ? 'text-base' : 'text-lg'}`} />
                    <div className={`font-semibold text-blue-600 ${isMobile ? 'text-base' : 'text-lg'}`}>
                      {apartment.daily_price?.toLocaleString() || '0'} ₸/день
                    </div>
                    {apartment.price > 0 && (
                      <div className="text-xs text-gray-500">
                        {apartment.price?.toLocaleString()} ₸/час
                      </div>
                    )}
                    {apartment.rental_type_hourly && apartment.rental_type_daily && (
                      <div className="text-xs text-blue-500 mt-1">
                        Почасовая и посуточная аренда
                      </div>
                    )}
                  </div>

                  {/* Статус модерации */}
                  <div className="text-center">
                    <Badge
                      status={getModerationStatusColor(apartment.status)}
                      text={getModerationStatusText(apartment.status)}
                    />
                    {apartment.moderator_comment && (
                      <Tooltip title={`Комментарий модератора: ${apartment.moderator_comment}`}>
                        <InfoCircleOutlined className="ml-2 text-gray-400 cursor-help" />
                      </Tooltip>
                    )}
                  </div>

                  {/* Описание квартиры */}
                  {apartment.description && (
                    <div className="bg-gray-50 p-2 rounded text-xs">
                      <Text strong className="block mb-1">Описание:</Text>
                      <Text className="text-gray-600">
                        {apartment.description.length > 100 
                          ? `${apartment.description.substring(0, 100)}...` 
                          : apartment.description}
                      </Text>
                    </div>
                  )}

                  {/* Дополнительная информация */}
                  <div className={`space-y-1 ${isMobile ? 'text-xs' : 'text-sm'}`}>
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <Text type="secondary" className="text-xs">Состояние:</Text>
                        <div className={isMobile ? 'text-xs' : 'text-sm'}>
                          {apartment.condition?.name || 'Не указано'}
                        </div>
                      </div>
                      <div>
                        <Text type="secondary" className="text-xs">Этаж:</Text>
                        <div className={isMobile ? 'text-xs' : 'text-sm'}>
                          {apartment.floor ? `${apartment.floor}/${apartment.total_floors || '?'}` : 'Не указан'}
                        </div>
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <Text type="secondary" className="text-xs">Кухня:</Text>
                        <div className={isMobile ? 'text-xs' : 'text-sm'}>
                          {apartment.kitchen_area ? `${apartment.kitchen_area} м²` : 'Не указана'}
                        </div>
                      </div>
                      <div>
                        <Text type="secondary" className="text-xs">Просмотры:</Text>
                        <div className={isMobile ? 'text-xs' : 'text-sm'}>
                          {apartment.view_count || 0}
                        </div>
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <Text type="secondary" className="text-xs">Бронирования:</Text>
                        <div className={isMobile ? 'text-xs' : 'text-sm'}>
                          {apartment.booking_count || 0}
                        </div>
                      </div>
                      <div>
                        <Text type="secondary" className="text-xs">Комиссия:</Text>
                        <div className={isMobile ? 'text-xs' : 'text-sm'}>
                          {apartment.service_fee_percentage || 0}%
                        </div>
                      </div>
                    </div>
                    
                    {!isMobile && (
                      <>
                        {apartment.microdistrict?.name && (
                          <div>
                            <Text type="secondary" className="text-xs">Микрорайон:</Text>
                            <div className="text-sm">{apartment.microdistrict.name}</div>
                          </div>
                        )}
                        <div>
                          <Text type="secondary" className="text-xs">Тип аренды:</Text>
                          <div className="text-sm">
                            {apartment.rental_type_hourly && apartment.rental_type_daily ? 'Почасовая и посуточная' :
                             apartment.rental_type_daily ? 'Посуточная' :
                             apartment.rental_type_hourly ? 'Почасовая' : 'Не указан'}
                          </div>
                        </div>
                        {apartment.contract_id && (
                          <div>
                            <Text type="secondary" className="text-xs">Договор:</Text>
                            <div className="text-sm">ID: {apartment.contract_id}</div>
                          </div>
                        )}
                        <div>
                          <Text type="secondary" className="text-xs">Тип размещения:</Text>
                          <div className="text-sm">
                            {apartment.listing_type === 'owner' ? 'Владелец' : 'Агент'}
                          </div>
                        </div>
                      </>
                    )}
                  </div>

                  {/* Удобства */}
                  {apartment.amenities && apartment.amenities.length > 0 && (
                    <div>
                      <Text strong className="text-xs block mb-2">Удобства:</Text>
                      <div className="flex flex-wrap gap-1">
                        {apartment.amenities.slice(0, 3).map((amenity, index) => (
                          <Tag key={index} size="small" color="blue">
                            {amenity.name}
                          </Tag>
                        ))}
                        {apartment.amenities.length > 3 && (
                          <Tooltip title={apartment.amenities.slice(3).map(a => a.name).join(', ')}>
                            <Tag size="small" color="geekblue">
                              +{apartment.amenities.length - 3}
                            </Tag>
                          </Tooltip>
                        )}
                      </div>
                    </div>
                  )}

                  {/* Информация о владельце */}
                  {apartment.owner?.user && (
                    <div className="border-t pt-3">
                      <div className="flex items-start space-x-2">
                        <Avatar 
                          size={isMobile ? "small" : "default"} 
                          icon={<UserOutlined />} 
                          className="bg-purple-500 flex-shrink-0" 
                        />
                        <div className="flex-1 min-w-0">
                          <div className={`${isMobile ? 'text-xs' : 'text-sm'} font-medium truncate`}>
                            {apartment.owner.user.first_name} {apartment.owner.user.last_name}
                          </div>
                          <div className={`${isMobile ? 'text-xs' : 'text-sm'} text-gray-500 flex items-center`}>
                            <PhoneOutlined className="mr-1 flex-shrink-0" />
                            <span className="truncate">{apartment.owner.user.phone}</span>
                          </div>
                          {apartment.owner.user.email && !isMobile && (
                            <div className="text-xs text-gray-500 flex items-center">
                              <MailOutlined className="mr-1 flex-shrink-0" />
                              <span className="truncate">{apartment.owner.user.email}</span>
                            </div>
                          )}
                          {apartment.owner.user.iin && !isMobile && (
                            <div className="text-xs text-gray-400 truncate">
                              ИИН: {apartment.owner.user.iin}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Дата создания */}
                  <div className="text-center pt-2 border-t">
                    <Text type="secondary" className="text-xs">
                      <CalendarOutlined className="mr-1" />
                      Добавлена: {dayjs(apartment.created_at).format('DD.MM.YYYY')}
                    </Text>
                  </div>
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      )}
    </div>
  );
};

export default CleanerApartmentsPage;
