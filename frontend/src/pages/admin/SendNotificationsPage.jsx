import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Table, Button, Card, Tag, Modal, Form, Input, Select, Space,
  Row, Col, Statistic, Typography, Badge, Tooltip, message,
  DatePicker, Popconfirm
} from 'antd';
import {
  SendOutlined, DeleteOutlined, BellOutlined, UserOutlined,
  CheckCircleOutlined, ExclamationCircleOutlined, InfoCircleOutlined,
  WarningOutlined, PlusOutlined
} from '@ant-design/icons';
import { notificationsAPI } from '../../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

const SendNotificationsPage = () => {
  const [filters, setFilters] = useState({});
  const [sendModalVisible, setSendModalVisible] = useState(false);
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  // Получение уведомлений
  const { data: notificationsData, isLoading } = useQuery({
    queryKey: ['notifications', filters],
    queryFn: () => notificationsAPI.getAll(filters)
  });

  // Получение количества непрочитанных
  const { data: unreadCount } = useQuery({
    queryKey: ['notifications-count'],
    queryFn: notificationsAPI.getCount
  });

  // Мутация для отправки уведомления
  const sendNotificationMutation = useMutation({
    mutationFn: (data) => {
      // Здесь должен быть endpoint для отправки админом уведомлений
      // Пока используем заглушку
      return Promise.resolve({ success: true });
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      setSendModalVisible(false);
      form.resetFields();
      message.success('Уведомление отправлено');
    }
  });

  // Мутация для удаления уведомления
  const deleteMutation = useMutation({
    mutationFn: notificationsAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      message.success('Уведомление удалено');
    }
  });

  // Мутация для отметки как прочитанное
  const markReadMutation = useMutation({
    mutationFn: notificationsAPI.markAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries(['notifications']);
      queryClient.invalidateQueries(['notifications-count']);
    }
  });

  const handleSendNotification = (values) => {
    sendNotificationMutation.mutate(values);
  };

  const getTypeColor = (type) => {
    const colors = {
      'info': 'blue',
      'success': 'green',
      'warning': 'orange',
      'error': 'red',
      'booking': 'purple',
      'payment': 'cyan'
    };
    return colors[type] || 'default';
  };

  const getTypeIcon = (type) => {
    const icons = {
      'info': <InfoCircleOutlined />,
      'success': <CheckCircleOutlined />,
      'warning': <WarningOutlined />,
      'error': <ExclamationCircleOutlined />,
      'booking': <BellOutlined />,
      'payment': <BellOutlined />
    };
    return icons[type] || <BellOutlined />;
  };

  const getTypeText = (type) => {
    const texts = {
      'info': 'Информация',
      'success': 'Успех',
      'warning': 'Предупреждение',
      'error': 'Ошибка',
      'booking': 'Бронирование',
      'payment': 'Платеж'
    };
    return texts[type] || type;
  };

  const columns = [
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      render: (type) => (
        <Tag color={getTypeColor(type)} icon={getTypeIcon(type)}>
          {getTypeText(type)}
        </Tag>
      ),
    },
    {
      title: 'Заголовок',
      dataIndex: 'title',
      key: 'title',
      render: (title, record) => (
        <div>
          <div className="font-medium">{title}</div>
          <div className="text-gray-500 text-sm line-clamp-2">{record.body}</div>
        </div>
      ),
    },
    {
      title: 'Получатель',
      dataIndex: ['user', 'full_name'],
      key: 'user',
      render: (name, record) => (
        <div>
          <div>{name}</div>
          <div className="text-gray-500 text-sm">{record.user?.phone}</div>
        </div>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'is_read',
      key: 'is_read',
      render: (isRead) => (
        <Tag color={isRead ? 'green' : 'orange'}>
          {isRead ? 'Прочитано' : 'Не прочитано'}
        </Tag>
      ),
    },
    {
      title: 'Отправлено',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => dayjs(date).format('DD.MM.YYYY HH:mm'),
    },
    {
      title: 'Прочитано',
      dataIndex: 'read_at',
      key: 'read_at',
      render: (date) => date ? dayjs(date).format('DD.MM.YYYY HH:mm') : '-',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space>
          {!record.is_read && (
            <Tooltip title="Отметить как прочитанное">
              <Button
                type="text"
                icon={<CheckCircleOutlined />}
                onClick={() => markReadMutation.mutate(record.id)}
              />
            </Tooltip>
          )}
          <Tooltip title="Удалить">
            <Popconfirm
              title="Удалить уведомление?"
              onConfirm={() => deleteMutation.mutate(record.id)}
              okText="Да"
              cancelText="Нет"
            >
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  // Подсчет статистики
  const stats = notificationsData?.notifications ? {
    total: notificationsData.notifications.length,
    unread: notificationsData.notifications.filter(n => !n.is_read).length,
    today: notificationsData.notifications.filter(n => 
      dayjs(n.created_at).isSame(dayjs(), 'day')
    ).length,
    week: notificationsData.notifications.filter(n => 
      dayjs(n.created_at).isAfter(dayjs().subtract(7, 'day'))
    ).length
  } : {};

  return (
    <div className="space-y-6">
      <div className="mb-6 flex flex-col lg:flex-row lg:justify-between lg:items-center space-y-4 lg:space-y-0">
        <div>
          <Title level={2}>Управление уведомлениями</Title>
          <Text type="secondary">
            Отправка и управление push-уведомлениями
          </Text>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setSendModalVisible(true)}
          className="w-full lg:w-auto"
        >
          Отправить уведомление
        </Button>
      </div>

      {/* Статистика */}
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Всего уведомлений"
              value={stats.total || 0}
              prefix={<BellOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Непрочитанные"
              value={stats.unread || 0}
              prefix={<Badge status="warning" />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Сегодня"
              value={stats.today || 0}
              prefix={<Badge status="processing" />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="За неделю"
              value={stats.week || 0}
              prefix={<Badge status="success" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Фильтры */}
      <Card className="mb-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Тип</label>
            <Select
              placeholder="Выберите тип"
              className="w-full"
              allowClear
              onChange={(value) => setFilters({ ...filters, type: value })}
            >
              <Option value="info">Информация</Option>
              <Option value="success">Успех</Option>
              <Option value="warning">Предупреждение</Option>
              <Option value="error">Ошибка</Option>
              <Option value="booking">Бронирование</Option>
              <Option value="payment">Платеж</Option>
            </Select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
            <Select
              placeholder="Статус прочтения"
              className="w-full"
              allowClear
              onChange={(value) => setFilters({ ...filters, is_read: value })}
            >
              <Option value={true}>Прочитано</Option>
              <Option value={false}>Не прочитано</Option>
            </Select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Период</label>
            <DatePicker.RangePicker
              className="w-full"
              onChange={(dates) => {
                if (dates) {
                  setFilters({
                    ...filters,
                    start_date: dates[0].format('YYYY-MM-DD'),
                    end_date: dates[1].format('YYYY-MM-DD')
                  });
                } else {
                  const { start_date, end_date, ...rest } = filters;
                  setFilters(rest);
                }
              }}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">&nbsp;</label>
            <Button 
              onClick={() => setFilters({})}
              className="w-full"
            >
              Сбросить
            </Button>
          </div>
        </div>
      </Card>

      {/* Таблица уведомлений */}
      <Card>
        <Table
          columns={columns}
          dataSource={notificationsData?.notifications || []}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 1200 }}
          pagination={{
            total: notificationsData?.total,
            pageSize: 20,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} уведомлений`,
            responsive: true,
          }}
        />
      </Card>

      {/* Модал отправки уведомления */}
      <Modal
        title="Отправить уведомление"
        open={sendModalVisible}
        onCancel={() => setSendModalVisible(false)}
        footer={null}
        width="100%"
        style={{ maxWidth: 600, top: 20 }}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSendNotification}
        >
          <Form.Item
            name="type"
            label="Тип уведомления"
            rules={[{ required: true, message: 'Выберите тип' }]}
          >
            <Select placeholder="Выберите тип уведомления">
              <Option value="info">Информация</Option>
              <Option value="success">Успех</Option>
              <Option value="warning">Предупреждение</Option>
              <Option value="error">Ошибка</Option>
              <Option value="booking">Бронирование</Option>
              <Option value="payment">Платеж</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="recipient_type"
            label="Получатели"
            rules={[{ required: true, message: 'Выберите получателей' }]}
          >
            <Select placeholder="Кому отправить">
              <Option value="all">Всем пользователям</Option>
              <Option value="owners">Только владельцам</Option>
              <Option value="concierges">Только консьержам</Option>
              <Option value="renters">Только арендаторам</Option>
              <Option value="verified">Только верифицированным</Option>
              <Option value="specific">Конкретному пользователю</Option>
            </Select>
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) =>
              prevValues.recipient_type !== currentValues.recipient_type
            }
          >
            {({ getFieldValue }) =>
              getFieldValue('recipient_type') === 'specific' ? (
                <Form.Item
                  name="user_id"
                  label="ID пользователя"
                  rules={[{ required: true, message: 'Введите ID пользователя' }]}
                >
                  <Input placeholder="Введите ID пользователя" />
                </Form.Item>
              ) : null
            }
          </Form.Item>

          <Form.Item
            name="title"
            label="Заголовок"
            rules={[{ required: true, message: 'Введите заголовок' }]}
          >
            <Input placeholder="Заголовок уведомления" />
          </Form.Item>

          <Form.Item
            name="body"
            label="Сообщение"
            rules={[{ required: true, message: 'Введите текст сообщения' }]}
          >
            <TextArea 
              rows={4} 
              placeholder="Текст уведомления..."
            />
          </Form.Item>

          <Form.Item name="data" label="Дополнительные данные (JSON)">
            <TextArea 
              rows={3} 
              placeholder='{"key": "value"}'
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button 
                type="primary" 
                htmlType="submit"
                icon={<SendOutlined />}
                loading={sendNotificationMutation.isPending}
              >
                Отправить
              </Button>
              <Button onClick={() => setSendModalVisible(false)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SendNotificationsPage; 