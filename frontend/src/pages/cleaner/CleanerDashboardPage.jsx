import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, Row, Col, Statistic, Typography, List, Tag, Avatar, Spin, Button, Space } from 'antd';
import {
  CalendarOutlined,
  ToolOutlined,
  HomeOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons';
import { cleanerAPI } from '../../lib/api.js';
import NotificationList from '../../components/NotificationList.jsx';
import dayjs from 'dayjs';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const CleanerDashboardPage = () => {
  const navigate = useNavigate();

  // Получение статистики уборщицы
  const { data: stats, isLoading: statsLoading, error: statsError } = useQuery({
    queryKey: ['cleaner-stats'],
    queryFn: () => cleanerAPI.getStats(),
    retry: 1,
    onError: (error) => console.error('Ошибка загрузки статистики:', error)
  });

  // Получение профиля уборщицы
  const { data: profile, isLoading: profileLoading, error: profileError } = useQuery({
    queryKey: ['cleaner-profile'],
    queryFn: () => cleanerAPI.getProfile(),
    retry: 1,
    onError: (error) => console.error('Ошибка загрузки профиля:', error)
  });

  // Получение квартир уборщицы
  const { data: apartments, isLoading: apartmentsLoading, error: apartmentsError } = useQuery({
    queryKey: ['cleaner-apartments'],
    queryFn: () => cleanerAPI.getApartments(),
    retry: 1,
    onError: (error) => console.error('Ошибка загрузки квартир:', error)
  });

  // Получение квартир для уборки
  const { data: cleaningApartments, isLoading: cleaningLoading, error: cleaningError } = useQuery({
    queryKey: ['cleaner-apartments-for-cleaning'],
    queryFn: () => cleanerAPI.getApartmentsForCleaning(),
    retry: 1,
    onError: (error) => console.error('Ошибка загрузки квартир для уборки:', error)
  });

  const getCleaningStatusColor = (status) => {
    const colors = {
      'needs_cleaning': 'orange',
      'cleaning_in_progress': 'blue',
      'cleaned': 'green',
      'cleaning_overdue': 'red'
    };
    return colors[status] || 'default';
  };

  const getCleaningStatusText = (status) => {
    const texts = {
      'needs_cleaning': 'Нужна уборка',
      'cleaning_in_progress': 'Убирается',
      'cleaned': 'Убрано',
      'cleaning_overdue': 'Просрочено'
    };
    return texts[status] || status;
  };

  const getCleaningStatusIcon = (status) => {
    const icons = {
      'needs_cleaning': <ClockCircleOutlined />,
      'cleaning_in_progress': <PlayCircleOutlined />,
      'cleaned': <CheckCircleOutlined />,
      'cleaning_overdue': <PauseCircleOutlined />
    };
    return icons[status] || <ToolOutlined />;
  };

  if (statsLoading || profileLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  // Если есть критические ошибки, показываем их
  if (statsError || profileError) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-center">
          <Text type="danger" className="text-lg">
            Ошибка загрузки данных
          </Text>
          <br />
          <Text type="secondary">
            {statsError?.message || profileError?.message || 'Неизвестная ошибка'}
          </Text>
          <br />
          <Button 
            type="primary" 
            onClick={() => window.location.reload()}
            className="mt-4"
          >
            Обновить страницу
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Приветствие */}
      <Card>
        <div className="flex items-center space-x-4">
          <Avatar size={64} icon={<ToolOutlined />} className="bg-green-500" />
          <div>
            <Title level={2} className="!mb-1">
              Добро пожаловать, {profile?.data?.user?.first_name || 'в Админ панель'}!
            </Title>
            <Text type="secondary">
              Уборщица • {apartments?.data?.length ? `${apartments.data.length} квартир(ы)` : 'Без квартир'}
            </Text>
            <br />
            <Text type="secondary">
              {Array.isArray(apartments?.data) && apartments.data.length > 0 ? 
                apartments.data.map((apt, index) => 
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
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="Всего квартир"
              value={stats?.data?.total_apartments || 0}
              prefix={<HomeOutlined className="text-blue-500" />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="Нужна уборка"
              value={stats?.data?.apartments_needing_cleaning || 0}
              prefix={<ClockCircleOutlined className="text-orange-500" />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="Убрано сегодня"
              value={stats?.data?.cleaned_today || 0}
              prefix={<CheckCircleOutlined className="text-green-500" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card className="text-center">
            <Statistic
              title="Убрано за месяц"
              value={stats?.data?.cleaned_this_month || 0}
              prefix={<ToolOutlined className="text-purple-500" />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* Квартиры для уборки */}
        <Col xs={24} lg={12}>
          <Card 
            title="Квартиры для уборки"
            extra={
              <Button 
                type="primary" 
                size="small"
                onClick={() => navigate('/cleaner/apartments-for-cleaning')}
              >
                Все
              </Button>
            }
            className="h-full"
          >
            {cleaningLoading ? (
              <div className="flex justify-center py-8">
                <Spin />
              </div>
            ) : (
              <List
                dataSource={cleaningApartments?.data?.slice(0, 5) || []}
                locale={{ emptyText: 'Нет квартир для уборки' }}
                renderItem={(apartment) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          icon={<HomeOutlined />} 
                          className="bg-orange-500"
                        />
                      }
                      title={
                        <div className="flex items-center justify-between">
                          <span>
                            {apartment.street}, д. {apartment.building}
                            {apartment.apartment_number && `, кв. ${apartment.apartment_number}`}
                          </span>
                          <Tag 
                            icon={getCleaningStatusIcon(apartment.cleaning_status)}
                            color={getCleaningStatusColor(apartment.cleaning_status)}
                          >
                            {getCleaningStatusText(apartment.cleaning_status)}
                          </Tag>
                        </div>
                      }
                      description={
                        <div className="text-sm text-gray-500">
                          {apartment.last_cleaning_date && (
                            <div>
                              Последняя уборка: {dayjs(apartment.last_cleaning_date).format('DD.MM.YYYY')}
                            </div>
                          )}
                          {apartment.next_cleaning_date && (
                            <div>
                              Следующая уборка: {dayjs(apartment.next_cleaning_date).format('DD.MM.YYYY')}
                            </div>
                          )}
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
          </Card>
        </Col>

        {/* Мои квартиры */}
        <Col xs={24} lg={12}>
          <Card 
            title="Мои квартиры"
            extra={
              <Button 
                type="primary" 
                size="small"
                onClick={() => navigate('/cleaner/apartments')}
              >
                Все
              </Button>
            }
            className="h-full"
          >
            {apartmentsLoading ? (
              <div className="flex justify-center py-8">
                <Spin />
              </div>
            ) : (
              <List
                dataSource={apartments?.data?.slice(0, 5) || []}
                locale={{ emptyText: 'Нет назначенных квартир' }}
                renderItem={(apartment) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          icon={<HomeOutlined />} 
                          className="bg-blue-500"
                        />
                      }
                      title={
                        <div className="flex items-center justify-between">
                          <span>
                            {apartment.street}, д. {apartment.building}
                            {apartment.apartment_number && `, кв. ${apartment.apartment_number}`}
                          </span>
                          <Tag color={apartment.is_free ? 'green' : 'red'}>
                            {apartment.is_free ? 'Свободна' : 'Занята'}
                          </Tag>
                        </div>
                      }
                      description={
                        <div className="text-sm text-gray-500">
                          <div>Тип: {apartment.apartment_type?.name || 'Не указан'}</div>
                          <div>Район: {apartment.district?.name || 'Не указан'}</div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* Быстрые действия */}
      <Card title="Быстрые действия">
        <Space wrap>
          <Button 
            type="primary" 
            icon={<ToolOutlined />}
            onClick={() => navigate('/cleaner/apartments-for-cleaning')}
          >
            Начать уборку
          </Button>
          <Button 
            icon={<HomeOutlined />}
            onClick={() => navigate('/cleaner/apartments')}
          >
            Мои квартиры
          </Button>
          <Button 
            icon={<CalendarOutlined />}
            onClick={() => navigate('/cleaner/schedule')}
          >
            Расписание
          </Button>
        </Space>
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
                  onClick={() => navigate('/cleaner/notifications')}
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

export default CleanerDashboardPage;
