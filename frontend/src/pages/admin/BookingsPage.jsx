import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Table, Button, Card, Tag, Modal, Form, Input, Select, Space,
  Row, Col, Statistic, Drawer, Descriptions, message, Popconfirm,
  DatePicker, Typography, Badge, Tooltip, Timeline
} from 'antd';
import {
  EditOutlined, DeleteOutlined, EyeOutlined, CheckOutlined,
  CloseOutlined, CalendarOutlined, DollarOutlined, UserOutlined,
  EnvironmentOutlined, ClockCircleOutlined
} from '@ant-design/icons';
import { bookingsAPI, contractsAPI } from '../../lib/api.js';
import UserDocumentsModal from '../../components/UserDocumentsModal.jsx';
import PaymentReceiptModal from '../../components/PaymentReceiptModal.jsx';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;
const { RangePicker } = DatePicker;

const BookingsPage = () => {
  const [filters, setFilters] = useState({});
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [selectedBooking, setSelectedBooking] = useState(null);
  const [statusModalVisible, setStatusModalVisible] = useState(false);
  const [documentsModalVisible, setDocumentsModalVisible] = useState(false);
  const [receiptModalVisible, setReceiptModalVisible] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [statusForm] = Form.useForm();
  const queryClient = useQueryClient();

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Мутация для обновления статуса бронирования
  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status, reason }) => bookingsAPI.adminUpdateBookingStatus(id, status, reason),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-bookings']);
      queryClient.invalidateQueries(['admin-bookings-statistics']);
      setStatusModalVisible(false);
      statusForm.resetFields();
      message.success('Статус бронирования обновлен');
    },
    onError: () => {
      message.error('Ошибка при обновлении статуса');
    }
  });

  // Мутация для отмены бронирования
  const cancelBookingMutation = useMutation({
    mutationFn: bookingsAPI.adminCancelBooking,
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-bookings']);
      message.success('Бронирование отменено');
    },
    onError: () => {
      message.error('Ошибка при отмене бронирования');
    }
  });

  const handleStatusUpdate = (values) => {
    updateStatusMutation.mutate({
      id: selectedBooking.id,
      ...values
    });
  };

  const handleCancelBooking = (id) => {
    cancelBookingMutation.mutate(id);
  };

  // Получение бронирований (админская версия)
  const { data: bookingsData, isLoading } = useQuery({
    queryKey: ['admin-bookings', filters, currentPage, pageSize],
    queryFn: () => {
      const params = {
        page: currentPage,
        page_size: pageSize,
        ...filters
      };
      return bookingsAPI.adminGetAllBookings(params);
    }
  });

  // Получение детальной статистики бронирований
  const { data: bookingsStatistics, isLoading: isLoadingStatistics } = useQuery({
    queryKey: ['admin-bookings-statistics'],
    queryFn: bookingsAPI.adminGetBookingsStatistics,
    staleTime: 5 * 60 * 1000, // 5 минут
  });

  const getStatusColor = (status) => {
    const colors = {
      'created': 'blue',
      'pending': 'orange',
      'approved': 'green',
      'rejected': 'red',
      'active': 'cyan',
      'completed': 'gray',
      'canceled': 'red'
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status) => {
    const texts = {
      'created': 'Создано',
      'pending': 'На рассмотрении',
      'approved': 'Одобрено',
      'rejected': 'Отклонено',
      'active': 'Активно',
      'completed': 'Завершено',
      'canceled': 'Отменено'
    };
    return texts[status] || status;
  };

  const calculateDuration = (startDate, endDate) => {
    const start = dayjs(startDate);
    const end = dayjs(endDate);
    const hours = end.diff(start, 'hour');
    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;
    
    if (days > 0) {
      return `${days} дн. ${remainingHours} ч.`;
    }
    return `${remainingHours} ч.`;
  };

  const handleViewContract = async (bookingId) => {
    try {
      // Сначала получаем ID договора через booking
      const contractResponse = await contractsAPI.getByBookingId(bookingId);
      const contractId = contractResponse.data.id; // Исправлен путь
      
      // Затем получаем HTML договора
      const htmlResponse = await contractsAPI.getContractHTML(contractId);
      const htmlContent = htmlResponse.data.html; // Исправлен путь
      
      const newWindow = window.open('', '_blank');
      newWindow.document.write(htmlContent);
      newWindow.document.close();
    } catch (error) {
      message.error('Ошибка при загрузке договора');
      console.error('Contract error:', error);
    }
  };

  const columns = [
    {
      title: 'Номер',
      dataIndex: 'booking_number',
      key: 'booking_number',
      width: 120,
      render: (number) => (
        <Text strong>#{number}</Text>
      ),
    },
    {
      title: 'Квартира',
      key: 'apartment',
      render: (_, record) => (
        <div>
          <div className="font-medium">
            {record.apartment?.street}, кв. {record.apartment?.apartment_number}
          </div>
          <div className="text-gray-500 text-sm">
            {record.apartment?.room_count}-комн., {record.apartment?.area} м²
          </div>
        </div>
      ),
    },
    {
      title: 'Арендатор',
      key: 'renter',
      render: (_, record) => (
        <div>
          <div>
            {record.renter?.user?.first_name} {record.renter?.user?.last_name}
          </div>
          <div className="text-gray-500 text-sm">{record.renter?.user?.phone}</div>
        </div>
      ),
    },
    {
      title: 'Период',
      key: 'period',
      render: (_, record) => (
        <div>
          <div className="text-sm">
            {dayjs(record.start_date).format('DD.MM.YYYY HH:mm')}
          </div>
          <div className="text-sm">
            {dayjs(record.end_date).format('DD.MM.YYYY HH:mm')}
          </div>
          <div className="text-xs text-gray-500">
            {calculateDuration(record.start_date, record.end_date)}
          </div>
        </div>
      ),
    },
    {
      title: 'Стоимость',
      key: 'amount',
      render: (_, record) => (
        <div>
          <Text strong>{record.final_price?.toLocaleString()} ₸</Text>
          <div className="text-xs text-gray-500">
            Базовая: {record.total_price?.toLocaleString()} ₸
          </div>
          {record.service_fee > 0 && (
            <div className="text-xs text-gray-500">
              Сервис: {record.service_fee?.toLocaleString()} ₸
            </div>
          )}
        </div>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: 'Создано',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD.MM HH:mm'),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="Просмотр">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedBooking(record);
                setDetailsVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Изменить статус">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedBooking(record);
                setStatusModalVisible(true);
                statusForm.setFieldsValue({
                  status: record.status
                });
              }}
            />
          </Tooltip>
          {record.status !== 'canceled' && record.status !== 'completed' && (
            <Tooltip title="Отменить">
              <Popconfirm
                title="Отменить бронирование?"
                description="Это действие нельзя отменить"
                onConfirm={() => handleCancelBooking(record.id)}
                okText="Да"
                cancelText="Нет"
              >
                <Button
                  type="text"
                  danger
                  icon={<CloseOutlined />}
                />
              </Popconfirm>
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];



  return (
    <div className="space-y-6">
      <div className="mb-6">
        <Title level={2}>Управление бронированиями</Title>
        <Text type="secondary">
          Просмотр и отслеживание всех бронирований в системе
        </Text>
      </div>

      {/* Статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Всего бронирований"
              value={bookingsStatistics?.data?.summary?.total_bookings || 0}
              loading={isLoadingStatistics}
              prefix={<CalendarOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setFilters(prev => ({ ...prev, status: 'completed', page: 1 }));
            }}
          >
            <Statistic
              title="Завершено"
              value={bookingsStatistics?.data?.by_status?.completed || 0}
              loading={isLoadingStatistics}
              prefix={<Badge status="success" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setFilters(prev => ({ ...prev, status: 'canceled', page: 1 }));
            }}
          >
            <Statistic
              title="Отменено"
              value={bookingsStatistics?.data?.by_status?.canceled || 0}
              loading={isLoadingStatistics}
              prefix={<Badge status="error" />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card 
            hoverable 
            className="cursor-pointer"
            onClick={() => {
              setFilters(prev => ({ ...prev, status: 'rejected', page: 1 }));
            }}
          >
            <Statistic
              title="Отклонено"
              value={bookingsStatistics?.data?.by_status?.rejected || 0}
              loading={isLoadingStatistics}
              prefix={<Badge status="warning" />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Дополнительная статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Общий доход"
              value={bookingsStatistics?.data?.summary?.total_revenue || 0}
              loading={isLoadingStatistics}
              prefix={<DollarOutlined />}
              suffix="₸"
              formatter={(value) => value.toLocaleString()}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Средняя цена"
              value={bookingsStatistics?.data?.summary?.avg_price || 0}
              loading={isLoadingStatistics}
              prefix={<DollarOutlined />}
              suffix="₸"
              formatter={(value) => value.toLocaleString()}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Средняя длительность"
              value={bookingsStatistics?.data?.summary?.avg_duration || 0}
              loading={isLoadingStatistics}
              suffix=" дн."
              precision={1}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <Card>
            <Statistic
              title="Двери закрыты"
              value={bookingsStatistics?.data?.by_door_status?.closed || 0}
              loading={isLoadingStatistics}
              prefix={<Badge status="default" />}
              valueStyle={{ color: '#8c8c8c' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Фильтры */}
      <Card className="mb-6">
        <Form 
          layout="vertical"
          className="grid grid-cols-1 md:grid-cols-4 gap-4"
        >
          <Form.Item label="Статус" className="mb-0">
            <Select
              placeholder="Выберите статус"
              allowClear
              value={filters.status}
              onChange={(value) => setFilters({ ...filters, status: value })}
            >
              <Option value="created">Создано</Option>
              <Option value="pending">На проверке</Option>
              <Option value="approved">Одобрено</Option>
              <Option value="active">Активно</Option>
              <Option value="completed">Завершено</Option>
              <Option value="canceled">Отменено</Option>
              <Option value="rejected">Отклонено</Option>
            </Select>
          </Form.Item>
          <Form.Item label="Период заезда" className="mb-0">
            <RangePicker
              className="w-full"
              onChange={(dates) => {
                if (dates) {
                  setFilters({
                    ...filters,
                    date_from: dates[0].format('YYYY-MM-DD'),
                    date_to: dates[1].format('YYYY-MM-DD')
                  });
                } else {
                  const { date_from, date_to, ...rest } = filters;
                  setFilters(rest);
                }
              }}
            />
          </Form.Item>
          <Form.Item label=" " className="mb-0">
            <Button 
              onClick={() => setFilters({})}
              className="w-full"
            >
              Сбросить
            </Button>
          </Form.Item>
        </Form>
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
            showSizeChanger: !isMobile,
            showQuickJumper: !isMobile,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} бронирований`,
            responsive: true,
            simple: isMobile,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            },
          }}
          scroll={{ x: 1400 }}
          size={isMobile ? 'small' : 'default'}
        />
      </Card>

      {/* Модал деталей бронирования */}
      <Drawer
        title="Детали бронирования"
        width={isMobile ? '100%' : 720}
        open={detailsVisible}
        onClose={() => setDetailsVisible(false)}
        placement={isMobile ? 'bottom' : 'right'}
        height={isMobile ? '90%' : undefined}
      >
        {selectedBooking && (
          <div>
            <Descriptions 
              column={isMobile ? 1 : 2} 
              bordered
              size={isMobile ? 'small' : 'default'}
            >
              <Descriptions.Item label="Номер бронирования" span={isMobile ? 1 : 2}>
                <Text strong>#{selectedBooking.booking_number}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Статус" span={isMobile ? 1 : 2}>
                <Tag color={getStatusColor(selectedBooking.status)} size="large">
                  {getStatusText(selectedBooking.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Квартира" span={isMobile ? 1 : 2}>
                <div className="break-words">
                  {selectedBooking.apartment?.street}, кв. {selectedBooking.apartment?.apartment_number}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Адрес" span={isMobile ? 1 : 2}>
                <div className="break-words">
                  г. {selectedBooking.apartment?.city?.name}, {selectedBooking.apartment?.district?.name}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Арендатор" span={isMobile ? 1 : 2}>
                <div className="break-words">
                  {selectedBooking.renter?.user?.first_name} {selectedBooking.renter?.user?.last_name}
                  <br />
                  <Text type="secondary" className="text-sm">
                    {selectedBooking.renter?.user?.phone}
                  </Text>
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Дата заезда">
                <div className="font-mono text-sm">
                  {dayjs(selectedBooking.start_date).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedBooking.start_date).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Дата выезда">
                <div className="font-mono text-sm">
                  {dayjs(selectedBooking.end_date).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedBooking.end_date).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Продолжительность" span={isMobile ? 1 : 2}>
                {calculateDuration(selectedBooking.start_date, selectedBooking.end_date)}
              </Descriptions.Item>
              <Descriptions.Item label="Базовая стоимость">
                <Text className="font-mono">
                  {selectedBooking.total_price?.toLocaleString()} ₸
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="Сервисный сбор">
                <Text className="font-mono">
                  {selectedBooking.service_fee?.toLocaleString()} ₸
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="Итоговая стоимость" span={isMobile ? 1 : 2}>
                <Text strong className="font-mono text-lg">
                  {selectedBooking.final_price?.toLocaleString()} ₸
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="Создано">
                <div className="font-mono text-sm">
                  {dayjs(selectedBooking.created_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedBooking.created_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Обновлено">
                <div className="font-mono text-sm">
                  {dayjs(selectedBooking.updated_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedBooking.updated_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              {selectedBooking.comment && (
                <Descriptions.Item label="Комментарий" span={isMobile ? 1 : 2}>
                  <div className="break-words">
                    {selectedBooking.comment}
                  </div>
                </Descriptions.Item>
              )}
              {selectedBooking.rejection_reason && (
                <Descriptions.Item label="Причина отклонения" span={isMobile ? 1 : 2}>
                  <Text type="danger" className="break-words">
                    {selectedBooking.rejection_reason}
                  </Text>
                </Descriptions.Item>
              )}

            </Descriptions>

            {/* Дополнительная информация */}
            <div className="mt-6">
              <Title level={4}>Дополнительная информация</Title>
              <Row gutter={[16, 16]}>
                <Col xs={24} sm={24} md={12}>
                  <Card size="small" title="Информация о квартире">
                    <div className="space-y-2">
                      <div>Комнат: {selectedBooking.apartment?.room_count}</div>
                      <div>Площадь: {selectedBooking.apartment?.area} м²</div>
                      <div>Этаж: {selectedBooking.apartment?.floor}/{selectedBooking.apartment?.total_floors}</div>
                      <div>Состояние: {selectedBooking.apartment?.condition?.name}</div>
                    </div>
                  </Card>
                </Col>
                <Col xs={24} sm={24} md={12}>
                  <Card size="small" title="Контактная информация">
                    <div className="space-y-3">
                      <div>
                        <Text strong>Владелец:</Text>
                        <div className="mt-1 text-sm break-words">
                          {selectedBooking.apartment?.owner?.user?.first_name} {selectedBooking.apartment?.owner?.user?.last_name}
                        </div>
                        <div className="text-sm break-words">
                          {selectedBooking.apartment?.owner?.user?.phone}
                        </div>
                      </div>
                      <div>
                        <Text strong>Арендатор:</Text>
                        <div className="mt-1 text-sm break-words">
                          {selectedBooking.renter?.user?.first_name} {selectedBooking.renter?.user?.last_name}
                        </div>
                        <div className="text-sm break-words">
                          {selectedBooking.renter?.user?.phone}
                        </div>
                        <div className="text-sm break-words text-gray-500">
                          {selectedBooking.renter?.user?.email}
                        </div>
                      </div>
                    </div>
                  </Card>
                </Col>
              </Row>
            </div>

            {/* Действия с документами */}
            <div className="mt-6">
              <Title level={4}>Документы и чеки</Title>
              <Space direction={isMobile ? 'vertical' : 'horizontal'} wrap className={isMobile ? 'w-full' : ''}>
                <Button 
                  icon={<UserOutlined />}
                  onClick={() => setDocumentsModalVisible(true)}
                  className={isMobile ? 'w-full' : ''}
                >
                  Документы пользователя
                </Button>
                {(selectedBooking.status === 'approved' || selectedBooking.status === 'active' || selectedBooking.status === 'completed') && (
                  <>
                    <Button 
                      type="primary"
                      onClick={() => handleViewContract(selectedBooking.id)}
                      className={`bg-blue-600 hover:bg-blue-700 border-blue-600 hover:border-blue-700 ${isMobile ? 'w-full' : ''}`}
                    >
                      📄 Просмотреть договор
                    </Button>
                    <Button 
                      icon={<DollarOutlined />}
                      onClick={() => setReceiptModalVisible(true)}
                      className={isMobile ? 'w-full' : ''}
                    >
                      Чек об оплате
                    </Button>
                  </>
                )}
              </Space>
            </div>

            {/* Timeline статусов */}
            <div className="mt-6">
              <Title level={4}>История изменений</Title>
              <Timeline
                items={[
                  {
                    dot: <CalendarOutlined className="timeline-clock-icon" />,
                    color: 'blue',
                    children: (
                      <div>
                        <Text strong>Бронирование создано</Text>
                        <div className="text-gray-500">
                          {dayjs(selectedBooking.created_at).format('DD.MM.YYYY HH:mm')}
                        </div>
                      </div>
                    )
                  },
                  ...(selectedBooking.status !== 'created' ? [{
                    color: getStatusColor(selectedBooking.status),
                    children: (
                      <div>
                        <Text strong>Статус изменен на: {getStatusText(selectedBooking.status)}</Text>
                        <div className="text-gray-500">
                          {dayjs(selectedBooking.updated_at).format('DD.MM.YYYY HH:mm')}
                        </div>
                      </div>
                    )
                  }] : [])
                ]}
              />
            </div>
          </div>
        )}
      </Drawer>

      {/* Модальное окно изменения статуса */}
      <Modal
        title="Изменение статуса бронирования"
        open={statusModalVisible}
        onOk={() => statusForm.submit()}
        onCancel={() => {
          setStatusModalVisible(false);
          statusForm.resetFields();
        }}
        okText="Сохранить"
        cancelText="Отмена"
        width={isMobile ? '95%' : 520}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={statusForm}
          layout="vertical"
          onFinish={handleStatusUpdate}
        >
          <Form.Item
            name="status"
            label="Новый статус"
            rules={[{ required: true, message: 'Выберите статус' }]}
          >
            <Select>
              <Option value="created">Создано</Option>
              <Option value="pending">На рассмотрении</Option>
              <Option value="approved">Одобрено</Option>
              <Option value="active">Активно</Option>
              <Option value="completed">Завершено</Option>
              <Option value="canceled">Отменено</Option>
              <Option value="rejected">Отклонено</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="reason"
            label="Причина изменения (опционально)"
          >
            <TextArea
              rows={3}
              placeholder="Укажите причину изменения статуса"
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* Модал документов пользователя */}
      <UserDocumentsModal
        visible={documentsModalVisible}
        onClose={() => setDocumentsModalVisible(false)}
        booking={selectedBooking}
      />

      {/* Модал чека об оплате */}
      <PaymentReceiptModal
        visible={receiptModalVisible}
        onClose={() => setReceiptModalVisible(false)}
        bookingId={selectedBooking?.id}
      />
    </div>
  );
};

export default BookingsPage; 