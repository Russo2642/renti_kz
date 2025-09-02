import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, Row, Col, Statistic, Typography, List, Tag, Avatar, Spin, Button } from 'antd';
import {
  CalendarOutlined,
  MessageOutlined,
  HomeOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { conciergeAPI } from '../../lib/api.js';
import NotificationList from '../../components/NotificationList.jsx';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

const ConciergeDashboardPage = () => {
  const navigate = useNavigate();
  
  // Получение статистики консьержа
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['concierge-stats'],
    queryFn: () => conciergeAPI.getStats()
  });

  // Получение профиля консьержа
  const { data: profile, isLoading: profileLoading } = useQuery({
    queryKey: ['concierge-profile'],
    queryFn: () => conciergeAPI.getProfile()
  });

  // Получение активных броней
  const { data: activeBookings, isLoading: bookingsLoading } = useQuery({
    queryKey: ['concierge-active-bookings'],
    queryFn: () => conciergeAPI.getBookings({ status: 'confirmed', active: true })
  });

  // Получение активных чатов
  const { data: activeChats, isLoading: chatsLoading } = useQuery({
    queryKey: ['concierge-active-chats'],
    queryFn: () => conciergeAPI.getChatRooms({ status: 'active' })
  });

  const getStatusColor = (status) => {
    const colors = {
      'confirmed': 'green',
      'active': 'blue',
      'completed': 'gray',
      'cancelled': 'red'
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status) => {
    const texts = {
      'confirmed': 'Подтверждено',
      'active': 'Активно',
      'completed': 'Завершено',
      'cancelled': 'Отменено'
    };
    return texts[status] || status;
  };

  if (profileLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Приветствие */}
      <Card>
        <div className="flex items-center space-x-4">
          <Avatar size={64} icon={<HomeOutlined />} className="bg-blue-500" />
          <div>
            <Title level={2} className="!mb-1">
              Добро пожаловать, {profile?.data?.user?.first_name} {profile?.data?.user?.last_name}!
            </Title>
            <Text type="secondary">
              Консьерж • {profile?.data?.apartments?.length ? `${profile.data.apartments.length} квартир(ы)` : 'Без квартир'}
            </Text>
            <br />
            <Text type="secondary">
              {profile?.data?.apartments?.length > 0 ? 
                profile.data.apartments.map((apt, index) => 
                  `${apt.street}, д. ${apt.building}`
                ).join('; ') 
                : 'Квартиры не назначены'
              }
            </Text>
          </div>
        </div>
      </Card>

      {/* Статистика */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Активные брони"
              value={stats?.data?.active_bookings || 0}
              prefix={<CalendarOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Активные чаты"
              value={stats?.data?.active_chats || 0}
              prefix={<MessageOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Всего броней сегодня"
              value={stats?.data?.today_bookings || 0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Завершённые чаты"
              value={stats?.data?.completed_chats || 0}
              prefix={<MessageOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* Активные бронирования */}
        <Col xs={24} lg={12}>
          <Card 
            title="Активные бронирования" 
            extra={<Text type="secondary">Последние 5</Text>}
            loading={bookingsLoading}
          >
            <List
              dataSource={activeBookings?.data?.bookings?.slice(0, 5) || []}
              renderItem={(booking) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar icon={<CalendarOutlined />} />}
                    title={
                      <div className="flex items-center justify-between">
                        <span>Бронь #{booking.id}</span>
                        <Tag color={getStatusColor(booking.status)}>
                          {getStatusText(booking.status)}
                        </Tag>
                      </div>
                    }
                    description={
                      <div>
                        <div>
                          {dayjs(booking.start_date).format('DD.MM.YYYY')} - {' '}
                          {dayjs(booking.end_date).format('DD.MM.YYYY')}
                        </div>
                        <div className="text-gray-500">
                          Гость: {booking.renter?.user?.first_name} {booking.renter?.user?.last_name}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* Активные чаты */}
        <Col xs={24} lg={12}>
          <Card 
            title="Активные чаты" 
            extra={<Text type="secondary">Требуют внимания</Text>}
            loading={chatsLoading}
          >
            <List
              dataSource={activeChats?.data?.rooms?.slice(0, 5) || []}
              renderItem={(chat) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar icon={<MessageOutlined />} />}
                    title={
                      <div className="flex items-center justify-between">
                        <span>Чат #{chat.id}</span>
                        {chat.unread_count > 0 && (
                          <Tag color="red">{chat.unread_count} новых</Tag>
                        )}
                      </div>
                    }
                    description={
                      <div>
                        <div>
                          Бронь #{chat.booking_id}
                        </div>
                        <div className="text-gray-500">
                          {chat.renter?.user?.first_name} {chat.renter?.user?.last_name}
                        </div>
                        {chat.last_message && (
                          <div className="text-gray-400 text-sm mt-1">
                            Последнее сообщение: {dayjs(chat.last_message_at).format('DD.MM HH:mm')}
                          </div>
                        )}
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

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
                  onClick={() => navigate('/concierge/notifications')}
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

export default ConciergeDashboardPage;