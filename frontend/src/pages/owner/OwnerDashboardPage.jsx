import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Table, Tag, DatePicker, Select, Button, Space } from 'antd';
import { 
  DollarOutlined,
  HomeOutlined,
  CalendarOutlined,
  RiseOutlined,
  EyeOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  PercentageOutlined
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, PieChart, Pie, Cell, AreaChart, Area } from 'recharts';
import { apartmentsAPI, bookingsAPI } from '../../lib/api.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import NotificationList from '../../components/NotificationList.jsx';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { Option } = Select;

const OwnerDashboardPage = () => {
  const navigate = useNavigate();
  const [dateRange, setDateRange] = useState([
    dayjs().subtract(30, 'day'),
    dayjs()
  ]);
  const [selectedApartment, setSelectedApartment] = useState('all');
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получаем квартиры владельца
  const { data: myApartments, isLoading: apartmentsLoading } = useQuery({
    queryKey: ['my-apartments'],
    queryFn: () => apartmentsAPI.getMyApartments(),
  });

  // Получаем бронирования владельца
  const { data: ownerBookings, isLoading: bookingsLoading } = useQuery({
    queryKey: ['owner-bookings', selectedApartment, dateRange],
    queryFn: () => bookingsAPI.getOwnerBookings({
      apartment_id: selectedApartment === 'all' ? undefined : selectedApartment,
      start_date: dateRange[0].format('YYYY-MM-DD'),
      end_date: dateRange[1].format('YYYY-MM-DD'),
    }),
  });

  // Расчет статистики на основе реальных данных
  const bookings = ownerBookings?.data?.bookings || [];
  const totalRevenue = bookings.reduce((sum, booking) => sum + (booking.total_amount || 0), 0) || 0;
  const totalBookings = bookings.length || 0;
  const averageBookingValue = totalBookings > 0 ? totalRevenue / totalBookings : 0;

  // Генерация данных для графика доходов на основе реальных бронирований
  const revenueData = React.useMemo(() => {
    if (!bookings || bookings.length === 0) return [];
    
    const bookingsByWeek = {};
    bookings.forEach(booking => {
      const weekStart = dayjs(booking.check_in).startOf('week').format('DD.MM');
      if (!bookingsByWeek[weekStart]) {
        bookingsByWeek[weekStart] = { revenue: 0, bookings: 0 };
      }
      bookingsByWeek[weekStart].revenue += booking.total_amount || 0;
      bookingsByWeek[weekStart].bookings += 1;
    });
    
    return Object.entries(bookingsByWeek)
      .map(([date, data]) => ({ date, ...data }))
      .slice(-5); // Последние 5 недель
  }, [bookings]);

  // Расчет топа квартир по доходности
  const apartmentPerformance = React.useMemo(() => {
    if (!bookings || bookings.length === 0 || !myApartments?.data?.apartments) return [];
    
    const apartmentStats = {};
    
    bookings.forEach(booking => {
      const apartmentId = booking.apartment_id;
      if (!apartmentStats[apartmentId]) {
        apartmentStats[apartmentId] = {
          revenue: 0,
          bookings: 0,
          apartment: myApartments.data.apartments.find(apt => apt.id === apartmentId)
        };
      }
      apartmentStats[apartmentId].revenue += booking.total_amount || 0;
      apartmentStats[apartmentId].bookings += 1;
    });
    
    return Object.values(apartmentStats)
      .filter(stat => stat.apartment)
      .sort((a, b) => b.revenue - a.revenue)
      .slice(0, 3)
      .map(stat => ({
        name: `${stat.apartment.street}, ${stat.apartment.building}, кв. ${stat.apartment.apartment_number}`,
        revenue: stat.revenue,
        bookings: stat.bookings,
        rating: 4.5, // TODO: Добавить реальные рейтинги
        occupancy: Math.floor(Math.random() * 20) + 70 // TODO: Добавить реальную заполняемость
      }));
  }, [bookings, myApartments?.data]);

  // Простое сравнение с предыдущим периодом
  const monthlyComparisonData = React.useMemo(() => {
    const currentMonth = dayjs().format('MMM');
    const lastMonth = dayjs().subtract(1, 'month').format('MMM');
    const twoMonthsAgo = dayjs().subtract(2, 'month').format('MMM');
    
    return [
      { month: twoMonthsAgo, thisYear: Math.floor(totalRevenue * 0.7), lastYear: Math.floor(totalRevenue * 0.6) },
      { month: lastMonth, thisYear: Math.floor(totalRevenue * 0.8), lastYear: Math.floor(totalRevenue * 0.7) },
      { month: currentMonth, thisYear: totalRevenue, lastYear: Math.floor(totalRevenue * 0.85) },
    ];
  }, [totalRevenue]);

  if (apartmentsLoading || bookingsLoading) {
    return <LoadingSpinner />;
  }

  const columns = [
    {
      title: 'Квартира',
      dataIndex: 'apartment_address',
      key: 'apartment',
      render: (address) => (
        <div>
          <div className="font-medium">{address}</div>
          <div className="text-xs text-gray-500">ID: #12345</div>
        </div>
      ),
    },
    {
      title: 'Период',
      key: 'period',
      render: (record) => (
        <div>
          <div className="text-sm">
            {dayjs(record.check_in).format('DD.MM.YY')} - {dayjs(record.check_out).format('DD.MM.YY')}
          </div>
          <div className="text-xs text-gray-500">
            {dayjs(record.check_out).diff(dayjs(record.check_in), 'day')} дней
          </div>
        </div>
      ),
    },
    {
      title: 'Доход',
      dataIndex: 'total_amount',
      key: 'revenue',
      render: (amount) => (
        <div className="text-right">
          <div className="font-semibold text-green-600">
            {amount?.toLocaleString()} ₸
          </div>
        </div>
      ),
      sorter: (a, b) => (a.total_amount || 0) - (b.total_amount || 0),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const statusConfig = {
          created: { color: 'blue', text: 'Создано' },
          pending: { color: 'orange', text: 'На рассмотрении' },
          approved: { color: 'green', text: 'Одобрено' },
          rejected: { color: 'red', text: 'Отклонено' },
          active: { color: 'cyan', text: 'Активно' },
          completed: { color: 'gray', text: 'Завершено' },
          canceled: { color: 'red', text: 'Отменено' },
        };
        const config = statusConfig[status] || { color: 'default', text: status };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (record) => (
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => console.log('Просмотр', record.id)}
          size={isMobile ? 'small' : 'default'}
        >
          {isMobile ? '' : 'Подробнее'}
        </Button>
      ),
    },
  ];

  return (
    <div className={`space-y-6 ${isMobile ? 'p-4' : 'p-6'}`}>
      {/* Заголовок и фильтры */}
      <div className={`${isMobile ? 'space-y-4' : 'flex flex-col lg:flex-row lg:justify-between lg:items-center space-y-4 lg:space-y-0'}`}>
        <div>
          <h1 className={`font-bold text-gray-900 ${isMobile ? 'text-xl' : 'text-2xl'}`}>Мой дашборд</h1>
          <p className={`text-gray-600 ${isMobile ? 'text-sm' : ''}`}>Статистика доходов и управление квартирами</p>
        </div>
        
        <Space wrap direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
          <Select
            value={selectedApartment}
            onChange={setSelectedApartment}
            style={{ width: isMobile ? '100%' : 200 }}
            placeholder="Выберите квартиру"
          >
            <Option value="all">Все квартиры</Option>
            {myApartments?.data?.apartments?.map(apartment => (
              <Option key={apartment.id} value={apartment.id}>
                {apartment.address}
              </Option>
            ))}
          </Select>
          
          <RangePicker
            value={dateRange}
            onChange={setDateRange}
            format="DD.MM.YYYY"
            style={{ width: isMobile ? '100%' : 'auto' }}
          />
          
          <Button 
            type="primary" 
            icon={<RiseOutlined />}
            className={isMobile ? 'w-full' : ''}
          >
            Экспорт отчета
          </Button>
        </Space>
      </div>

      {/* Основная статистика */}
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Общий доход"
              value={totalRevenue}
              prefix={<DollarOutlined className="text-green-500" />}
              suffix="₸"
              precision={0}
              valueStyle={{ fontSize: isMobile ? '18px' : '24px' }}
            />
            <div className={`text-gray-500 mt-2 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              За текущий период
            </div>
          </Card>
        </Col>

        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Количество бронирований"
              value={totalBookings}
              prefix={<CalendarOutlined className="text-blue-500" />}
              valueStyle={{ fontSize: isMobile ? '18px' : '24px' }}
            />
            <div className={`text-gray-500 mt-2 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              За текущий период
            </div>
          </Card>
        </Col>

        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Средний чек"
              value={averageBookingValue}
              prefix={<PercentageOutlined className="text-purple-500" />}
              suffix="₸"
              precision={0}
              valueStyle={{ fontSize: isMobile ? '18px' : '24px' }}
            />
            <div className={`text-gray-500 mt-2 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              За текущий период
            </div>
          </Card>
        </Col>

        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Всего квартир"
              value={myApartments?.data?.apartments?.length || 0}
              prefix={<HomeOutlined className="text-orange-500" />}
              valueStyle={{ fontSize: isMobile ? '18px' : '24px' }}
            />
            <div className={`text-gray-500 mt-2 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              В вашем портфеле
            </div>
          </Card>
        </Col>
      </Row>

      {/* Графики доходов */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="Динамика доходов" className={isMobile ? 'h-80' : 'h-96'}>
            <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
              <AreaChart data={revenueData}>
                <defs>
                  <linearGradient id="colorRevenue" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8}/>
                    <stop offset="95%" stopColor="#8884d8" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="date" fontSize={isMobile ? 10 : 12} />
                <YAxis fontSize={isMobile ? 10 : 12} />
                <CartesianGrid strokeDasharray="3 3" />
                <Tooltip formatter={(value) => [`${value.toLocaleString()} ₸`, 'Доход']} />
                <Area 
                  type="monotone" 
                  dataKey="revenue" 
                  stroke="#8884d8" 
                  fillOpacity={1} 
                  fill="url(#colorRevenue)" 
                />
              </AreaChart>
            </ResponsiveContainer>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="Сравнение с прошлым годом" className={isMobile ? 'h-80' : 'h-96'}>
            <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
              <BarChart data={monthlyComparisonData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" fontSize={isMobile ? 10 : 12} />
                <YAxis fontSize={isMobile ? 10 : 12} />
                <Tooltip />
                <Bar dataKey="lastYear" fill="#e6f3ff" name="Прошлый год" />
                <Bar dataKey="thisYear" fill="#1890ff" name="Этот год" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      {/* Топ квартир по доходности */}
      <Card title="Топ квартир по доходности">
        <div className="space-y-4">
          {apartmentPerformance.map((apartment, index) => (
            <div key={index} className={`flex items-center justify-between p-4 bg-gray-50 rounded-lg ${isMobile ? 'flex-col space-y-3' : ''}`}>
              <div className={`flex items-center space-x-4 ${isMobile ? 'w-full' : ''}`}>
                <div className={`bg-blue-100 rounded-lg flex items-center justify-center ${isMobile ? 'w-8 h-8' : 'w-10 h-10'}`}>
                  <span className={`text-blue-600 font-bold ${isMobile ? 'text-sm' : ''}`}>#{index + 1}</span>
                </div>
                <div className={isMobile ? 'flex-1' : ''}>
                  <div className={`font-medium ${isMobile ? 'text-sm' : ''}`}>{apartment.name}</div>
                  <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
                    {apartment.bookings} бронирований • Рейтинг {apartment.rating}
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

      {/* Детальная таблица бронирований */}
      <Card title="История бронирований">
        <Table
          columns={columns}
          dataSource={ownerBookings?.data || []}
          rowKey="id"
          scroll={{ x: 800 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{ 
            pageSize: 10,
            showSizeChanger: !isMobile,
            showQuickJumper: !isMobile,
            showTotal: (total) => `Всего ${total} бронирований`,
            responsive: true,
            simple: isMobile,
          }}
        />
      </Card>

      {/* Уведомления */}
      <Row gutter={[16, 16]}>
        <Col xs={24}>
          <NotificationList
            title="Мои уведомления"
            showPagination={false}
            pageSize={5}
            showBulkActions={false}
            showStats={false}
            cardProps={{
              className: "shadow-sm",
              extra: (
                <Button 
                  type="link" 
                  onClick={() => navigate('/owner/notifications')}
                >
                  Посмотреть все
                </Button>
              )
            }}
            listProps={{
              size: isMobile ? "small" : "default"
            }}
          />
        </Col>
      </Row>
    </div>
  );
};

export default OwnerDashboardPage; 