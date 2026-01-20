import React, { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Card, Row, Col, Statistic, Select, DatePicker, Typography, Table,
  Space, Progress, Tag, Divider, Alert, Spin, Empty
} from 'antd';
import {
  DollarOutlined, CalendarOutlined, HomeOutlined, RiseOutlined,
  PercentageOutlined, ClockCircleOutlined, StarOutlined, UserOutlined,
  CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined
} from '@ant-design/icons';
import {
  LineChart, Line, AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer
} from 'recharts';
import { apartmentsAPI } from '../../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { RangePicker } = DatePicker;

const OwnerStatisticsPage = () => {
  const [dateRange, setDateRange] = useState([
    dayjs().subtract(1, 'month'),
    dayjs()
  ]);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const { data: statsResponse, isLoading, isError, error } = useQuery({
    queryKey: ['owner-statistics', dateRange],
    queryFn: () => apartmentsAPI.getOwnerStatistics({
      date_from: dateRange?.[0]?.format('YYYY-MM-DD'),
      date_to: dateRange?.[1]?.format('YYYY-MM-DD')
    }),
    enabled: !!dateRange && dateRange.length === 2,
  });

  const stats = statsResponse?.data;
  const apartmentStats = stats?.apartments || {};
  const bookingStats = stats?.bookings || {};
  const financialStats = stats?.financial || {};
  const efficiencyStats = stats?.efficiency || {};

  // Преобразование данных для графиков
  const revenueByMonthData = React.useMemo(() => {
    const monthlyRevenue = financialStats?.revenue_by_month || {};
    return Object.entries(monthlyRevenue)
      .map(([month, revenue]) => ({
        month: dayjs(month).format('MMM YYYY'),
        revenue,
        sortKey: month
      }))
      .sort((a, b) => a.sortKey.localeCompare(b.sortKey))
      .map(({ month, revenue }) => ({ month, revenue }));
  }, [financialStats?.revenue_by_month]);

  // Данные для круговой диаграммы статусов бронирований
  const bookingStatusData = React.useMemo(() => {
    const byStatus = bookingStats?.by_status || {};
    const statusConfig = {
      created: { name: 'Создано', color: '#1890ff' },
      pending: { name: 'Ожидает', color: '#faad14' },
      approved: { name: 'Одобрено', color: '#52c41a' },
      rejected: { name: 'Отклонено', color: '#ff4d4f' },
      active: { name: 'Активно', color: '#13c2c2' },
      completed: { name: 'Завершено', color: '#722ed1' },
      canceled: { name: 'Отменено', color: '#f5222d' },
    };

    return Object.entries(byStatus)
      .filter(([_, value]) => value > 0)
      .map(([status, value]) => ({
        name: statusConfig[status]?.name || status,
        value,
        color: statusConfig[status]?.color || '#8884d8'
      }));
  }, [bookingStats?.by_status]);

  // Данные для диаграммы статусов квартир
  const apartmentStatusData = React.useMemo(() => {
    const byStatus = apartmentStats?.by_status || {};
    const statusConfig = {
      approved: { name: 'Одобрено', color: '#52c41a' },
      pending: { name: 'На модерации', color: '#faad14' },
      rejected: { name: 'Отклонено', color: '#ff4d4f' },
      blocked: { name: 'Заблокировано', color: '#f5222d' },
    };

    return Object.entries(byStatus)
      .filter(([_, value]) => value > 0)
      .map(([status, value]) => ({
        name: statusConfig[status]?.name || status,
        value,
        color: statusConfig[status]?.color || '#8884d8'
      }));
  }, [apartmentStats?.by_status]);

  // Данные для диаграммы дохода по длительности
  const revenueByDurationData = React.useMemo(() => {
    const byDuration = financialStats?.revenue_by_duration || {};
    const durationConfig = {
      short: { name: 'Краткосрочные (до 3ч)', color: '#1890ff' },
      medium: { name: 'Средние (3-12ч)', color: '#52c41a' },
      long: { name: 'Долгие (12-24ч)', color: '#722ed1' },
      daily: { name: 'Посуточные', color: '#fa8c16' },
    };

    return Object.entries(byDuration)
      .filter(([_, value]) => value > 0)
      .map(([duration, value]) => ({
        name: durationConfig[duration]?.name || duration,
        value,
        color: durationConfig[duration]?.color || '#8884d8'
      }));
  }, [financialStats?.revenue_by_duration]);

  // Колонки для таблицы эффективности квартир
  const apartmentColumns = [
    {
      title: 'Квартира',
      key: 'name',
      render: (record) => (
        <Text strong>
          {record.street}, {record.building}, кв. {record.apartment_number}
        </Text>
      )
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const config = {
          approved: { color: 'green', text: 'Активна' },
          pending: { color: 'orange', text: 'На модерации' },
          rejected: { color: 'red', text: 'Отклонена' },
          blocked: { color: 'red', text: 'Заблокирована' },
        };
        const cfg = config[status] || { color: 'default', text: status };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      }
    },
    {
      title: 'Доход',
      dataIndex: 'total_revenue',
      key: 'total_revenue',
      render: (revenue) => (
        <Text strong style={{ color: '#52c41a' }}>
          {revenue?.toLocaleString()} ₸
        </Text>
      ),
      sorter: (a, b) => a.total_revenue - b.total_revenue
    },
    {
      title: 'Бронирований',
      dataIndex: 'total_bookings',
      key: 'total_bookings',
      render: (bookings) => <Text>{bookings}</Text>,
      sorter: (a, b) => a.total_bookings - b.total_bookings
    },
    {
      title: 'Завершено',
      dataIndex: 'completed_bookings',
      key: 'completed_bookings',
      render: (completed, record) => (
        <div className="flex items-center space-x-2">
          <Text>{completed}</Text>
          {record.total_bookings > 0 && (
            <Text type="secondary" className="text-xs">
              ({Math.round((completed / record.total_bookings) * 100)}%)
            </Text>
          )}
        </div>
      )
    },
    {
      title: 'Ср. доход',
      dataIndex: 'avg_revenue',
      key: 'avg_revenue',
      render: (avg) => (
        <Text>{Math.round(avg || 0).toLocaleString()} ₸</Text>
      ),
      sorter: (a, b) => (a.avg_revenue || 0) - (b.avg_revenue || 0)
    }
  ];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Spin indicator={<LoadingOutlined style={{ fontSize: 48 }} spin />} />
      </div>
    );
  }

  if (isError) {
    return (
      <div className={`space-y-6 ${isMobile ? 'p-4' : 'p-6'}`}>
        <Alert
          message="Ошибка загрузки статистики"
          description={error?.message || 'Не удалось загрузить данные. Попробуйте позже.'}
          type="error"
          showIcon
        />
      </div>
    );
  }

  const totalBookings = bookingStats?.total_bookings || 0;
  const revenueBookings = financialStats?.revenue_bookings || 0;
  const completedBookings = bookingStats?.by_status?.completed || 0;
  const canceledBookings = bookingStats?.by_status?.canceled || 0;

  return (
    <div className={`space-y-6 ${isMobile ? 'p-4' : 'p-6'}`}>
      {/* Заголовок и фильтры */}
      <div className={`${isMobile ? 'space-y-4' : 'flex flex-col lg:flex-row lg:justify-between lg:items-center space-y-4 lg:space-y-0'}`}>
        <div>
          <Title level={2} className={isMobile ? 'text-xl' : ''}>Статистика и аналитика</Title>
          <Text type="secondary" className={isMobile ? 'text-sm' : ''}>
            Подробная аналитика доходов и эффективности за период
          </Text>
        </div>
        
        <Space wrap direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
          <RangePicker
            value={dateRange}
            onChange={setDateRange}
            format="DD.MM.YYYY"
            style={{ width: isMobile ? '100%' : 'auto' }}
            presets={[
              { label: 'Неделя', value: [dayjs().subtract(7, 'day'), dayjs()] },
              { label: 'Месяц', value: [dayjs().subtract(1, 'month'), dayjs()] },
              { label: '3 месяца', value: [dayjs().subtract(3, 'month'), dayjs()] },
              { label: 'Год', value: [dayjs().subtract(1, 'year'), dayjs()] },
            ]}
          />
        </Space>
      </div>

      {/* Основная статистика */}
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Общий доход"
              value={financialStats?.total_revenue || 0}
              prefix={<DollarOutlined className="text-green-500" />}
              suffix="₸"
              precision={0}
              valueStyle={{ 
                color: '#52c41a',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Text className={`text-green-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              За выбранный период
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Средний чек"
              value={Math.round(financialStats?.avg_booking_value || 0)}
              prefix={<PercentageOutlined className="text-blue-500" />}
              suffix="₸"
              precision={0}
              valueStyle={{ 
                color: '#1890ff',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Text className={`text-blue-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              Средняя стоимость бронирования
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Всего бронирований"
              value={totalBookings}
              prefix={<CalendarOutlined className="text-purple-500" />}
              valueStyle={{ 
                color: '#722ed1',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Text className={`text-purple-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              За выбранный период
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Оплаченные"
              value={revenueBookings}
              prefix={<CheckCircleOutlined className="text-cyan-500" />}
              valueStyle={{ 
                color: '#13c2c2',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Text className={`text-cyan-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              Одобрено, активных и завершенных
            </Text>
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Завершено"
              value={completedBookings}
              prefix={<CheckCircleOutlined className="text-green-500" />}
              valueStyle={{ 
                color: '#52c41a',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            {totalBookings > 0 && (
              <Progress 
                percent={Math.round((completedBookings / totalBookings) * 100)} 
                showInfo={false} 
                strokeColor="#52c41a"
                size={isMobile ? 'small' : 'default'}
              />
            )}
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Отменено"
              value={canceledBookings}
              prefix={<CloseCircleOutlined className="text-red-500" />}
              valueStyle={{ 
                color: '#ff4d4f',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            {totalBookings > 0 && (
              <Progress 
                percent={Math.round((canceledBookings / totalBookings) * 100)} 
                showInfo={false} 
                strokeColor="#ff4d4f"
                size={isMobile ? 'small' : 'default'}
              />
            )}
          </Card>
        </Col>

        <Col xs={12} sm={8} md={6} lg={4}>
          <Card className="text-center">
            <Statistic
              title="Квартир"
              value={apartmentStats?.total_apartments || 0}
              prefix={<HomeOutlined className="text-cyan-500" />}
              valueStyle={{ 
                color: '#13c2c2',
                fontSize: isMobile ? '18px' : '24px'
              }}
            />
            <Text className={`text-cyan-600 ${isMobile ? 'text-xs' : 'text-sm'}`}>
              Всего в портфеле
            </Text>
          </Card>
        </Col>
      </Row>

      {/* Графики */}
      <Row gutter={[16, 16]}>
        {/* График доходов по месяцам */}
        <Col xs={24} lg={16}>
          <Card title="Динамика доходов по месяцам" className={isMobile ? 'h-80' : 'h-96'}>
            {revenueByMonthData.length > 0 ? (
              <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
                <AreaChart data={revenueByMonthData}>
                  <defs>
                    <linearGradient id="colorRevenue" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#52c41a" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="#52c41a" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <XAxis 
                    dataKey="month" 
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
            ) : (
              <Empty description="Нет данных за выбранный период" />
            )}
          </Card>
        </Col>

        {/* Распределение по статусам бронирований */}
        <Col xs={24} lg={8}>
          <Card title="Статусы бронирований" className={isMobile ? 'h-80' : 'h-96'}>
            {bookingStatusData.length > 0 ? (
              <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
                <PieChart>
                  <Pie
                    data={bookingStatusData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }) => isMobile ? `${(percent * 100).toFixed(0)}%` : `${name} ${(percent * 100).toFixed(0)}%`}
                    outerRadius={isMobile ? 70 : 80}
                    fill="#8884d8"
                    dataKey="value"
                    fontSize={isMobile ? 10 : 12}
                  >
                    {bookingStatusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                  {!isMobile && <Legend />}
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <Empty description="Нет бронирований" />
            )}
          </Card>
        </Col>
      </Row>

      {/* Доход по типу аренды */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="Доход по типу аренды" className={isMobile ? 'h-80' : 'h-96'}>
            {revenueByDurationData.length > 0 ? (
              <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
                <BarChart data={revenueByDurationData} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis type="number" fontSize={isMobile ? 10 : 12} />
                  <YAxis 
                    dataKey="name" 
                    type="category" 
                    fontSize={isMobile ? 10 : 12}
                    width={isMobile ? 100 : 150}
                  />
                  <Tooltip 
                    formatter={(value) => [`${value.toLocaleString()} ₸`, 'Доход']}
                  />
                  <Bar dataKey="value" name="Доход">
                    {revenueByDurationData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <Empty description="Нет данных о доходе" />
            )}
          </Card>
        </Col>

        {/* Статусы квартир */}
        <Col xs={24} lg={12}>
          <Card title="Статусы квартир" className={isMobile ? 'h-80' : 'h-96'}>
            {apartmentStatusData.length > 0 ? (
              <ResponsiveContainer width="100%" height={isMobile ? 250 : 300}>
                <PieChart>
                  <Pie
                    data={apartmentStatusData}
                    cx="50%"
                    cy="50%"
                    labelLine={true}
                    label={({ name, value }) => `${name}: ${value}`}
                    outerRadius={isMobile ? 70 : 100}
                    fill="#8884d8"
                    dataKey="value"
                    fontSize={isMobile ? 10 : 12}
                  >
                    {apartmentStatusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <Empty description="Нет данных о квартирах" />
            )}
          </Card>
        </Col>
      </Row>

      {/* Топ квартир по доходности */}
      {efficiencyStats?.top_performers?.length > 0 && (
        <Card title="Топ квартир по доходности">
          <div className="space-y-3">
            {efficiencyStats.top_performers.map((apartment, index) => (
              <div key={apartment.apartment_id} className={`flex items-center justify-between p-3 bg-gray-50 rounded-lg ${isMobile ? 'flex-col space-y-2' : ''}`}>
                <div className={`flex items-center space-x-3 ${isMobile ? 'w-full' : ''}`}>
                  <div className={`bg-blue-100 rounded-lg flex items-center justify-center ${isMobile ? 'w-8 h-8' : 'w-10 h-10'}`}>
                    <span className={`text-blue-600 font-bold ${isMobile ? 'text-sm' : ''}`}>#{index + 1}</span>
                  </div>
                  <div className={isMobile ? 'flex-1' : ''}>
                    <div className={`font-medium ${isMobile ? 'text-sm' : ''}`}>
                      {apartment.street}, {apartment.building}, кв. {apartment.apartment_number}
                    </div>
                    <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
                      {apartment.completed_bookings} завершено из {apartment.total_bookings} бронирований
                    </div>
                  </div>
                </div>
                <div className={`text-right ${isMobile ? 'w-full text-center' : ''}`}>
                  <div className={`font-bold text-green-600 ${isMobile ? 'text-base' : 'text-lg'}`}>
                    {apartment.total_revenue.toLocaleString()} ₸
                  </div>
                  <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
                    Ср. чек: {Math.round(apartment.avg_revenue || 0).toLocaleString()} ₸
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Детальная таблица эффективности квартир */}
      <Card title="Эффективность квартир">
        <Table
          columns={apartmentColumns}
          dataSource={efficiencyStats?.apartment_performance || []}
          rowKey="apartment_id"
          scroll={{ x: 800 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{ 
            pageSize: 10,
            showSizeChanger: !isMobile,
            showTotal: (total) => `Всего ${total} квартир`,
            responsive: true,
            simple: isMobile,
          }}
          locale={{
            emptyText: <Empty description="Нет данных о квартирах" />
          }}
        />
      </Card>

      {/* Популярные квартиры */}
      {bookingStats?.popular_apartments?.length > 0 && (
        <Card title="Самые популярные квартиры (по количеству бронирований)">
          <Row gutter={[16, 16]}>
            {bookingStats.popular_apartments.map((apt, index) => (
              <Col xs={24} sm={12} md={8} lg={6} key={apt.apartment_id}>
                <Card size="small" className="text-center">
                  <div className="text-2xl font-bold text-blue-600 mb-2">#{index + 1}</div>
                  <div className="font-medium text-sm mb-1">
                    {apt.street}, {apt.building}
                  </div>
                  <div className="text-gray-500 text-xs mb-2">кв. {apt.apartment_number}</div>
                  <Tag color="blue">{apt.booking_count} бронирований</Tag>
                </Card>
              </Col>
            ))}
          </Row>
        </Card>
      )}

      {/* Средние показатели квартир */}
      <Card title="Средние показатели портфеля">
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8}>
            <Statistic
              title="Средняя площадь"
              value={Math.round(apartmentStats?.avg_area || 0)}
              suffix="м²"
              prefix={<HomeOutlined />}
            />
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Statistic
              title="Средняя цена (месяц)"
              value={Math.round(apartmentStats?.avg_price || 0)}
              suffix="₸"
              prefix={<DollarOutlined />}
            />
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Statistic
              title="Средняя цена (сутки)"
              value={Math.round(apartmentStats?.avg_daily_price || 0)}
              suffix="₸"
              prefix={<DollarOutlined />}
            />
          </Col>
        </Row>

        {Object.keys(apartmentStats?.by_room_count || {}).length > 0 && (
          <>
            <Divider />
            <Title level={5}>Распределение по комнатности</Title>
            <Space wrap>
              {Object.entries(apartmentStats.by_room_count)
                .sort(([a], [b]) => Number(a) - Number(b))
                .map(([rooms, count]) => (
                  <Tag key={rooms} color="blue">
                    {rooms}-комн: {count} шт.
                  </Tag>
                ))}
            </Space>
          </>
        )}
      </Card>
    </div>
  );
};

export default OwnerStatisticsPage;
