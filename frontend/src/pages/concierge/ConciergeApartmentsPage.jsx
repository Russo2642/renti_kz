import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, Row, Col, Typography, Tag, Descriptions, Badge, Spin, Empty } from 'antd';
import {
  HomeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DollarOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { conciergeAPI } from '../../lib/api.js';

const { Title, Text } = Typography;

const ConciergeApartmentsPage = () => {
  // Получение квартир консьержа
  const { data: apartments, isLoading } = useQuery({
    queryKey: ['concierge-apartments'],
    queryFn: () => conciergeAPI.getApartments()
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

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  const apartmentData = apartments?.data || [];

  if (!apartmentData || apartmentData.length === 0) {
    return (
      <Card>
        <Empty 
          description="У вас нет закрепленных квартир"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <Title level={2}>Мои квартиры</Title>
        <Text type="secondary">
          Всего квартир: {apartmentData.length}
        </Text>
      </div>

      <Row gutter={[16, 16]}>
        {apartmentData.map((apartment) => (
          <Col xs={24} lg={12} xl={8} key={apartment.id}>
            <Card
              title={
                <div className="flex items-center space-x-2">
                  <HomeOutlined />
                  <span>Квартира #{apartment.id}</span>
                  <Tag color={getStatusColor(apartment.is_free)}>
                    {getStatusText(apartment.is_free)}
                  </Tag>
                </div>
              }
              extra={getStatusIcon(apartment.is_free)}
              className="h-full"
            >
              <div className="space-y-4">
                {/* Основная информация */}
                <Descriptions size="small" column={1}>
                  <Descriptions.Item label="Описание">
                    {apartment.description || 'Не указано'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Адрес">
                    {apartment.street}, д. {apartment.building}
                    {apartment.apartment_number && `, кв. ${apartment.apartment_number}`}
                  </Descriptions.Item>
                  <Descriptions.Item label="Город">
                    {apartment.city?.name || 'Не указан'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Цена за сутки">
                    <Text strong>
                      <DollarOutlined /> {apartment.price} ₸
                    </Text>
                  </Descriptions.Item>
                  {apartment.owner && (
                    <Descriptions.Item label="Владелец">
                      <div className="flex items-center space-x-1">
                        <UserOutlined />
                        <span>
                          {apartment.owner.user?.first_name} {apartment.owner.user?.last_name}
                        </span>
                      </div>
                    </Descriptions.Item>
                  )}
                </Descriptions>

                {/* Дополнительная информация */}
                {apartment.rooms && (
                  <div>
                    <Text type="secondary">Комнат: </Text>
                    <Text>{apartment.rooms}</Text>
                  </div>
                )}

                {apartment.square && (
                  <div>
                    <Text type="secondary">Площадь: </Text>
                    <Text>{apartment.square} м²</Text>
                  </div>
                )}

                {apartment.floor && (
                  <div>
                    <Text type="secondary">Этаж: </Text>
                    <Text>{apartment.floor}</Text>
                  </div>
                )}

                {/* Удобства */}
                {apartment.amenities && apartment.amenities.length > 0 && (
                  <div>
                    <Text type="secondary" className="block mb-2">Удобства:</Text>
                    <div className="flex flex-wrap gap-1">
                      {apartment.amenities.map((amenity, index) => (
                        <Tag key={index} size="small">
                          {amenity}
                        </Tag>
                      ))}
                    </div>
                  </div>
                )}

                {/* Правила */}
                {apartment.rules && apartment.rules.length > 0 && (
                  <div>
                    <Text type="secondary" className="block mb-2">Правила:</Text>
                    <div className="flex flex-wrap gap-1">
                      {apartment.rules.map((rule, index) => (
                        <Tag key={index} color="blue" size="small">
                          {rule}
                        </Tag>
                      ))}
                    </div>
                  </div>
                )}

                {/* Статус активности */}
                <div className="pt-2 border-t border-gray-100">
                  <div className="flex items-center justify-between">
                    <Text type="secondary">Статус активности:</Text>
                    <Badge 
                      status={apartment.is_active ? "success" : "error"} 
                      text={apartment.is_active ? "Активна" : "Неактивна"} 
                    />
                  </div>
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
};

export default ConciergeApartmentsPage;