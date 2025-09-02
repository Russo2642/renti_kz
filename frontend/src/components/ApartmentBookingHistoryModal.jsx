import { CalendarOutlined, DollarOutlined, HistoryOutlined, UserOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { 
  Button, Card, Col, DatePicker, Descriptions, Modal, Row, 
  Select, Space, Statistic, Table, Tag, Typography, Pagination, Progress 
} from 'antd';
import dayjs from 'dayjs';
import React, { useState, useEffect } from 'react';
import { apartmentsAPI } from '../lib/api.js';

const { RangePicker } = DatePicker;
const { Option } = Select;
const { Title, Text } = Typography;

const ApartmentBookingHistoryModal = ({ visible, onClose, apartment }) => {
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [filters, setFilters] = useState({
    status: [],
    date_from: null,
    date_to: null,
    page: 1,
    page_size: 10
  });

  // Отслеживание размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Загрузка истории бронирований квартиры
  const { data: historyData, isLoading } = useQuery({
    queryKey: ['apartmentBookingHistory', apartment?.id, filters],
    queryFn: () => apartmentsAPI.adminGetApartmentBookingsHistory(apartment?.id, filters),
    enabled: visible && !!apartment?.id,
  });

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      page: 1 // Сбрасываем на первую страницу при изменении фильтров
    }));
  };

  const handleDateRangeChange = (dates) => {
    if (dates && dates.length === 2) {
      setFilters(prev => ({
        ...prev,
        date_from: dates[0].format('YYYY-MM-DD'),
        date_to: dates[1].format('YYYY-MM-DD'),
        page: 1
      }));
    } else {
      setFilters(prev => ({
        ...prev,
        date_from: null,
        date_to: null,
        page: 1
      }));
    }
  };

  const getStatusColor = (status) => {
    const colors = {
      created: 'blue',
      awaiting_payment: 'orange',
      pending: 'orange',
      approved: 'green',
      rejected: 'red',
      active: 'cyan',
      completed: 'success',
      canceled: 'error'
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status) => {
    const texts = {
      created: 'Создано',
      awaiting_payment: 'Ожидает оплаты',
      pending: 'На рассмотрении',
      approved: 'Одобрено',
      rejected: 'Отклонено',
      active: 'Активно',
      completed: 'Завершено',
      canceled: 'Отменено'
    };
    return texts[status] || status;
  };

  const getStatusTextShort = (status) => {
    const shortTexts = {
      created: 'Созд.',
      awaiting_payment: 'Ожид.',
      pending: 'На расс.',
      approved: 'Одоб.',
      rejected: 'Откл.',
      active: 'Актив.',
      completed: 'Заверш.',
      canceled: 'Отмен.'
    };
    return shortTexts[status] || status;
  };

  const columns = [
    {
      title: 'Номер',
      dataIndex: 'booking_number',
      key: 'booking_number',
      width: isMobile ? 80 : 120,
      render: (number) => <Text strong className={isMobile ? 'text-xs' : ''}>{number}</Text>
    },
    {
      title: 'Арендатор',
      key: 'renter',
      width: isMobile ? 120 : 180,
      render: (record) => (
        <div className={isMobile ? 'text-xs' : ''}>
          <Text strong>{record.renter?.user?.first_name} {record.renter?.user?.last_name}</Text>
          <br />
          <Text type="secondary">{record.renter?.user?.phone}</Text>
        </div>
      )
    },
    ...(isMobile ? [] : [{
      title: 'Период',
      key: 'period',
      width: 140,
      render: (record) => (
        <div className="text-xs">
          <div>{dayjs(record.start_date).format('DD.MM.YY HH:mm')}</div>
          <div>{dayjs(record.end_date).format('DD.MM.YY HH:mm')}</div>
          <Text type="secondary">{record.duration} ч.</Text>
        </div>
      )
    }]),
    {
      title: 'Сумма',
      dataIndex: 'final_price',
      key: 'final_price',
      width: isMobile ? 80 : 100,
      render: (price) => (
        <Text strong className={isMobile ? 'text-xs' : ''}>
          {isMobile ? `${Math.round(price / 1000)}к` : `${price?.toLocaleString()} ₸`}
        </Text>
      )
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: isMobile ? 80 : 120,
      render: (status) => (
        <Tag color={getStatusColor(status)} className={isMobile ? 'text-xs px-1' : ''}>
          {isMobile ? getStatusTextShort(status) : getStatusText(status)}
        </Tag>
      )
    },
    ...(isMobile ? [] : [{
      title: 'Статус замка',
      dataIndex: 'door_status',
      key: 'door_status',
      width: 100,
      render: (status) => (
        <Tag color={status === 'open' ? 'red' : 'green'} className="text-xs">
          {status === 'open' ? 'Открыт' : 'Закрыт'}
        </Tag>
      )
    }]),
    ...(isMobile ? [{
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 70,
      render: (date) => (
        <Text className="text-xs">
          {dayjs(date).format('DD.MM')}
        </Text>
      )
    }] : [{
      title: 'Создано',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 130,
      render: (date) => dayjs(date).format('DD.MM.YYYY HH:mm')
    }])
  ];

  const analytics = historyData?.data?.analytics || {};
  const bookings = historyData?.data?.bookings || [];
  const pagination = historyData?.data?.pagination || {};
  const apartmentInfo = historyData?.data?.apartment || apartment;

  return (
    <Modal
      title={
        <Space>
          <HistoryOutlined />
          <span className={isMobile ? 'text-sm' : ''}>
            История бронирований квартиры
          </span>
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={isMobile ? '100%' : 1200}
      footer={null}
      styles={{ 
        body: { 
          maxHeight: isMobile ? '80vh' : '70vh', 
          overflowY: 'auto',
          padding: isMobile ? '12px' : '24px'
        } 
      }}
    >
      {apartmentInfo && (
        <div className="space-y-4">
          {/* Информация о квартире */}
          <Card size="small">
            <Descriptions column={isMobile ? 1 : 3} size="small">
              <Descriptions.Item label="Адрес">
                {apartmentInfo.street}, {apartmentInfo.building}-{apartmentInfo.apartment_number}
              </Descriptions.Item>
              <Descriptions.Item label="Описание">
                {apartmentInfo.description ? 
                  apartmentInfo.description.substring(0, isMobile ? 40 : 80) + '...' : 
                  '—'
                }
              </Descriptions.Item>
              <Descriptions.Item label="Цена">
                {isMobile ? 
                  `${Math.round(apartmentInfo.price / 1000)}к/ч` :
                  `${apartmentInfo.price?.toLocaleString()} ₸/час`
                }
                {apartmentInfo.daily_price && (
                  <>, {isMobile ? 
                    `${Math.round(apartmentInfo.daily_price / 1000)}к/сут` :
                    `${apartmentInfo.daily_price?.toLocaleString()} ₸/сут`
                  }</>
                )}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          {/* Фильтры */}
          <Card size="small" title={<span className={isMobile ? 'text-sm' : ''}>Фильтры</span>}>
            <Row gutter={[8, 8]}>
              <Col xs={24} sm={12} md={8}>
                <Text className={isMobile ? 'text-xs' : ''}>Статус:</Text>
                <Select
                  mode="multiple"
                  placeholder="Все статусы"
                  style={{ width: '100%' }}
                  size={isMobile ? 'small' : 'middle'}
                  value={filters.status}
                  onChange={(value) => handleFilterChange('status', value)}
                >
                  <Option value="created">Создано</Option>
                  <Option value="pending">На рассмотрении</Option>
                  <Option value="approved">Одобрено</Option>
                  <Option value="active">Активно</Option>
                  <Option value="completed">Завершено</Option>
                  <Option value="canceled">Отменено</Option>
                </Select>
              </Col>
              <Col xs={24} sm={12} md={10}>
                <Text className={isMobile ? 'text-xs' : ''}>Период:</Text>
                <RangePicker
                  style={{ width: '100%' }}
                  size={isMobile ? 'small' : 'middle'}
                  onChange={handleDateRangeChange}
                />
              </Col>
              <Col xs={24} sm={12} md={6}>
                <Button 
                  size={isMobile ? 'small' : 'middle'}
                  style={{ marginTop: isMobile ? '16px' : '0' }}
                  onClick={() => setFilters({ 
                    status: [], 
                    date_from: null, 
                    date_to: null, 
                    page: 1, 
                    page_size: 10 
                  })}
                >
                  Сбросить
                </Button>
              </Col>
            </Row>
          </Card>

          {/* Аналитика */}
          {analytics && Object.keys(analytics).length > 0 && (
            <Card size="small" title={<span className={isMobile ? 'text-sm' : ''}>Статистика по квартире</span>}>
              <Row gutter={[8, 8]}>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Всего бронирований"
                    value={analytics.total_bookings || 0}
                    prefix={<CalendarOutlined />}
                    valueStyle={{ fontSize: isMobile ? '16px' : '24px' }}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Завершенных"
                    value={analytics.completed_bookings || 0}
                    valueStyle={{ color: '#52c41a', fontSize: isMobile ? '16px' : '24px' }}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Отмененных"
                    value={analytics.canceled_bookings || 0}
                    valueStyle={{ color: '#f5222d', fontSize: isMobile ? '16px' : '24px' }}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Общий доход"
                    value={analytics.total_revenue || 0}
                    suffix="₸"
                    prefix={<DollarOutlined />}
                    valueStyle={{ color: '#52c41a', fontSize: isMobile ? '16px' : '24px' }}
                    formatter={(value) => isMobile ? `${Math.round(value / 1000)}к` : value.toLocaleString()}
                  />
                </Col>
              </Row>
              
              <Row gutter={[8, 8]} style={{ marginTop: 16 }}>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Средняя цена"
                    value={analytics.avg_price || 0}
                    suffix="₸"
                    precision={0}
                    valueStyle={{ fontSize: isMobile ? '16px' : '24px' }}
                    formatter={(value) => isMobile ? `${Math.round(value / 1000)}к` : value.toLocaleString()}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Средняя длительность"
                    value={analytics.avg_duration || 0}
                    suffix=" ч."
                    precision={1}
                    valueStyle={{ fontSize: isMobile ? '16px' : '24px' }}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Доход за час"
                    value={analytics.revenue_per_hour || 0}
                    suffix="₸"
                    precision={0}
                    valueStyle={{ fontSize: isMobile ? '16px' : '24px' }}
                    formatter={(value) => isMobile ? `${Math.round(value / 1000)}к` : value.toLocaleString()}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <div>
                    <Text type="secondary" className={isMobile ? 'text-xs' : ''}>Заполняемость</Text>
                    <Progress 
                      percent={analytics.occupancy_rate || 0} 
                      size={isMobile ? 'small' : 'default'}
                      format={(percent) => `${percent?.toFixed(1)}%`}
                    />
                  </div>
                </Col>
              </Row>
            </Card>
          )}

          {/* Таблица бронирований */}
          <Card size="small" title={<span className={isMobile ? 'text-sm' : ''}>Список бронирований</span>}>
            <Table
              columns={columns}
              dataSource={bookings}
              rowKey="id"
              loading={isLoading}
              pagination={false}
              size="small"
              scroll={{ x: isMobile ? 400 : 900 }}
            />
            
            {pagination.total > 0 && (
              <div style={{ marginTop: 16, textAlign: 'center' }}>
                <Pagination
                  current={filters.page}
                  pageSize={filters.page_size}
                  total={pagination.total}
                  showSizeChanger={!isMobile}
                  showQuickJumper={!isMobile}
                  size={isMobile ? 'small' : 'default'}
                  showTotal={(total, range) => 
                    isMobile ? 
                      `${range[0]}-${range[1]} / ${total}` :
                      `${range[0]}-${range[1]} из ${total} записей`
                  }
                  onChange={(page, pageSize) => {
                    setFilters(prev => ({
                      ...prev,
                      page,
                      page_size: pageSize
                    }));
                  }}
                />
              </div>
            )}
          </Card>
        </div>
      )}
    </Modal>
  );
};

export default ApartmentBookingHistoryModal; 