import React, { useState, useEffect, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Table, Button, Card, Tag, Modal, Form, Input, Select, Space,
  Row, Col, Statistic, Typography, Badge, Tooltip, message,
  List, Avatar, Drawer, Timeline, Divider
} from 'antd';
import {
  MessageOutlined, EyeOutlined, UserOutlined, CloseOutlined,
  CheckCircleOutlined, ExclamationCircleOutlined, SendOutlined,
  DeleteOutlined, TeamOutlined, WifiOutlined
} from '@ant-design/icons';
import { chatAPI } from '../../lib/api.js';
import useWebSocket from '../../hooks/useWebSocket.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

const ChatPage = () => {
  const [filters, setFilters] = useState({});
  const [selectedRoom, setSelectedRoom] = useState(null);
  const [chatVisible, setChatVisible] = useState(false);
  const [allMessages, setAllMessages] = useState([]); // Объединенные сообщения из API + WebSocket
  const [newMessage, setNewMessage] = useState('');
  const messagesEndRef = useRef(null);
  const queryClient = useQueryClient();

  // WebSocket подключение для выбранной комнаты
  const { 
    isConnected: wsConnected, 
    messages: wsMessages, 
    sendMessage: wsSendMessage,
    sendTypingIndicator 
  } = useWebSocket(selectedRoom?.id);

  // Получение чат-комнат
  const { data: roomsData, isLoading } = useQuery({
    queryKey: ['chat-rooms', filters],
    queryFn: () => chatAPI.getRooms(filters)
  });

  // Получение сообщений для выбранной комнаты (только первоначальная загрузка)
  const { data: messagesData, isLoading: messagesLoading } = useQuery({
    queryKey: ['chat-messages', selectedRoom?.id],
    queryFn: () => selectedRoom ? chatAPI.getMessages(selectedRoom.id) : null,
    enabled: !!selectedRoom
    // Убираем refetchInterval - новые сообщения приходят через WebSocket
  });

  // Объединяем сообщения из API и WebSocket
  useEffect(() => {
    const apiMessages = messagesData?.data || [];
    
    // Если есть API сообщения, используем их как основу
    if (apiMessages.length > 0) {
      // Фильтруем WebSocket сообщения, чтобы избежать дублирования
      const newWsMessages = wsMessages.filter(wsMsg => 
        !apiMessages.some(apiMsg => apiMsg.id === wsMsg.id)
      );
      
      // Объединяем и сортируем по времени создания
      const combined = [...apiMessages, ...newWsMessages].sort((a, b) => 
        new Date(a.created_at) - new Date(b.created_at)
      );
      
      setAllMessages(combined);
    } else if (wsMessages.length > 0) {
      // Если нет API сообщений, но есть WebSocket сообщения
      setAllMessages(wsMessages);
    }
  }, [messagesData, wsMessages, selectedRoom?.id, wsConnected]);

  // Прокрутка к последнему сообщению
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [allMessages]);

  // Сброс сообщений при смене комнаты (только если ID реально изменился)
  const prevRoomIdRef = useRef(null);
  useEffect(() => {
    if (prevRoomIdRef.current !== null && prevRoomIdRef.current !== selectedRoom?.id) {
      setAllMessages([]);
    }
    prevRoomIdRef.current = selectedRoom?.id;
  }, [selectedRoom?.id]);

  // Мутация для закрытия чата
  const closeRoomMutation = useMutation({
    mutationFn: (roomId) => chatAPI.closeChat(roomId),
    onSuccess: () => {
      queryClient.invalidateQueries(['chat-rooms']);
      message.success('Чат закрыт');
    }
  });

  // Мутация для активации чата
  const activateRoomMutation = useMutation({
    mutationFn: (roomId) => chatAPI.activateChat(roomId),
    onSuccess: () => {
      queryClient.invalidateQueries(['chat-rooms']);
      message.success('Чат активирован');
    }
  });

  // Мутация для отправки сообщения
  const sendMessageMutation = useMutation({
    mutationFn: ({ roomId, message }) => chatAPI.sendMessage(roomId, { 
      type: 'text', 
      content: message 
    }),
    onSuccess: () => {
      queryClient.invalidateQueries(['chat-messages', selectedRoom?.id]);
      setNewMessage('');
      message.success('Сообщение отправлено');
    }
  });

  // Мутация для удаления сообщения
  const deleteMessageMutation = useMutation({
    mutationFn: chatAPI.deleteMessage,
    onSuccess: () => {
      queryClient.invalidateQueries(['chat-messages', selectedRoom?.id]);
      message.success('Сообщение удалено');
    }
  });

  const getStatusColor = (status) => {
    const colors = {
      'active': 'green',
      'closed': 'red',
      'pending': 'orange'
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status) => {
    const texts = {
      'active': 'Активный',
      'closed': 'Закрыт',
      'pending': 'Ожидает'
    };
    return texts[status] || status;
  };

  const handleSendMessage = () => {
    if (newMessage.trim() && selectedRoom) {
      sendMessageMutation.mutate({
        roomId: selectedRoom.id,
        message: newMessage.trim()
      });
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  // Индикатор набора текста
  const handleInputChange = (e) => {
    setNewMessage(e.target.value);
    
    // Отправляем индикатор набора текста через WebSocket
    if (wsConnected && e.target.value.length > 0) {
      sendTypingIndicator();
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: 'Участники',
      key: 'participants',
      render: (_, record) => (
        <div>
          {record.participants?.map((participant, index) => (
            <div key={index} className="flex items-center space-x-2 mb-1">
              <Avatar size="small" icon={<UserOutlined />} src={participant.user?.avatar} />
              <span className="text-sm">
                {participant.user?.full_name}
                <Text type="secondary" className="ml-1">
                  ({participant.user?.phone})
                </Text>
              </span>
            </div>
          ))}
        </div>
      ),
    },
    {
      title: 'Тип',
      dataIndex: 'room_type',
      key: 'room_type',
      render: (type) => {
        const typeText = {
          'booking': 'Бронирование',
          'support': 'Поддержка',
          'general': 'Общий'
        };
        return <Tag color="blue">{typeText[type] || type}</Tag>;
      },
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
      title: 'Сообщений',
      dataIndex: 'messages_count',
      key: 'messages_count',
      render: (count) => (
        <Badge count={count || 0} showZero color="blue" />
      ),
    },
    {
      title: 'Непрочитанных',
      dataIndex: 'unread_count',
      key: 'unread_count',
      render: (count) => (
        <Badge count={count || 0} showZero color="red" />
      ),
    },
    {
      title: 'Последнее сообщение',
      dataIndex: 'last_message_at',
      key: 'last_message_at',
      render: (date) => date ? dayjs(date).format('DD.MM HH:mm') : '—',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="Просмотр чата">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedRoom(record);
                setChatVisible(true);
              }}
            />
          </Tooltip>
          {record.status === 'active' ? (
            <Tooltip title="Закрыть чат">
              <Button
                type="text"
                icon={<CloseOutlined />}
                onClick={() => closeRoomMutation.mutate(record.id)}
              />
            </Tooltip>
          ) : (
            <Tooltip title="Активировать чат">
              <Button
                type="text"
                icon={<CheckCircleOutlined />}
                onClick={() => activateRoomMutation.mutate(record.id)}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  // Подсчет статистики
  const stats = roomsData?.rooms ? {
    total: roomsData.rooms.length,
    active: roomsData.rooms.filter(r => r.status === 'active').length,
    closed: roomsData.rooms.filter(r => r.status === 'closed').length,
    totalMessages: roomsData.rooms.reduce((sum, r) => sum + (r.messages_count || 0), 0),
    totalUnread: roomsData.rooms.reduce((sum, r) => sum + (r.unread_count || 0), 0)
  } : {};

  return (
    <div className="p-6">
      <div className="mb-6">
        <Title level={2}>Управление чатами</Title>
        <Text type="secondary">
          Просмотр и модерация чатов пользователей
        </Text>
      </div>

      {/* Статистика */}
      <Row gutter={16} className="mb-6">
        <Col span={5}>
          <Card>
            <Statistic
              title="Всего чатов"
              value={stats.total || 0}
              prefix={<MessageOutlined />}
            />
          </Card>
        </Col>
        <Col span={5}>
          <Card>
            <Statistic
              title="Активные"
              value={stats.active || 0}
              prefix={<Badge status="success" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={5}>
          <Card>
            <Statistic
              title="Закрытые"
              value={stats.closed || 0}
              prefix={<Badge status="error" />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col span={5}>
          <Card>
            <Statistic
              title="Всего сообщений"
              value={stats.totalMessages || 0}
              prefix={<TeamOutlined />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="Непрочитанных"
              value={stats.totalUnread || 0}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Фильтры */}
      <Card className="mb-6">
        <Form layout="inline">
          <Form.Item label="Статус">
            <Select
              placeholder="Статус чата"
              style={{ width: 150 }}
              allowClear
              onChange={(value) => setFilters({ ...filters, status: value })}
            >
              <Option value="active">Активные</Option>
              <Option value="closed">Закрытые</Option>
              <Option value="pending">Ожидающие</Option>
            </Select>
          </Form.Item>
          <Form.Item label="Тип">
            <Select
              placeholder="Тип чата"
              style={{ width: 150 }}
              allowClear
              onChange={(value) => setFilters({ ...filters, room_type: value })}
            >
              <Option value="booking">Бронирование</Option>
              <Option value="support">Поддержка</Option>
              <Option value="general">Общий</Option>
            </Select>
          </Form.Item>
          <Form.Item label="Поиск">
            <Input
              placeholder="ID или участник"
              style={{ width: 200 }}
              onChange={(e) => setFilters({ ...filters, search: e.target.value })}
            />
          </Form.Item>
          <Form.Item>
            <Button 
              onClick={() => setFilters({})}
            >
              Сбросить
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {/* Таблица чатов */}
      <Card>
        <Table
          columns={columns}
          dataSource={roomsData?.rooms || []}
          loading={isLoading}
          rowKey="id"
          pagination={{
            total: roomsData?.total,
            pageSize: 15,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} из ${total} чатов`,
          }}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* Drawer с чатом */}
      <Drawer
        title={selectedRoom ? `Чат #${selectedRoom.id}` : 'Чат'}
        width={720}
        open={chatVisible}
        onClose={() => setChatVisible(false)}
        footer={
          selectedRoom && selectedRoom.status === 'active' ? (
            <div className="flex space-x-2">
              <TextArea
                value={newMessage}
                onChange={handleInputChange}
                onKeyPress={handleKeyPress}
                placeholder="Введите сообщение..."
                rows={2}
                className="flex-1"
              />
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={handleSendMessage}
                loading={sendMessageMutation.isPending}
                disabled={!newMessage.trim()}
              >
                Отправить
              </Button>
            </div>
          ) : null
        }
      >
        {selectedRoom && (
          <div>
            {/* Информация о чате */}
            <Card className="mb-4">
              <div className="space-y-2">
                <div>
                  <Text strong>Тип: </Text>
                  <Tag color="blue">
                    {selectedRoom.room_type === 'booking' ? 'Бронирование' :
                     selectedRoom.room_type === 'support' ? 'Поддержка' : 'Общий'}
                  </Tag>
                </div>
                <div>
                  <Text strong>Статус: </Text>
                  <Tag color={getStatusColor(selectedRoom.status)}>
                    {getStatusText(selectedRoom.status)}
                  </Tag>
                  <Tag 
                    color={wsConnected ? 'green' : 'red'} 
                    icon={<WifiOutlined />}
                    size="small"
                    className="ml-2"
                  >
                    {wsConnected ? 'Online' : 'Offline'}
                  </Tag>
                </div>
                <div>
                  <Text strong>Участники: </Text>
                  {selectedRoom.participants?.map((participant, index) => (
                    <span key={index}>
                      {participant.user?.full_name}
                      {index < selectedRoom.participants.length - 1 ? ', ' : ''}
                    </span>
                  ))}
                </div>
                <div>
                  <Text strong>Создан: </Text>
                  {dayjs(selectedRoom.created_at).format('DD.MM.YYYY HH:mm')}
                </div>
              </div>
            </Card>

            {/* Сообщения */}
            <div className="space-y-4 max-h-96 overflow-y-auto">
              {allMessages.map((msg) => (
                <div
                  key={msg.id}
                  className={`flex ${msg.sender?.role === 'admin' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                      msg.sender?.role === 'admin'
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-200 text-gray-900'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <Text
                        className={`text-xs ${
                          msg.sender?.role === 'admin' ? 'text-blue-100' : 'text-gray-500'
                        }`}
                      >
                        {msg.sender ? `${msg.sender.first_name} ${msg.sender.last_name}` : 'Администратор'}
                      </Text>
                      {msg.sender?.role === 'admin' && (
                        <Button
                          type="text"
                          size="small"
                          icon={<DeleteOutlined />}
                          className="text-blue-100 hover:text-white"
                          onClick={() => deleteMessageMutation.mutate(msg.id)}
                        />
                      )}
                    </div>
                    <div className="text-sm">{msg.content}</div>
                    <div
                      className={`text-xs mt-1 ${
                        msg.sender?.role === 'admin' ? 'text-blue-100' : 'text-gray-500'
                      }`}
                    >
                      {dayjs(msg.created_at).format('DD.MM HH:mm')}
                    </div>
                  </div>
                </div>
              )) || (
                <div className="text-center text-gray-500 py-8">
                  Сообщений пока нет
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default ChatPage; 