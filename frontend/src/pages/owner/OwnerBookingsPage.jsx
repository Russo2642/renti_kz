import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Table, Card, Button, Tag, Space, Tooltip, message, Modal, Form, 
  Row, Col, Statistic, Typography, Input, Drawer, Badge,
  Descriptions, Timeline, Select, DatePicker, Alert
} from 'antd';
import { 
  EyeOutlined, EditOutlined, DeleteOutlined, CheckOutlined,
  CloseOutlined, CalendarOutlined, DollarOutlined, UserOutlined,
  ClockCircleOutlined, PhoneOutlined
} from '@ant-design/icons';
import { bookingsAPI, contractsAPI } from '../../lib/api.js';
import LoadingSpinner from '../../components/LoadingSpinner.jsx';
import UserDocumentsModal from '../../components/UserDocumentsModal.jsx';
import PaymentReceiptModal from '../../components/PaymentReceiptModal.jsx';
import dayjs from 'dayjs';

const { Option } = Select;
const { Title, Text } = Typography;
const { TextArea } = Input;
const { RangePicker } = DatePicker;

const OwnerBookingsPage = () => {
  const [filters, setFilters] = useState({});
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [approveModalVisible, setApproveModalVisible] = useState(false);
  const [rejectModalVisible, setRejectModalVisible] = useState(false);
  const [documentsModalVisible, setDocumentsModalVisible] = useState(false);
  const [receiptModalVisible, setReceiptModalVisible] = useState(false);
  const [selectedBooking, setSelectedBooking] = useState(null);
  const [approveForm] = Form.useForm();
  const [rejectForm] = Form.useForm();
  const queryClient = useQueryClient();
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);

  // Отслеживание изменения размера экрана
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth < 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Получение бронирований владельца
  const { data: bookingsData, isLoading } = useQuery({
    queryKey: ['owner-bookings', filters, currentPage, pageSize],
    queryFn: () => {
      const params = {
        page: currentPage,
        page_size: pageSize,
        ...filters
      };
      return bookingsAPI.getOwnerBookings(params);
    }
  });

  // Мутация для одобрения бронирования
  const approveMutation = useMutation({
    mutationFn: ({ id, data }) => bookingsAPI.approve(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['owner-bookings']);
      setApproveModalVisible(false);
      approveForm.resetFields();
      message.success('Бронирование одобрено');
    }
  });

  // Мутация для отклонения бронирования
  const rejectMutation = useMutation({
    mutationFn: ({ id, data }) => bookingsAPI.reject(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['owner-bookings']);
      setRejectModalVisible(false);
      rejectForm.resetFields();
      message.success('Бронирование отклонено');
    }
  });

  // Мутация для отмены бронирования
  const cancelMutation = useMutation({
    mutationFn: ({ id, data }) => bookingsAPI.cancel(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['owner-bookings']);
      message.success('Бронирование отменено');
    }
  });

  // Мутация для завершения бронирования
  const finishMutation = useMutation({
    mutationFn: bookingsAPI.finish,
    onSuccess: () => {
      queryClient.invalidateQueries(['owner-bookings']);
      message.success('Бронирование завершено');
    }
  });

  const handleApprove = (values) => {
    approveMutation.mutate({
      id: selectedBooking.id,
      data: values
    });
  };

  const handleReject = (values) => {
    rejectMutation.mutate({
      id: selectedBooking.id,
      data: values
    });
  };

  // Обработчик просмотра контракта
  const handleViewContract = async (bookingId) => {
    try {
      // Сначала получаем ID договора через booking
      const contractResponse = await contractsAPI.getByBookingId(bookingId);
      const contractId = contractResponse.data.id;
      
      // Затем получаем HTML договора
      const htmlResponse = await contractsAPI.getContractHTML(contractId);
      const htmlContent = htmlResponse.data.html;
      
      const newWindow = window.open('', '_blank');
      newWindow.document.write(htmlContent);
      newWindow.document.close();
    } catch (error) {
      message.error('Ошибка при загрузке договора');
      console.error('Contract error:', error);
    }
  };

  const handleCancel = (booking) => {
    Modal.confirm({
      title: 'Отменить бронирование?',
      content: 'Вы уверены, что хотите отменить это бронирование?',
      onOk: () => {
        cancelMutation.mutate({
          id: booking.id,
          data: { reason: 'Отменено владельцем' }
        });
      }
    });
  };

  const handleFinish = (booking) => {
    Modal.confirm({
      title: 'Завершить бронирование?',
      content: 'Это действие означает, что аренда успешно завершена.',
      onOk: () => {
        finishMutation.mutate(booking.id);
      }
    });
  };

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

  const getActionButtons = (record) => {
    const buttons = [];

    // Кнопка просмотра всегда доступна
    buttons.push(
      <Tooltip key="view" title="Просмотр">
        <Button
          type="text"
          icon={<EyeOutlined />}
          onClick={() => {
            setSelectedBooking(record);
            setDetailsVisible(true);
          }}
        />
      </Tooltip>
    );

    // Действия в зависимости от статуса
    switch (record.status) {
      case 'pending':
        buttons.push(
          <Tooltip key="approve" title="Одобрить">
            <Button
              type="text"
              icon={<CheckOutlined />}
              style={{ color: '#52c41a' }}
              onClick={() => {
                setSelectedBooking(record);
                setApproveModalVisible(true);
              }}
            />
          </Tooltip>,
          <Tooltip key="reject" title="Отклонить">
            <Button
              type="text"
              icon={<CloseOutlined />}
              danger
              onClick={() => {
                setSelectedBooking(record);
                setRejectModalVisible(true);
              }}
            />
          </Tooltip>
        );
        break;
      
      case 'approved':
      case 'active':
        buttons.push(
          <Tooltip key="cancel" title="Отменить">
            <Button
              type="text"
              icon={<CloseOutlined />}
              danger
              onClick={() => handleCancel(record)}
            />
          </Tooltip>
        );
        
        if (record.status === 'active') {
          buttons.push(
            <Tooltip key="finish" title="Завершить">
              <Button
                type="text"
                icon={<CheckOutlined />}
                style={{ color: '#52c41a' }}
                onClick={() => handleFinish(record)}
              />
            </Tooltip>
          );
        }
        break;
    }

    return buttons;
  };

  const columns = [
    {
      title: 'Номер',
      dataIndex: 'booking_number',
      key: 'booking_number',
      width: isMobile ? 80 : 120,
      render: (number) => (
        <Text strong className={isMobile ? 'text-xs' : ''}>#{number}</Text>
      ),
    },
    {
      title: 'Квартира',
      key: 'apartment',
      render: (_, record) => (
        <div>
          <div className={`font-medium ${isMobile ? 'text-sm' : ''}`}>
            {record.apartment?.description ? 
              (record.apartment.description.length > 40 ? 
                record.apartment.description.substring(0, 40) + '...' : 
                record.apartment.description
              ) :
              `${record.apartment?.street}, ${record.apartment?.building}-${record.apartment?.apartment_number}`
            }
          </div>
          <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-sm'}`}>
            {record.apartment?.room_count} комн., {record.apartment?.total_area} м²
          </div>
        </div>
      ),
    },
    ...(isMobile ? [] : [{
      title: 'Арендатор',
      key: 'renter',
      render: (_, record) => (
        <div>
          <div className="flex items-center space-x-2">
            <UserOutlined />
            <span>
              {record.renter?.user?.first_name} {record.renter?.user?.last_name}
            </span>
          </div>
          <div className="text-gray-500 text-sm flex items-center space-x-1">
            <PhoneOutlined />
            <span>{record.renter?.user?.phone}</span>
          </div>
        </div>
      ),
    }]),
    {
      title: 'Период',
      key: 'period',
      render: (_, record) => (
        <div>
          <div className={isMobile ? 'text-xs' : 'text-sm'}>
            <CalendarOutlined className="mr-1" />
            {dayjs(record.start_date).format(isMobile ? 'DD.MM' : 'DD.MM.YYYY HH:mm')}
          </div>
          <div className={isMobile ? 'text-xs' : 'text-sm'}>
            {dayjs(record.end_date).format(isMobile ? 'DD.MM' : 'DD.MM.YYYY HH:mm')}
          </div>
          <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-xs'}`}>
            {calculateDuration(record.start_date, record.end_date)}
          </div>
        </div>
      ),
    },
    {
      title: 'Стоимость',
      key: 'final_price',
      render: (_, record) => (
        <div>
          <div className="flex items-center space-x-1">
            <DollarOutlined />
            <Text strong className={isMobile ? 'text-xs' : ''}>{record.final_price?.toLocaleString()} ₸</Text>
          </div>
          <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-xs'}`}>
            Базовая: {record.total_price?.toLocaleString()} ₸
          </div>
          {record.service_fee > 0 && (
            <div className={`text-gray-500 ${isMobile ? 'text-xs' : 'text-xs'}`}>
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
        <Tag color={getStatusColor(status)} className={isMobile ? 'text-xs' : ''}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    ...(isMobile ? [] : [{
      title: 'Создано',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD.MM HH:mm'),
    }]),
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space size={isMobile ? 'small' : 'middle'}>
          {getActionButtons(record).map((button, index) => 
            React.cloneElement(button, { key: index, size: isMobile ? 'small' : 'default' })
          )}
        </Space>
      ),
    },
  ];

  const bookings = bookingsData?.data?.bookings || [];
  const stats = bookings.length > 0 ? {
    total: bookings.length,
    pending: bookings.filter(b => b.status === 'pending').length,
    approved: bookings.filter(b => b.status === 'approved').length,
    active: bookings.filter(b => b.status === 'active').length,
    completed: bookings.filter(b => b.status === 'completed').length,
    totalRevenue: bookings
      .filter(b => ['approved', 'completed', 'active'].includes(b.status))
      .reduce((sum, b) => sum + (b.final_price || 0), 0)
  } : {};

  return (
    <div className={`${isMobile ? 'p-4' : 'p-6'}`}>
      <div className="mb-6">
        <Title level={2} className={isMobile ? 'text-xl' : ''}>Мои бронирования</Title>
        <Text type="secondary" className={isMobile ? 'text-sm' : ''}>
          Управление бронированиями ваших квартир
        </Text>
      </div>

      {/* Статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={8} md={4}>
          <Card>
            <Statistic
              title="Всего бронирований"
              value={stats.total || 0}
              prefix={<CalendarOutlined />}
              valueStyle={{ fontSize: isMobile ? '18px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card>
            <Statistic
              title="Ожидают одобрения"
              value={stats.pending || 0}
              prefix={<Badge status="warning" />}
              valueStyle={{ color: '#faad14', fontSize: isMobile ? '18px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card>
            <Statistic
              title="Одобрено"
              value={stats.approved || 0}
              prefix={<Badge status="success" />}
              valueStyle={{ color: '#52c41a', fontSize: isMobile ? '18px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card>
            <Statistic
              title="Активные"
              value={stats.active || 0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#13c2c2', fontSize: isMobile ? '18px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card>
            <Statistic
              title="Завершено"
              value={stats.completed || 0}
              prefix={<CheckOutlined />}
              valueStyle={{ color: '#52c41a', fontSize: isMobile ? '18px' : '24px' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card>
            <Statistic
              title="Общий доход"
              value={stats.totalRevenue || 0}
              prefix={<DollarOutlined />}
              suffix="₸"
              formatter={(value) => value.toLocaleString()}
              valueStyle={{ color: '#52c41a', fontSize: isMobile ? '18px' : '24px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Уведомления о действиях */}
      {stats.pending > 0 && (
        <Alert
          message={`У вас ${stats.pending} бронирований ожидают одобрения`}
          type="warning"
          showIcon
          className="mb-6"
          action={
            <Button size="small" type="link">
              Просмотреть
            </Button>
          }
        />
      )}

      {/* Фильтры */}
      <Card className="mb-6">
        <div className={`${isMobile ? 'space-y-4' : 'flex flex-wrap gap-4 items-end'}`}>
          <div className="flex-1 min-w-[150px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
            <Select
              placeholder="Выберите статус"
              style={{ width: '100%' }}
              allowClear
              onChange={(value) => setFilters({ ...filters, status: value })}
            >
              <Option value="created">Создано</Option>
              <Option value="pending">На рассмотрении</Option>
              <Option value="approved">Одобрено</Option>
              <Option value="rejected">Отклонено</Option>
              <Option value="canceled">Отменено</Option>
              <Option value="completed">Завершено</Option>
              <Option value="active">Активно</Option>
            </Select>
          </div>
          <div className="flex-1 min-w-[200px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Квартира</label>
            <Select
              placeholder="Выберите квартиру"
              style={{ width: '100%' }}
              allowClear
              onChange={(value) => setFilters({ ...filters, apartment_id: value })}
            >
              {/* Здесь должен быть список квартир владельца */}
            </Select>
          </div>
          <div className="flex-1 min-w-[200px]">
            <label className="block text-sm font-medium text-gray-700 mb-1">Период</label>
            <RangePicker
              style={{ width: '100%' }}
              onChange={(dates) => {
                if (dates && dates.length === 2 && dates[0] && dates[1]) {
                  setFilters({
                    ...filters,
                    date_from: dates[0].format('YYYY-MM-DD'),
                    date_to: dates[1].format('YYYY-MM-DD')
                  });
                } else {
                  const newFilters = { ...filters };
                  delete newFilters.date_from;
                  delete newFilters.date_to;
                  setFilters(newFilters);
                }
              }}
            />
          </div>
          <div className="flex-shrink-0">
            <Button onClick={() => setFilters({})}>
              Сбросить
            </Button>
          </div>
        </div>
      </Card>

      {/* Таблица бронирований */}
      <Card>
        <Table
          columns={columns}
          dataSource={bookingsData?.data?.bookings || []}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: isMobile ? 600 : 1200 }}
          size={isMobile ? 'small' : 'default'}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: bookingsData?.data?.pagination?.total || 0,
            showSizeChanger: !isMobile,
            showQuickJumper: !isMobile,
            showTotal: (total, range) => 
              range ? `${range[0]}-${range[1]} из ${total} бронирований` : `${total} бронирований`,
            responsive: true,
            simple: isMobile,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            },
          }}
        />
      </Card>

      {/* Модал одобрения */}
      <Modal
        title="Одобрить бронирование"
        open={approveModalVisible}
        onCancel={() => setApproveModalVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 600}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={approveForm}
          layout="vertical"
          onFinish={handleApprove}
        >
          <Form.Item
            name="comment"
            label="Комментарий"
            rules={[{ required: true, message: 'Добавьте комментарий' }]}
          >
            <TextArea 
              rows={4} 
              placeholder="Добавьте комментарий для арендатора..."
            />
          </Form.Item>
          <Form.Item>
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={approveMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Одобрить бронирование
              </Button>
              <Button 
                onClick={() => setApproveModalVisible(false)}
                className={isMobile ? 'w-full' : ''}
              >
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Модал отклонения */}
      <Modal
        title="Отклонить бронирование"
        open={rejectModalVisible}
        onCancel={() => setRejectModalVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 600}
        style={isMobile ? { top: 20 } : {}}
      >
        <Form
          form={rejectForm}
          layout="vertical"
          onFinish={handleReject}
        >
          <Form.Item
            name="reason"
            label="Причина отклонения"
            rules={[{ required: true, message: 'Укажите причину отклонения' }]}
          >
            <TextArea 
              rows={4} 
              placeholder="Укажите причину отклонения бронирования..."
            />
          </Form.Item>
          <Form.Item>
            <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
              <Button 
                type="primary" 
                danger
                htmlType="submit"
                loading={rejectMutation.isPending}
                className={isMobile ? 'w-full' : ''}
              >
                Отклонить бронирование
              </Button>
              <Button 
                onClick={() => setRejectModalVisible(false)}
                className={isMobile ? 'w-full' : ''}
              >
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Drawer деталей */}
      <Drawer
        title="Детали бронирования"
        placement={isMobile ? 'bottom' : 'right'}
        size={isMobile ? 'default' : 'large'}
        onClose={() => setDetailsVisible(false)}
        open={detailsVisible}
        width={isMobile ? '100%' : 720}
        height={isMobile ? '90%' : undefined}
      >
        {selectedBooking && (
          <div>
            <Descriptions
              column={isMobile ? 1 : 2}
              bordered
              size={isMobile ? 'small' : 'default'}
            >
              <Descriptions.Item label="Номер бронирования">
                <Text strong>#{selectedBooking.booking_number}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                <Tag color={getStatusColor(selectedBooking.status)}>
                  {getStatusText(selectedBooking.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Квартира" span={isMobile ? 1 : 2}>
                <div className="break-words">
                  {selectedBooking.apartment?.description ? 
                    selectedBooking.apartment.description :
                    `${selectedBooking.apartment?.street}, д. ${selectedBooking.apartment?.building}, кв. ${selectedBooking.apartment?.apartment_number}`
                  }
                  <br />
                  <Text type="secondary">
                    {selectedBooking.apartment?.room_count} комн., {selectedBooking.apartment?.total_area} м²
                  </Text>
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Арендатор">
                {selectedBooking.renter?.user?.first_name} {selectedBooking.renter?.user?.last_name}
              </Descriptions.Item>
              <Descriptions.Item label="Телефон">
                {selectedBooking.renter?.user?.phone}
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
              <Descriptions.Item label="Длительность" span={isMobile ? 1 : 2}>
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
              <Descriptions.Item label="Создано" span={isMobile ? 1 : 2}>
                <div className="font-mono text-sm">
                  {dayjs(selectedBooking.created_at).format('DD.MM.YYYY')}
                  <br />
                  {dayjs(selectedBooking.created_at).format('HH:mm')}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="Обновлено" span={isMobile ? 1 : 2}>
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
                      <div>Площадь: {selectedBooking.apartment?.total_area} м²</div>
                      <div>Этаж: {selectedBooking.apartment?.floor}/{selectedBooking.apartment?.total_floors}</div>
                      <div>Состояние: {selectedBooking.apartment?.condition?.name}</div>
                    </div>
                  </Card>
                </Col>
                <Col xs={24} sm={24} md={12}>
                  <Card size="small" title="Контактная информация">
                    <div className="space-y-2">
                      <div>
                        <UserOutlined className="mr-2" />
                        {selectedBooking.renter?.user?.first_name} {selectedBooking.renter?.user?.last_name}
                      </div>
                      <div>
                        <PhoneOutlined className="mr-2" />
                        {selectedBooking.renter?.user?.phone}
                      </div>
                      <div>
                        <span className="mr-2">📧</span>
                        {selectedBooking.renter?.user?.email}
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

            {/* История изменений */}
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

            {/* Быстрые действия */}
            {(selectedBooking.status === 'pending' || 
              selectedBooking.status === 'approved' || 
              selectedBooking.status === 'active') && (
              <div className="mt-6">
                <Title level={4}>Действия</Title>
                <Space direction={isMobile ? 'vertical' : 'horizontal'} className={isMobile ? 'w-full' : ''}>
                  {selectedBooking.status === 'pending' && (
                    <>
                      <Button
                        type="primary"
                        icon={<CheckOutlined />}
                        onClick={() => {
                          setApproveModalVisible(true);
                          setDetailsVisible(false);
                        }}
                        className={isMobile ? 'w-full' : ''}
                      >
                        Одобрить
                      </Button>
                      <Button
                        danger
                        icon={<CloseOutlined />}
                        onClick={() => {
                          setRejectModalVisible(true);
                          setDetailsVisible(false);
                        }}
                        className={isMobile ? 'w-full' : ''}
                      >
                        Отклонить
                      </Button>
                    </>
                  )}
                  {(selectedBooking.status === 'approved' || selectedBooking.status === 'active') && (
                    <Button
                      danger
                      icon={<CloseOutlined />}
                      onClick={() => handleCancel(selectedBooking)}
                      className={isMobile ? 'w-full' : ''}
                    >
                      Отменить
                    </Button>
                  )}
                  {selectedBooking.status === 'active' && (
                    <Button
                      type="primary"
                      icon={<CheckOutlined />}
                      onClick={() => handleFinish(selectedBooking)}
                      className={isMobile ? 'w-full' : ''}
                    >
                      Завершить
                    </Button>
                  )}
                </Space>
              </div>
            )}
          </div>
        )}
      </Drawer>

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

export default OwnerBookingsPage; 