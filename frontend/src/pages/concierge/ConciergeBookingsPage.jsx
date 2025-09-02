import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { 
  Table, Card, Tag, Button, Space, Typography, Input, Select, 
  DatePicker, Row, Col, Avatar, Descriptions, Modal, Badge
} from 'antd';
import {
  CalendarOutlined,
  UserOutlined,
  SearchOutlined,
  EyeOutlined,
  MessageOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import { conciergeAPI, chatAPI } from '../../lib/api.js';
import dayjs from 'dayjs';
import { useNavigate } from 'react-router-dom';

const { Title } = Typography;
const { RangePicker } = DatePicker;
const { Option } = Select;

const ConciergeBookingsPage = () => {
  const [filters, setFilters] = useState({});
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(15);
  const [selectedBooking, setSelectedBooking] = useState(null);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const navigate = useNavigate();

  // Получение броней консьержа
  const { data: bookingsData, isLoading } = useQuery({
    queryKey: ['concierge-bookings', filters, currentPage, pageSize],
    queryFn: () => {
      const params = {
        page: currentPage,
        page_size: pageSize,
        ...filters
      };
      return conciergeAPI.getBookings(params);
    }
  });

  const getStatusColor = (status) => {
    const colors = {
      'created': 'blue',
      'confirmed': 'green',
      'active': 'cyan',
      'completed': 'gray',
      'cancelled': 'red',
      'awaiting_payment': 'orange'
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status) => {
    const texts = {
      'created': 'Создано',
      'confirmed': 'Подтверждено',
      'active': 'Активно',
      'completed': 'Завершено',
      'cancelled': 'Отменено',
      'awaiting_payment': 'Ожидает оплаты'
    };
    return texts[status] || status;
  };

  const handleViewDetails = (booking) => {
    setSelectedBooking(booking);
    setDetailsVisible(true);
  };

  const handleOpenChat = async (booking) => {
    try {
      // Попробуем найти существующий чат для этого бронирования
      const chatRoomsResponse = await chatAPI.getRooms({ booking_id: booking.id });
      const existingRoom = chatRoomsResponse.data?.rooms?.[0];
      
      if (existingRoom) {
        navigate(`/concierge/chat?room=${existingRoom.id}`);
      } else {
        // Создаем новый чат
        const newRoom = await chatAPI.createRoom({ booking_id: booking.id });
        navigate(`/concierge/chat?room=${newRoom.data.id}`);
      }
    } catch (error) {
      console.error('Ошибка при открытии чата:', error);
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (id) => `#${id}`,
    },
    {
      title: 'Даты',
      key: 'dates',
      width: 200,
      render: (_, booking) => (
        <div>
          <div>{dayjs(booking.start_date).format('DD.MM.YYYY')}</div>
          <div className="text-gray-500 text-sm">
            до {dayjs(booking.end_date).format('DD.MM.YYYY')}
          </div>
          <div className="text-gray-400 text-xs">
            {dayjs(booking.end_date).diff(dayjs(booking.start_date), 'day')} дней
          </div>
        </div>
      ),
    },
    {
      title: 'Гость',
      key: 'renter',
      width: 200,
      render: (_, booking) => (
        <div className="flex items-center space-x-2">
          <Avatar size="small" icon={<UserOutlined />} />
          <div>
            <div className="font-medium">
              {booking.renter?.user?.first_name} {booking.renter?.user?.last_name}
            </div>
            <div className="text-gray-500 text-sm">
              {booking.renter?.user?.phone}
            </div>
          </div>
        </div>
      ),
    },
    {
      title: 'Квартира',
      key: 'apartment',
      width: 200,
      render: (_, booking) => (
        <div>
          <div className="font-medium">
            {booking.apartment?.description || `Квартира #${booking.apartment_id}`}
          </div>
          <div className="text-gray-500 text-sm">
            {booking.apartment?.street}, д. {booking.apartment?.building}
          </div>
        </div>
      ),
    },
    {
      title: 'Сумма',
      dataIndex: 'total_price',
      key: 'total_price',
      width: 120,
      render: (price) => (
        <div className="font-medium">
          <DollarOutlined /> {price} ₸
        </div>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 150,
      render: (_, booking) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetails(booking)}
            title="Подробнее"
          />
          {booking.status === 'active' && (
            <Button
              type="text"
              icon={<MessageOutlined />}
              onClick={() => handleOpenChat(booking)}
              title="Открыть чат"
            />
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <Title level={2}>Бронирования</Title>
      </div>

      {/* Фильтры */}
      <Card>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={6}>
            <Input
              placeholder="Поиск по ID или гостю"
              prefix={<SearchOutlined />}
              onChange={(e) =>
                setFilters({ ...filters, search: e.target.value })
              }
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Статус"
              allowClear
              style={{ width: '100%' }}
              onChange={(value) =>
                setFilters({ ...filters, status: value })
              }
            >
              <Option value="created">Создано</Option>
              <Option value="confirmed">Подтверждено</Option>
              <Option value="active">Активно</Option>
              <Option value="completed">Завершено</Option>
              <Option value="cancelled">Отменено</Option>
              <Option value="awaiting_payment">Ожидает оплаты</Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <RangePicker
              style={{ width: '100%' }}
              placeholder={['Дата заезда от', 'Дата заезда до']}
              onChange={(dates) =>
                setFilters({
                  ...filters,
                  start_date_from: dates?.[0]?.format('YYYY-MM-DD'),
                  start_date_to: dates?.[1]?.format('YYYY-MM-DD'),
                })
              }
            />
          </Col>
        </Row>
      </Card>

      {/* Таблица бронирований */}
      <Card>
        <Table
          columns={columns}
          dataSource={bookingsData?.data?.bookings || []}
          loading={isLoading}
          rowKey="id"
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: bookingsData?.data?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} из ${total} броней`,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            },
          }}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* Модал с деталями бронирования */}
      <Modal
        title={`Бронирование #${selectedBooking?.id}`}
        open={detailsVisible}
        onCancel={() => setDetailsVisible(false)}
        footer={[
          selectedBooking?.status === 'active' && (
            <Button key="chat" type="primary" icon={<MessageOutlined />} onClick={() => {
              setDetailsVisible(false);
              handleOpenChat(selectedBooking);
            }}>
              Открыть чат
            </Button>
          ),
          <Button key="close" onClick={() => setDetailsVisible(false)}>
            Закрыть
          </Button>,
        ].filter(Boolean)}
        width={800}
      >
        {selectedBooking && (
          <div className="space-y-4">
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Card title="Информация о госте" size="small">
                  <Descriptions column={1} size="small">
                    <Descriptions.Item label="Имя">
                      {selectedBooking.renter?.user?.first_name} {selectedBooking.renter?.user?.last_name}
                    </Descriptions.Item>
                    <Descriptions.Item label="Телефон">
                      {selectedBooking.renter?.user?.phone}
                    </Descriptions.Item>
                    <Descriptions.Item label="Email">
                      {selectedBooking.renter?.user?.email}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="Информация о бронировании" size="small">
                  <Descriptions column={1} size="small">
                    <Descriptions.Item label="Даты">
                      {dayjs(selectedBooking.start_date).format('DD.MM.YYYY')} - {dayjs(selectedBooking.end_date).format('DD.MM.YYYY')}
                    </Descriptions.Item>
                    <Descriptions.Item label="Количество дней">
                      {dayjs(selectedBooking.end_date).diff(dayjs(selectedBooking.start_date), 'day')}
                    </Descriptions.Item>
                    <Descriptions.Item label="Стоимость">
                      {selectedBooking.total_price} ₸
                    </Descriptions.Item>
                    <Descriptions.Item label="Статус">
                      <Tag color={getStatusColor(selectedBooking.status)}>
                        {getStatusText(selectedBooking.status)}
                      </Tag>
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>

            <Card title="Информация о квартире" size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="Описание">
                  {selectedBooking.apartment?.description}
                </Descriptions.Item>
                <Descriptions.Item label="Адрес">
                  {selectedBooking.apartment?.street}, д. {selectedBooking.apartment?.building}
                </Descriptions.Item>
                <Descriptions.Item label="Комнат">
                  {selectedBooking.apartment?.rooms}
                </Descriptions.Item>
                <Descriptions.Item label="Площадь">
                  {selectedBooking.apartment?.square} м²
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {selectedBooking.notes && (
              <Card title="Заметки" size="small">
                <p>{selectedBooking.notes}</p>
              </Card>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default ConciergeBookingsPage;