import React, { useMemo, useCallback } from 'react';
import { Row, Col, Select, DatePicker, Button, Space, Typography, Spin } from 'antd';
import { 
  UserOutlined,
  HomeOutlined, 
  CalendarOutlined,
  DollarOutlined,
  LockOutlined,
  ReloadOutlined,
  DownloadOutlined
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { apartmentsAPI } from '../../lib/api.js';
import useDashboardStore from '../../store/useDashboardStore.js';
import {
  KPICard,
  DashboardLineChart,
  DashboardBarChart,
  DashboardPieChart,
  DashboardAreaChart,
  DashboardProgressBar
} from '../../components/dashboard/index.js';
import NotificationList from '../../components/NotificationList.jsx';

const { Title } = Typography;
const { Option } = Select;
const { RangePicker } = DatePicker;

const DashboardPage = () => {
  const navigate = useNavigate();
  
  // Состояние из store
  const {
    period,
    dateRange,
    setPeriod,
    setDateRange,
    getApiParams,
    autoUpdateDates,
    setAutoUpdateDates
  } = useDashboardStore();

  // Получение данных из API
  const { 
    data: dashboardData, 
    isLoading, 
    refetch,
    isFetching 
  } = useQuery({
    queryKey: ['admin-dashboard-statistics', period, dateRange],
    queryFn: () => apartmentsAPI.adminGetDashboardStatistics(getApiParams()),
    staleTime: 5 * 60 * 1000, // 5 минут
    refetchOnWindowFocus: false,
  });

  const stats = dashboardData?.data;

  // Форматеры для графиков
  const formatCurrency = useCallback((value) => `${value.toLocaleString()} ₸`, []);
  const formatNumber = useCallback((value) => value.toLocaleString(), []);
  const formatDate = useCallback((date) => {
    if (period === 'day') return dayjs(date).format('DD.MM');
    if (period === 'week') {
      // Обрабатываем формат "2025-W22"
      if (date.includes('W')) {
        const [year, week] = date.split('-W');
        return `Н${week}/${year.slice(-2)}`;
      }
      return dayjs(date).format('DD.MM');
    }
    return dayjs(date).format('MM.YYYY');
  }, [period]);

  // Обработчики событий
  const handlePeriodChange = useCallback((newPeriod) => {
    setPeriod(newPeriod);
  }, [setPeriod]);

  const handleDateRangeChange = useCallback((dates) => {
    if (dates && dates.length === 2) {
      setDateRange(
        dates[0].format('YYYY-MM-DD'),
        dates[1].format('YYYY-MM-DD')
      );
    }
  }, [setDateRange]);

  const handleRefresh = useCallback(() => {
    refetch();
  }, [refetch]);

  // Обработчики навигации для кликабельных плашек
  const handleUsersClick = useCallback(() => {
    navigate('/admin/users');
  }, [navigate]);

  const handleApartmentsClick = useCallback(() => {
    navigate('/admin/apartments');
  }, [navigate]);

  const handleBookingsClick = useCallback(() => {
    navigate('/admin/bookings');
  }, [navigate]);

  // Мемоизированные данные для графиков
  const timeSeriesData = useMemo(() => {
    if (!stats?.time_series) return [];
    
    const dates = stats.time_series.apartments?.map(item => item.date) || [];
    
    return dates.map(date => ({
      date,
      apartments: stats.time_series.apartments?.find(item => item.date === date)?.value || 0,
      bookings: stats.time_series.bookings?.find(item => item.date === date)?.value || 0,
      revenue: stats.time_series.revenue?.find(item => item.date === date)?.value || 0,
      users: stats.time_series.users?.find(item => item.date === date)?.value || 0,
    }));
  }, [stats?.time_series]);

  const userRoleData = useMemo(() => {
    return stats?.users?.by_role || {};
  }, [stats?.users?.by_role]);

  const bookingStatusData = useMemo(() => {
    return stats?.bookings?.by_status || {};
  }, [stats?.bookings?.by_status]);

  const bookingDurationData = useMemo(() => {
    return stats?.bookings?.by_duration || {};
  }, [stats?.bookings?.by_duration]);

  const apartmentStatusData = useMemo(() => {
    return stats?.apartments?.by_status || {};
  }, [stats?.apartments?.by_status]);

  const apartmentListingTypeData = useMemo(() => {
    return stats?.apartments?.by_listing_type || {};
  }, [stats?.apartments?.by_listing_type]);

  const lockStatusData = useMemo(() => {
    return stats?.locks?.by_status || {};
  }, [stats?.locks?.by_status]);

  const lockBatteryData = useMemo(() => {
    return stats?.locks?.by_battery || {};
  }, [stats?.locks?.by_battery]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Заголовок и управление */}
      <div className="flex flex-col lg:flex-row lg:justify-between lg:items-center space-y-4 lg:space-y-0">
        <div>
          <Title level={2} className="!mb-2">Дашборд администратора</Title>
          <p className="text-gray-600">
            Аналитика с {dayjs(dateRange.date_from).format('DD.MM.YYYY')} по {dayjs(dateRange.date_to).format('DD.MM.YYYY')}
          </p>
        </div>
        
        <Space wrap>
          <Select
            value={period}
            onChange={handlePeriodChange}
            style={{ width: 120 }}
          >
            <Option value="day">День</Option>
            <Option value="week">Неделя</Option>
            <Option value="month">Месяц</Option>
          </Select>
          
          <RangePicker
            value={[dayjs(dateRange.date_from), dayjs(dateRange.date_to)]}
            onChange={handleDateRangeChange}
            format="DD.MM.YYYY"
            allowClear={false}
          />
          
          <Button 
            icon={<ReloadOutlined />} 
            onClick={handleRefresh}
            loading={isFetching}
          >
            Обновить
          </Button>
          
          <Button 
            type="primary" 
            icon={<DownloadOutlined />}
          >
            Экспорт
          </Button>
        </Space>
      </div>

      {/* KPI Карточки */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Активные пользователи"
            value={stats?.totals?.active_users || 0}
            prefix={<UserOutlined />}
            growthRate={stats?.users?.growth_rate}
            loading={isLoading}
            tooltipTitle="Количество активных пользователей в системе. Нажмите для перехода к управлению пользователями"
            clickable={true}
            onClick={handleUsersClick}
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Всего квартир"
            value={stats?.totals?.total_apartments || 0}
            prefix={<HomeOutlined />}
            growthRate={stats?.apartments?.growth_rate}
            loading={isLoading}
            tooltipTitle="Общее количество квартир в системе. Нажмите для перехода к управлению квартирами"
            clickable={true}
            onClick={handleApartmentsClick}
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Всего бронирований"
            value={stats?.totals?.total_bookings || 0}
            prefix={<CalendarOutlined />}
            growthRate={stats?.bookings?.growth_rate}
            loading={isLoading}
            tooltipTitle="Общее количество бронирований. Нажмите для перехода к управлению бронированиями"
            clickable={true}
            onClick={handleBookingsClick}
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Общий доход"
            value={stats?.financial?.total_revenue || 0}
            suffix="₸"
            prefix={<DollarOutlined />}
            growthRate={stats?.bookings?.revenue_growth_rate}
            formatter={formatNumber}
            loading={isLoading}
            tooltipTitle="Общий доход от всех бронирований"
          />
        </Col>
      </Row>

      {/* Дополнительные KPI */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Онлайн замки"
            value={stats?.totals?.online_locks || 0}
            prefix={<LockOutlined />}
            loading={isLoading}
            size="small"
            tooltipTitle="Количество активных замков"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Средний чек"
            value={stats?.bookings?.avg_booking_value || 0}
            suffix="₸"
            formatter={formatNumber}
            loading={isLoading}
            size="small"
            tooltipTitle="Средняя стоимость бронирования"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Комиссия"
            value={stats?.financial?.commission || 0}
            suffix="₸"
            formatter={formatNumber}
            loading={isLoading}
            size="small"
            tooltipTitle="Общая сумма комиссии"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <KPICard
            title="Доходы за период"
            value={stats?.bookings?.period_revenue || 0}
            suffix="₸"
            formatter={formatNumber}
            loading={isLoading}
            size="small"
            tooltipTitle="Доходы за выбранный период"
          />
        </Col>
      </Row>

      {/* Временные ряды - Line Charts */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <DashboardLineChart
            title="Динамика бронирований и пользователей"
            data={timeSeriesData}
            dataKeys={[
              { dataKey: 'bookings', name: 'Бронирования', color: '#1890ff' },
              { dataKey: 'users', name: 'Пользователи', color: '#52c41a' }
            ]}
            height={300}
            loading={isLoading}
            formatXAxisLabel={formatDate}
            formatTooltip={(value, name) => [formatNumber(value), name]}
          />
        </Col>
        <Col xs={24} lg={12}>
          <DashboardAreaChart
            title="Накопительный доход"
            data={timeSeriesData}
            dataKeys={[
              { dataKey: 'revenue', name: 'Доход', color: '#faad14' }
            ]}
            height={300}
            loading={isLoading}
            formatXAxisLabel={formatDate}
            formatYAxisLabel={formatCurrency}
            formatTooltip={(value, name) => [formatCurrency(value), name]}
          />
        </Col>
      </Row>

      {/* Статистика по категориям - Bar Charts */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <DashboardBarChart
            title="Бронирования по статусам"
            data={bookingStatusData}
            height={300}
            loading={isLoading}
            colors={['#52c41a', '#1890ff', '#faad14', '#f5222d']}
            formatValue={formatNumber}
          />
        </Col>
        <Col xs={24} lg={8}>
          <DashboardBarChart
            title="Квартиры по статусам"
            data={apartmentStatusData}
            height={300}
            loading={isLoading}
            colors={['#52c41a', '#faad14', '#f5222d']}
            formatValue={formatNumber}
          />
        </Col>
        <Col xs={24} lg={8}>
          <DashboardBarChart
            title="Квартиры по типу листинга"
            data={apartmentListingTypeData}
            height={300}
            loading={isLoading}
            colors={['#1890ff', '#722ed1']}
            formatValue={formatNumber}
          />
        </Col>
      </Row>

      {/* Pie Charts - распределения */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <DashboardPieChart
            title="Пользователи по ролям"
            data={userRoleData}
            height={300}
            loading={isLoading}
            colors={['#f5222d', '#faad14', '#1890ff', '#52c41a']}
            formatValue={formatNumber}
          />
        </Col>
        <Col xs={24} lg={8}>
          <DashboardPieChart
            title="Бронирования по длительности"
            data={bookingDurationData}
            height={300}
            loading={isLoading}
            colors={['#52c41a', '#1890ff', '#faad14']}
            formatValue={formatNumber}
          />
        </Col>
        <Col xs={24} lg={8}>
          <DashboardPieChart
            title="Статус замков"
            data={lockStatusData}
            height={300}
            loading={isLoading}
            colors={['#52c41a', '#f5222d', '#8c8c8c']}
            formatValue={formatNumber}
            innerRadius={40}
          />
        </Col>
      </Row>

      {/* Progress Bars - уровни батареи и состояние */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <DashboardProgressBar
            title="Уровень батареи замков"
            data={lockBatteryData}
            type="battery"
            layout="horizontal"
            loading={isLoading}
            size="default"
          />
        </Col>
        <Col xs={24} lg={12}>
          <DashboardProgressBar
            title="Квартиры по городам"
            data={stats?.users?.by_city || {}}
            type="occupancy"
            layout="horizontal"
            loading={isLoading}
            size="default"
          />
        </Col>
      </Row>

      {/* Последние уведомления */}
      <Row gutter={[16, 16]}>
        <Col xs={24}>
          <NotificationList
            title="Последние уведомления"
            showPagination={false}
            pageSize={5}
            showBulkActions={false}
            showStats={false}
            cardProps={{
              className: "shadow-sm",
              extra: (
                <Button 
                  type="link" 
                  onClick={() => navigate('/admin/notifications')}
                >
                  Посмотреть все
                </Button>
              )
            }}
            listProps={{
              size: "small"
            }}
          />
        </Col>
      </Row>
    </div>
  );
};

export default DashboardPage; 