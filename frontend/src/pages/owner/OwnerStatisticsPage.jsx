import React, { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Card, Row, Col, Statistic, Select, DatePicker, Typography, Table,
  Space, Progress, Tag, Divider, Alert
} from 'antd';
import {
  DollarOutlined, CalendarOutlined, HomeOutlined, RiseOutlined,
  PercentageOutlined, ClockCircleOutlined, StarOutlined, UserOutlined
} from '@ant-design/icons';
import {
  LineChart, Line, AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer
} from 'recharts';
import { apartmentsAPI, bookingsAPI } from '../../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { RangePicker } = DatePicker;

const OwnerStatisticsPage = () => {
  const [dateRange, setDateRange] = useState([
    dayjs().subtract(1, 'month'),
    dayjs()
  ]);
  const [selectedApartment, setSelectedApartment] = useState(null);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получение квартир владельца
  const { data: apartmentsData } = useQuery({
    queryKey: ['owner-apartments'],
    queryFn: () => apartmentsAPI.getMyApartments()
  });

  // Получение статистики доходов
  const { data: revenueData, isLoading: revenueLoading } = useQuery({
    queryKey: ['owner-revenue-stats', dateRange, selectedApartment],
    queryFn: () => {
      // Здесь должен быть endpoint для получения статистики доходов
      // Пока используем заглушку
      return Promise.resolve({
        total_revenue: 850000,
        period_revenue: 125000,
        growth_percentage: 15.3,
        daily_revenue: generateDailyRevenue(),
        monthly_revenue: generateMonthlyRevenue(),
        apartment_revenue: generateApartmentRevenue(),
        occupancy_rate: 78.5,
        average_booking_value: 25000,
        top_apartments: generateTopApartments()
      });
    }
  });

  // Получение статистики бронирований
  const { data: bookingStats } = useQuery({
    queryKey: ['owner-booking-stats', dateRange, selectedApartment],
    queryFn: () => {
      return Promise.resolve({
        total_bookings: 45,
        completed_bookings: 38,
        cancelled_bookings: 4,
        pending_bookings: 3,
        booking_trends: generateBookingTrends(),
        status_distribution: [
          { name: 'Завершено', value: 38, color: '#52c41a' },
          { name: 'Активные', value: 3, color: '#1890ff' },
          { name: 'Отменено', value: 4, color: '#ff4d4f' }
        ]
      });
    }
  });

  // Функции для генерации мок данных
  function generateDailyRevenue() {
    const data = [];
    for (let i = 30; i >= 0; i--) {
      const date = dayjs().subtract(i, 'day');
      data.push({
        date: date.format('DD.MM'),
        revenue: Math.floor(Math.random() * 15000) + 5000,
        bookings: Math.floor(Math.random() * 5) + 1
      });
    }
    return data;
  }

  function generateMonthlyRevenue() {
    const data = [];
    for (let i = 11; i >= 0; i--) {
      const date = dayjs().subtract(i, 'month');
      data.push({
        month: date.format('MMM YYYY'),
        revenue: Math.floor(Math.random() * 100000) + 50000,
        bookings: Math.floor(Math.random() * 20) + 10,
        occupancy: Math.floor(Math.random() * 30) + 60
      });
    }
    return data;
  }

  function generateApartmentRevenue() {
    return apartmentsData?.apartments?.map((apt, index) => ({
      id: apt.id,
      name: apt.title,
      revenue: Math.floor(Math.random() * 50000) + 20000,
      bookings: Math.floor(Math.random() * 10) + 5,
      occupancy: Math.floor(Math.random() * 40) + 50,
      rating: (Math.random() * 2 + 3).toFixed(1)
    })) || [];
  }

  function generateTopApartments() {
    return apartmentsData?.apartments?.slice(0, 5).map((apt, index) => ({
      id: apt.id,
      name: apt.title,
      revenue: Math.floor(Math.random() * 50000) + 30000,
      bookings: Math.floor(Math.random() * 8) + 8,
      growth: (Math.random() * 40 - 10).toFixed(1)
    })) || [];
  }

  function generateBookingTrends() {
    const data = [];
    for (let i = 6; i >= 0; i--) {
      const date = dayjs().subtract(i, 'month');
      data.push({
        month: date.format('MMM'),
        bookings: Math.floor(Math.random() * 15) + 5,
        revenue: Math.floor(Math.random() * 80000) + 40000
      });
    }
    return data;
  }

  const apartmentColumns = [
    {
      title: 'Квартира',
      dataIndex: 'name',
      key: 'name',
      render: (name) => <Text strong>{name}</Text>
    },
    {
      title: 'Доход',
      dataIndex: 'revenue',
      key: 'revenue',
      render: (revenue) => (
        <Text strong style={{ color: '#52c41a' }}>
          {revenue?.toLocaleString()} ₸
        </Text>
      ),
      sorter: (a, b) => a.revenue - b.revenue
    },
    {
      title: 'Бронирований',
      dataIndex: 'bookings',
      key: 'bookings',
      render: (bookings) => <Text>{bookings}</Text>,
      sorter: (a, b) => a.bookings - b.bookings
    },
    {
      title: 'Загруженность',
      dataIndex: 'occupancy',
      key: 'occupancy',
      render: (occupancy) => (
        <Progress
          percent={occupancy}
          size="small"
          status={occupancy > 70 ? 'success' : occupancy > 50 ? 'normal' : 'exception'}
        />
      ),
      sorter: (a, b) => a.occupancy - b.occupancy
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating) => (
        <div className="flex items-center space-x-1">
          <StarOutlined style={{ color: '#faad14' }} />
          <Text>{rating}</Text>
        </div>
      ),
      sorter: (a, b) => parseFloat(a.rating) - parseFloat(b.rating)
    }
  ];

  return (
    <div className={`space-y-6 ${isMobile ? 'p-4' : 'p-6'}`}>
      {/* Заголовок и фильтры */}
      <div className={`${isMobile ? 'space-y-4' : 'flex flex-col lg:flex-row lg:justify-between lg:items-center space-y-4 lg:space-y-0'}`}>
        <div>
          <Title level={2} className={isMobile ? 'text-xl' : ''}>Статистика и аналитика</Title>
          <Text type="secondary" className={isMobile ? 'text-sm' : ''}>
            Подробная аналитика доходов и эффективности
          </Text>
        </div>
        
        <Space wrap direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
          <Select
            value={selectedApartment}
            onChange={setSelectedApartment}
            placeholder="Выберите квартиру"
            style={{ width: isMobile ? '100%' : 200 }}
            allowClear
          >
            <Option value={null}>Все квартиры</Option>
            {(apartmentsData?.data?.apartments || []).map(apartment => (
              <Option key={apartment.id} value={apartment.id}>
                {apartment.title || `${apartment.street}, ${apartment.building}`}
              </Option>
            ))}
          </Select>
          
          <RangePicker
            value={dateRange}
            onChange={setDateRange}
            format="DD.MM.YYYY"
            style={{ width: isMobile ? '100%' : 'auto' }}
          />
        </Space>
      </div>

      {/* Основная статистика */}
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Общий доход"
              value={revenueData?.total_revenue || 0}
              prefix={<DollarOutlined className="text-green-500" />}
              suffix="₸"
              precision={0}
              valueStyle={{ 
                color: '#52c41a',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Progress 
              percent={75} 
              showInfo={false} 
              strokeColor="#52c41a" 
              size={isMobile ? 'small' : 'default'}
            />
            <Text className={`text-green-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              +{revenueData?.growth_percentage || 0}% к прошлому периоду
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Заполняемость"
              value={revenueData?.occupancy_rate || 0}
              prefix={<PercentageOutlined className="text-blue-500" />}
              suffix="%"
              precision={1}
              valueStyle={{ 
                color: '#1890ff',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Progress 
              percent={revenueData?.occupancy_rate || 0} 
              showInfo={false} 
              strokeColor="#1890ff"
              size={isMobile ? 'small' : 'default'}
            />
            <Text className={`text-blue-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              Средняя заполняемость
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Всего бронирований"
              value={bookingStats?.total_bookings || 0}
              prefix={<CalendarOutlined className="text-purple-500" />}
              valueStyle={{ 
                color: '#722ed1',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Progress 
              percent={80} 
              showInfo={false} 
              strokeColor="#722ed1"
              size={isMobile ? 'small' : 'default'}
            />
            <Text className={`text-purple-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              За выбранный период
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Средний чек"
              value={revenueData?.average_booking_value || 0}
              prefix={<DollarOutlined className="text-orange-500" />}
              suffix="₸"
              precision={0}
              valueStyle={{ 
                color: '#fa8c16',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Progress 
              percent={60} 
              showInfo={false} 
              strokeColor="#fa8c16"
              size={isMobile ? 'small' : 'default'}
            />
            <Text className={`text-orange-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              Средняя стоимость бронирования
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Активные квартиры"
              value={apartmentsData?.data?.apartments?.length || 0}
              prefix={<HomeOutlined className="text-cyan-500" />}
              valueStyle={{ 
                color: '#13c2c2',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Progress 
              percent={90} 
              showInfo={false} 
              strokeColor="#13c2c2"
              size={isMobile ? 'small' : 'default'}
            />
            <Text className={`text-cyan-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              Количество активных объектов
            </Text>
          </Card>
        </Col>
      </Row>

      {/* Графики */}
      <Row gutter={[16, 16]}>
        {/* График доходов */}
        <Col xs={24} lg={16}>
          <Card title="Динамика доходов" className={isMobile ? 'h-80' : 'h-96'}>
            <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
              <AreaChart data={revenueData?.daily_revenue || []}>
                <defs>
                  <linearGradient id="colorRevenue" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#52c41a" stopOpacity={0.8}/>
                    <stop offset="95%" stopColor="#52c41a" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <XAxis 
                  dataKey="date" 
                  fontSize={isMobile ? 10 : 12}
                  interval={isMobile ? 'preserveStartEnd' : 0}
                />
                <YAxis fontSize={isMobile ? 10 : 12} />
                <CartesianGrid strokeDasharray="3 3" />
                <Tooltip 
                  formatter={(value) => [`${value.toLocaleString()} ₸`, 'Доход']}
                  labelStyle={{ fontSize: isMobile ? '12px' : '14px' }}
                />
                <Area 
                  type="monotone" 
                  dataKey="revenue" 
                  stroke="#52c41a" 
                  fillOpacity={1} 
                  fill="url(#colorRevenue)" 
                />
              </AreaChart>
            </ResponsiveContainer>
          </Card>
        </Col>

        {/* Распределение по статусам */}
        <Col xs={24} lg={8}>
          <Card title="Статусы бронирований" className={isMobile ? 'h-80' : 'h-96'}>
            <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
              <PieChart>
                <Pie
                  data={bookingStats?.status_distribution || []}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => isMobile ? `${(percent * 100).toFixed(0)}%` : `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={isMobile ? 70 : 80}
                  fill="#8884d8"
                  dataKey="value"
                  fontSize={isMobile ? 10 : 12}
                >
                  {(bookingStats?.status_distribution || []).map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
                {!isMobile && <Legend />}
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      {/* Месячная динамика */}
      <Row gutter={[16, 16]}>
        <Col xs={24}>
          <Card title="Месячная динамика доходов" className={isMobile ? 'h-80' : 'h-96'}>
            <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
              <BarChart data={revenueData?.monthly_revenue || []}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="month" 
                  fontSize={isMobile ? 10 : 12}
                  angle={isMobile ? -45 : 0}
                  textAnchor={isMobile ? 'end' : 'middle'}
                  height={isMobile ? 60 : 30}
                />
                <YAxis fontSize={isMobile ? 10 : 12} />
                <Tooltip 
                  formatter={(value, name) => [
                    `${value.toLocaleString()} ₸`, 
                    name === 'revenue' ? 'Доход' : name === 'bookings' ? 'Бронирования' : 'Заполняемость %'
                  ]}
                />
                <Legend />
                <Bar dataKey="revenue" fill="#52c41a" name="Доход" />
                <Bar dataKey="bookings" fill="#1890ff" name="Бронирования" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      {/* Топ квартир */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="Топ квартир по доходности">
            <div className="space-y-3">
              {(revenueData?.top_apartments || []).map((apartment, index) => (
                <div key={index} className={`flex items-center justify-between p-3 bg-gray-50 rounded-lg ${isMobile ? 'flex-col space-y-2' : ''}`}>
                  <div className={`flex items-center space-x-3 ${isMobile ? 'w-full' : ''}`}>
                    <div className={`bg-blue-100 rounded-lg flex items-center justify-center ${isMobile ? 'w-8 h-8' : 'w-10 h-10'}`}>
                      <span className={`text-blue-600 font-bold ${isMobile ? 'text-sm' : ''}`}>#{index + 1}</span>
                    </div>
                    <div className={isMobile ? 'flex-1' : ''}>
                      <div className={`font-medium ${isMobile ? 'text-sm' : ''}`}>{apartment.name}</div>
                      <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
                        {apartment.bookings} бронирований
                      </div>
                    </div>
                  </div>
                  <div className={`text-right ${isMobile ? 'w-full text-center' : ''}`}>
                    <div className={`font-bold text-green-600 ${isMobile ? 'text-base' : 'text-lg'}`}>
                      {apartment.revenue.toLocaleString()} ₸
                    </div>
                    <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
                      Заполняемость {apartment.occupancy}%
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="Ключевые показатели">
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className={isMobile ? 'text-sm' : ''}>Средняя длительность аренды</span>
                <span className={`font-semibold ${isMobile ? 'text-sm' : ''}`}>3.2 дня</span>
              </div>
              <Divider className="my-2" />
              
              <div className="flex justify-between items-center">
                <span className={isMobile ? 'text-sm' : ''}>Коэффициент конверсии</span>
                <span className={`font-semibold text-green-600 ${isMobile ? 'text-sm' : ''}`}>87%</span>
              </div>
              <Progress percent={87} strokeColor="#52c41a" size={isMobile ? 'small' : 'default'} />
              
              <div className="flex justify-between items-center">
                <span className={isMobile ? 'text-sm' : ''}>Время отклика</span>
                <span className={`font-semibold ${isMobile ? 'text-sm' : ''}`}>1.2 часа</span>
              </div>
              <Progress percent={95} strokeColor="#1890ff" size={isMobile ? 'small' : 'default'} />
              
              <div className="flex justify-between items-center">
                <span className={isMobile ? 'text-sm' : ''}>Повторные бронирования</span>
                <span className={`font-semibold text-purple-600 ${isMobile ? 'text-sm' : ''}`}>23%</span>
              </div>
              <Progress percent={23} strokeColor="#722ed1" size={isMobile ? 'small' : 'default'} />
              
              <div className="flex justify-between items-center">
                <span className={isMobile ? 'text-sm' : ''}>Отмены после подтверждения</span>
                <span className={`font-semibold text-orange-600 ${isMobile ? 'text-sm' : ''}`}>5%</span>
              </div>
              <Progress percent={5} strokeColor="#fa8c16" size={isMobile ? 'small' : 'default'} />
            </div>
          </Card>
        </Col>
      </Row>

      {/* Рекомендации */}
      <Row gutter={[16, 16]}>
        <Col xs={24}>
          <Card title="Рекомендации для улучшения">
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12} md={8}>
                <Alert
                  message="Повысьте цены"
                  description="Ваша заполняемость выше 80%. Рассмотрите возможность повышения цен на 10-15%."
                  type="success"
                  showIcon
                  className={isMobile ? 'text-sm' : ''}
                />
              </Col>
              <Col xs={24} sm={12} md={8}>
                <Alert
                  message="Улучшите фото"
                  description="Квартиры с профессиональными фото получают на 40% больше бронирований."
                  type="info"
                  showIcon
                  className={isMobile ? 'text-sm' : ''}
                />
              </Col>
              <Col xs={24} sm={12} md={8}>
                <Alert
                  message="Быстрее отвечайте"
                  description="Сократите время ответа до 30 минут для увеличения конверсии."
                  type="warning"
                  showIcon
                  className={isMobile ? 'text-sm' : ''}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default OwnerStatisticsPage; 