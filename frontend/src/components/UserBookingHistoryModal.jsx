import { CalendarOutlined, DollarOutlined, HistoryOutlined, HomeOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { 
  Button, Card, Col, DatePicker, Descriptions, Modal, Row, 
  Select, Space, Statistic, Table, Tag, Typography, Pagination 
} from 'antd';
import dayjs from 'dayjs';
import React, { useState, useEffect } from 'react';
import { usersAPI } from '../lib/api.js';

const { RangePicker } = DatePicker;
const { Option } = Select;
const { Title, Text } = Typography;

const UserBookingHistoryModal = ({ visible, onClose, user }) => {
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [filters, setFilters] = useState({
    role: 'renter',
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

  // Загрузка истории бронирований
  const { data: historyData, isLoading } = useQuery({
    queryKey: ['userBookingHistory', user?.id, filters],
    queryFn: () => usersAPI.adminGetUserBookingHistory(user?.id, filters),
    enabled: visible && !!user?.id,
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
      title: filters.role === 'owner' ? 'Арендатор' : 'Квартира',
      key: 'main_info',
      width: isMobile ? 120 : 200,
      render: (record) => (
        <div className={isMobile ? 'text-xs' : ''}>
          {filters.role === 'owner' ? (
            <div>
              <Text strong>{record.renter?.user?.first_name} {record.renter?.user?.last_name}</Text>
              <br />
              <Text type="secondary">{record.renter?.user?.phone}</Text>
            </div>
          ) : (
            <div>
              <Text strong>
                {record.apartment?.description ? 
                  record.apartment.description.substring(0, isMobile ? 30 : 50) + '...' :
                  `${record.apartment?.street}, ${record.apartment?.building}-${record.apartment?.apartment_number}`
                }
              </Text>
              <br />
              <Text type="secondary">{record.apartment?.room_count} комн.</Text>
            </div>
          )}
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

  return (
    <Modal
      title={
        <Space>
          <HistoryOutlined />
          <span className={isMobile ? 'text-sm' : ''}>
            История бронирований: {user?.first_name} {user?.last_name}
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
      {user && (
        <div className="space-y-4">
          {/* Информация о пользователе */}
          <Card size="small">
            <Descriptions column={isMobile ? 1 : 3} size="small">
              <Descriptions.Item label="Email">{user.email}</Descriptions.Item>
              <Descriptions.Item label="Телефон">{user.phone}</Descriptions.Item>
              <Descriptions.Item label="Роль">{user.role}</Descriptions.Item>
            </Descriptions>
          </Card>

          {/* Фильтры */}
          <Card size="small" title={<span className={isMobile ? 'text-sm' : ''}>Фильтры</span>}>
            <Row gutter={[8, 8]}>
              <Col xs={24} sm={12} md={6}>
                <Text className={isMobile ? 'text-xs' : ''}>Роль:</Text>
                <Select
                  value={filters.role}
                  style={{ width: '100%' }}
                  size={isMobile ? 'small' : 'middle'}
                  onChange={(value) => handleFilterChange('role', value)}
                >
                  <Option value="renter">Как арендатор</Option>
                  <Option value="owner">Как владелец</Option>
                </Select>
              </Col>
              <Col xs={24} sm={12} md={6}>
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
              <Col xs={24} sm={12} md={8}>
                <Text className={isMobile ? 'text-xs' : ''}>Период:</Text>
                <RangePicker
                  style={{ width: '100%' }}
                  size={isMobile ? 'small' : 'middle'}
                  onChange={handleDateRangeChange}
                />
              </Col>
              <Col xs={24} sm={12} md={4}>
                <Button 
                  size={isMobile ? 'small' : 'middle'}
                  style={{ marginTop: isMobile ? '16px' : '0' }}
                  onClick={() => setFilters({ 
                    role: 'renter', 
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
            <Card size="small" title={<span className={isMobile ? 'text-sm' : ''}>Статистика</span>}>
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
                  {filters.role === 'renter' ? (
                    <Statistic
                      title="Потрачено"
                      value={analytics.total_spent || 0}
                      suffix="₸"
                      prefix={<DollarOutlined />}
                      valueStyle={{ color: '#f5222d', fontSize: isMobile ? '16px' : '24px' }}
                      formatter={(value) => isMobile ? `${Math.round(value / 1000)}к` : value.toLocaleString()}
                    />
                  ) : (
                    <Statistic
                      title="Заработано"
                      value={analytics.total_earned || 0}
                      suffix="₸"
                      prefix={<DollarOutlined />}
                      valueStyle={{ color: '#52c41a', fontSize: isMobile ? '16px' : '24px' }}
                      formatter={(value) => isMobile ? `${Math.round(value / 1000)}к` : value.toLocaleString()}
                    />
                  )}
                </Col>
                <Col xs={12} sm={6}>
                  {filters.role === 'renter' ? (
                    <Statistic
                      title="Лояльность"
                      value={analytics.loyalty_score || 0}
                      suffix="%"
                      precision={1}
                      valueStyle={{ fontSize: isMobile ? '16px' : '24px' }}
                    />
                  ) : (
                    <Statistic
                      title="Квартир"
                      value={analytics.apartments_count || 0}
                      prefix={<HomeOutlined />}
                      valueStyle={{ fontSize: isMobile ? '16px' : '24px' }}
                    />
                  )}
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
              scroll={{ x: isMobile ? 400 : 800 }}
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

export default UserBookingHistoryModal; 