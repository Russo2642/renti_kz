import React, { useState, useEffect, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import {
  Row, Col, List, Card, Input, Button, Typography, Avatar, Badge,
  Tag, Space, Divider, Empty, message, Spin, Drawer
} from 'antd';
import {
  MessageOutlined, SendOutlined, UserOutlined, CalendarOutlined,
  HomeOutlined, CheckCircleOutlined, CloseCircleOutlined, WifiOutlined,
  ArrowLeftOutlined, MenuOutlined, PhoneOutlined
} from '@ant-design/icons';
import { conciergeAPI, chatAPI } from '../../lib/api.js';
import useWebSocket from '../../hooks/useWebSocket.js';
import dayjs from 'dayjs';
import './ConciergeChatPage.css';

const { Title, Text } = Typography;
const { TextArea } = Input;

const ConciergeChatPage = () => {
  const [selectedRoom, setSelectedRoom] = useState(null);
  const [newMessage, setNewMessage] = useState('');
  const [allMessages, setAllMessages] = useState([]); // Объединенные сообщения из API + WebSocket
  const [chatListVisible, setChatListVisible] = useState(false); // Для мобильного drawer
  const [searchParams, setSearchParams] = useSearchParams();
  const messagesEndRef = useRef(null);
  const queryClient = useQueryClient();

  // WebSocket подключение для выбранной комнаты
  const { 
    isConnected: wsConnected, 
    messages: wsMessages, 
    sendMessage: wsSendMessage,
    sendTypingIndicator 
  } = useWebSocket(selectedRoom?.id);

  // Получение чат-комнат консьержа
  const { data: roomsData, isLoading: roomsLoading } = useQuery({
    queryKey: ['concierge-chat-rooms'],
    queryFn: () => conciergeAPI.getChatRooms()
  });

  // Получение сообщений выбранной комнаты (только первоначальная загрузка)
  const { data: messagesData, isLoading: messagesLoading } = useQuery({
    queryKey: ['chat-messages', selectedRoom?.id],
    queryFn: () => selectedRoom ? chatAPI.getMessages(selectedRoom.id) : null,
    enabled: !!selectedRoom,
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

  // Мутация для отправки сообщения
  const sendMessageMutation = useMutation({
    mutationFn: ({ roomId, content }) => chatAPI.sendMessage(roomId, { 
      type: 'text', 
      content 
    }),
    onSuccess: () => {
      queryClient.invalidateQueries(['chat-messages', selectedRoom?.id]);
      queryClient.invalidateQueries(['concierge-chat-rooms']);
      setNewMessage('');
    },
    onError: () => {
      message.error('Ошибка при отправке сообщения');
    }
  });

  // Выбор комнаты при загрузке (если есть в URL)
  useEffect(() => {
    const roomId = searchParams.get('room');
    if (roomId && roomsData?.data?.rooms) {
      const room = roomsData.data.rooms.find(r => r.id === parseInt(roomId));
      if (room) {
        setSelectedRoom(room);
      }
    }
  }, [roomsData, searchParams]);

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

  const handleRoomSelect = (room) => {
    setSelectedRoom(room);
    setSearchParams({ room: room.id.toString() });
    setChatListVisible(false); // Закрываем мобильный drawer
    
    // Отмечаем сообщения как прочитанные
    if (room.unread_count > 0) {
      chatAPI.markAsRead(room.id).then(() => {
        queryClient.invalidateQueries(['concierge-chat-rooms']);
      });
    }
  };

  const handleSendMessage = () => {
    if (newMessage.trim() && selectedRoom) {
      sendMessageMutation.mutate({
        roomId: selectedRoom.id,
        content: newMessage.trim()
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

  const formatMessageTime = (timestamp) => {
    const messageTime = dayjs(timestamp);
    const now = dayjs();
    
    if (now.diff(messageTime, 'day') === 0) {
      return messageTime.format('HH:mm');
    } else if (now.diff(messageTime, 'day') === 1) {
      return `Вчера в ${messageTime.format('HH:mm')}`;
    } else {
      return messageTime.format('DD.MM.YYYY HH:mm');
    }
  };

  // Компонент списка чатов для повторного использования
  const ChatList = ({ className = "" }) => (
    <div className={`h-full ${className}`}>
      <div className="h-full bg-white rounded-lg border shadow-sm">
        {/* Заголовок списка чатов */}
        <div className="px-6 py-4 border-b bg-gradient-to-r from-blue-50 to-indigo-50">
          <div className="flex items-center justify-between">
            <div>
              <Title level={4} className="mb-1 text-gray-800">Чаты с гостями</Title>
              <Text type="secondary" className="text-sm">
                {roomsData?.data?.rooms?.length || 0} активных чата
              </Text>
            </div>
            <Badge 
              count={roomsData?.data?.rooms?.reduce((acc, room) => acc + (room.unread_count || 0), 0) || 0}
              showZero={false}
              className="ml-2"
            />
          </div>
        </div>

        {/* Список чатов */}
        <div className="h-[calc(100%-80px)] overflow-y-auto custom-scrollbar">
          {roomsLoading ? (
            <div className="flex justify-center items-center h-32">
              <Spin size="large" />
            </div>
          ) : roomsData?.data?.rooms?.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full px-6 py-12">
              <MessageOutlined className="text-5xl text-gray-300 mb-4" />
              <Text type="secondary" className="text-center">
                Нет активных чатов
              </Text>
            </div>
          ) : (
            <div className="space-y-1 p-2">
              {roomsData?.data?.rooms?.map((room) => (
                <div
                  key={room.id}
                  className={`relative p-4 mx-2 rounded-xl cursor-pointer chat-item-hover smooth-transition ${
                    selectedRoom?.id === room.id 
                      ? 'bg-gradient-to-r from-blue-500 to-indigo-500 text-white chat-shadow-lg chat-item-selected' 
                      : 'bg-gray-50 hover:bg-gray-100 chat-shadow'
                  }`}
                  onClick={() => handleRoomSelect(room)}
                >
                  <div className="flex items-start space-x-3">
                    {/* Аватар с уведомлениями */}
                    <div className="relative flex-shrink-0">
                      <Avatar 
                        size={48}
                        icon={<UserOutlined />}
                        className={`${
                          selectedRoom?.id === room.id 
                            ? 'bg-white text-blue-500' 
                            : 'bg-blue-500'
                        }`}
                      />
                      {room.unread_count > 0 && (
                        <Badge 
                          count={room.unread_count} 
                          size="small"
                          className="absolute -top-1 -right-1"
                        />
                      )}
                    </div>

                    {/* Информация о чате */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between mb-1">
                        <Text 
                          className={`font-semibold text-base truncate ${
                            selectedRoom?.id === room.id ? 'text-white' : 'text-gray-900'
                          }`}
                        >
                          {room.renter?.user?.first_name} {room.renter?.user?.last_name}
                        </Text>
                        <Tag 
                          color={selectedRoom?.id === room.id ? 'default' : getStatusColor(room.status)} 
                          size="small"
                          className={selectedRoom?.id === room.id ? 'bg-white/20 text-white border-white/30' : ''}
                        >
                          {getStatusText(room.status)}
                        </Tag>
                      </div>

                      <div className={`text-sm space-y-1 ${
                        selectedRoom?.id === room.id ? 'text-blue-100' : 'text-gray-600'
                      }`}>
                        <div className="flex items-center">
                          <HomeOutlined className="mr-2 text-xs" />
                          <span className="truncate">
                            {room.apartment?.description || `Квартира #${room.apartment_id}`}
                          </span>
                        </div>
                        
                        {room.booking && (
                          <div className="flex items-center">
                            <CalendarOutlined className="mr-2 text-xs" />
                            <span>
                              {dayjs(room.booking.start_date).format('DD.MM')} - {dayjs(room.booking.end_date).format('DD.MM')}
                            </span>
                          </div>
                        )}

                        {room.last_message_at && (
                          <div className={`text-xs ${
                            selectedRoom?.id === room.id ? 'text-blue-200' : 'text-gray-400'
                          }`}>
                            Последнее: {formatMessageTime(room.last_message_at)}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );

  return (
    <div className="h-[calc(100vh-140px)] p-4 bg-gray-50">
      {/* Десктопная версия */}
      <div className="hidden lg:block h-full">
        <Row gutter={24} className="h-full">
          {/* Список чатов */}
          <Col span={8} className="h-full">
            <ChatList />
          </Col>

        {/* Окно чата */}
        <Col span={16} className="h-full">
          {selectedRoom ? (
            <div className="h-full bg-white rounded-lg border shadow-sm flex flex-col">
              {/* Красивый заголовок чата */}
              <div className="px-6 py-4 border-b bg-gradient-to-r from-white to-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <div className="relative">
                      <Avatar 
                        size={56}
                        icon={<UserOutlined />}
                        className="bg-gradient-to-r from-blue-500 to-indigo-500"
                      />
                      <div className={`absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-white status-indicator ${
                        wsConnected ? 'bg-green-500' : 'bg-gray-400'
                      }`} />
                    </div>
                    <div>
                      <div className="flex items-center space-x-2">
                        <Text className="text-lg font-semibold text-gray-900">
                          {selectedRoom.renter?.user?.first_name} {selectedRoom.renter?.user?.last_name}
                        </Text>
                        <Tag color={getStatusColor(selectedRoom.status)} className="font-medium">
                          {getStatusText(selectedRoom.status)}
                        </Tag>
                      </div>
                      <div className="flex items-center space-x-4 mt-1">
                        <div className="flex items-center text-sm text-gray-600">
                          <HomeOutlined className="mr-1" />
                          {selectedRoom.apartment?.description || `Квартира #${selectedRoom.apartment_id}`}
                        </div>
                        <div className="flex items-center text-sm text-gray-600">
                          <CalendarOutlined className="mr-1" />
                          Бронь #{selectedRoom.booking_id}
                        </div>
                        {selectedRoom.renter?.user?.phone && (
                          <div className="flex items-center text-sm text-gray-600">
                            <PhoneOutlined className="mr-1" />
                            {selectedRoom.renter.user.phone}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Tag 
                      color={wsConnected ? 'success' : 'error'} 
                      icon={<WifiOutlined />}
                      className="font-medium"
                    >
                      {wsConnected ? 'Online' : 'Offline'}
                    </Tag>
                  </div>
                </div>
              </div>
              {/* Область сообщений */}
              <div className="flex-1 overflow-y-auto px-6 py-4 custom-scrollbar" style={{ 
                background: 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%)'
              }}>
                {messagesLoading ? (
                  <div className="flex justify-center items-center h-32">
                    <Spin size="large" />
                  </div>
                ) : allMessages.length === 0 ? (
                  <div className="flex flex-col items-center justify-center h-full">
                    <div className="text-center p-8">
                      <MessageOutlined className="text-6xl text-gray-300 mb-4" />
                      <Text type="secondary" className="text-lg">
                        Начните беседу с гостем
                      </Text>
                      <Text type="secondary" className="text-sm block mt-2">
                        Здесь будут отображаться ваши сообщения
                      </Text>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {allMessages.map((msg, index) => {
                      const isConsierge = msg.sender?.role === 'concierge';
                      const isSystem = msg.sender_id === 0;
                      const prevMsg = allMessages[index - 1];
                      const showAvatar = !prevMsg || prevMsg.sender_id !== msg.sender_id;

                      if (isSystem) {
                        return (
                          <div key={msg.id} className="flex justify-center my-6">
                            <div className="bg-white/70 backdrop-blur-sm px-4 py-2 rounded-full shadow-sm">
                              <Text type="secondary" className="text-xs">
                                {formatMessageTime(msg.created_at)} • {msg.content}
                              </Text>
                            </div>
                          </div>
                        );
                      }

                      return (
                        <div
                          key={msg.id}
                          className={`flex items-end space-x-2 ${
                            isConsierge ? 'justify-end' : 'justify-start'
                          }`}
                        >
                          {!isConsierge && showAvatar && (
                            <Avatar 
                              size={32}
                              icon={<UserOutlined />}
                              className="bg-gray-400 flex-shrink-0"
                            />
                          )}
                          {!isConsierge && !showAvatar && (
                            <div className="w-8" />
                          )}

                          <div className={`flex flex-col ${isConsierge ? 'items-end' : 'items-start'}`}>
                            <div
                              className={`relative max-w-xs lg:max-w-md px-4 py-3 rounded-2xl shadow-sm message-bubble smooth-transition ${
                                isConsierge
                                  ? 'bg-gradient-to-r from-blue-500 to-indigo-500 text-white'
                                  : 'bg-white text-gray-900 border border-gray-200'
                              } ${
                                isConsierge 
                                  ? 'rounded-br-md' 
                                  : 'rounded-bl-md'
                              }`}
                            >
                              <div className="text-sm leading-relaxed whitespace-pre-wrap">
                                {msg.content}
                              </div>
                              <div
                                className={`text-xs mt-2 ${
                                  isConsierge
                                    ? 'text-blue-100'
                                    : 'text-gray-500'
                                }`}
                              >
                                {formatMessageTime(msg.created_at)}
                              </div>
                            </div>
                          </div>

                          {isConsierge && showAvatar && (
                            <Avatar 
                              size={32}
                              icon={<UserOutlined />}
                              className="bg-gradient-to-r from-blue-500 to-indigo-500 flex-shrink-0"
                            />
                          )}
                          {isConsierge && !showAvatar && (
                            <div className="w-8" />
                          )}
                        </div>
                      );
                    })}
                    <div ref={messagesEndRef} />
                  </div>
                )}
              </div>

              {/* Поле ввода сообщения */}
              {selectedRoom.status === 'active' ? (
                <div className="border-t bg-white p-4">
                  <div className="flex items-end space-x-3">
                    <div className="flex-1">
                      <TextArea
                        value={newMessage}
                        onChange={handleInputChange}
                        onKeyPress={handleKeyPress}
                        placeholder="Напишите сообщение..."
                        autoSize={{ minRows: 1, maxRows: 4 }}
                        className="resize-none border-gray-300 rounded-xl focus:border-blue-500 focus:shadow-sm"
                        style={{ 
                          fontSize: '14px',
                          lineHeight: '1.5'
                        }}
                      />
                    </div>
                    <Button
                      type="primary"
                      icon={<SendOutlined />}
                      onClick={handleSendMessage}
                      loading={sendMessageMutation.isPending}
                      disabled={!newMessage.trim()}
                      size="large"
                      className="bg-gradient-to-r from-blue-500 to-indigo-500 border-0 rounded-xl shadow-md hover:shadow-lg send-button min-w-[60px] h-[40px] flex items-center justify-center"
                    />
                  </div>
                </div>
              ) : (
                <div className="border-t bg-gray-50 p-6 text-center">
                  <div className="flex items-center justify-center space-x-2 text-gray-500">
                    <CloseCircleOutlined className="text-lg" />
                    <Text type="secondary" className="text-base">
                      {selectedRoom.status === 'closed' ? 'Чат закрыт' : 'Чат неактивен'}
                    </Text>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="h-full bg-white rounded-lg border shadow-sm flex items-center justify-center">
              <div className="text-center p-8">
                <MessageOutlined className="text-6xl text-gray-300 mb-4" />
                <Text type="secondary" className="text-lg block mb-2">
                  Выберите чат для начала общения
                </Text>
                <Text type="secondary" className="text-sm">
                  Выберите чат из списка слева
                </Text>
              </div>
            </div>
          )}
        </Col>
      </Row>
      </div>

      {/* Мобильная версия */}
      <div className="lg:hidden h-full">
        {selectedRoom ? (
          <div className="h-full bg-white rounded-lg border shadow-sm flex flex-col">
            {/* Мобильный заголовок чата */}
            <div className="px-4 py-3 border-b bg-gradient-to-r from-white to-gray-50">
              <div className="flex items-center space-x-3">
                <Button 
                  type="text" 
                  icon={<ArrowLeftOutlined />}
                  onClick={() => setSelectedRoom(null)}
                  className="flex-shrink-0"
                />
                <div className="relative">
                  <Avatar 
                    size={40}
                    icon={<UserOutlined />}
                    className="bg-gradient-to-r from-blue-500 to-indigo-500"
                  />
                  <div className={`absolute -bottom-1 -right-1 w-3 h-3 rounded-full border-2 border-white ${
                    wsConnected ? 'bg-green-500' : 'bg-gray-400'
                  }`} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-2">
                    <Text className="font-semibold text-gray-900 truncate">
                      {selectedRoom.renter?.user?.first_name} {selectedRoom.renter?.user?.last_name}
                    </Text>
                    <Tag color={getStatusColor(selectedRoom.status)} size="small">
                      {getStatusText(selectedRoom.status)}
                    </Tag>
                  </div>
                  <div className="text-xs text-gray-600 truncate">
                    Бронь #{selectedRoom.booking_id}
                  </div>
                </div>
                <Tag 
                  color={wsConnected ? 'success' : 'error'} 
                  icon={<WifiOutlined />}
                  size="small"
                >
                  {wsConnected ? 'Online' : 'Offline'}
                </Tag>
              </div>
            </div>

            {/* Мобильная область сообщений */}
            <div className="flex-1 overflow-y-auto px-4 py-3" style={{ 
              background: 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%)'
            }}>
              {messagesLoading ? (
                <div className="flex justify-center items-center h-32">
                  <Spin size="large" />
                </div>
              ) : allMessages.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full">
                  <div className="text-center p-6">
                    <MessageOutlined className="text-5xl text-gray-300 mb-3" />
                    <Text type="secondary" className="text-base">
                      Начните беседу с гостем
                    </Text>
                  </div>
                </div>
              ) : (
                <div className="space-y-3">
                  {allMessages.map((msg, index) => {
                    const isConsierge = msg.sender?.role === 'concierge';
                    const isSystem = msg.sender_id === 0;
                    const prevMsg = allMessages[index - 1];
                    const showAvatar = !prevMsg || prevMsg.sender_id !== msg.sender_id;

                    if (isSystem) {
                      return (
                        <div key={msg.id} className="flex justify-center my-4">
                          <div className="bg-white/70 backdrop-blur-sm px-3 py-1 rounded-full shadow-sm">
                            <Text type="secondary" className="text-xs">
                              {formatMessageTime(msg.created_at)} • {msg.content}
                            </Text>
                          </div>
                        </div>
                      );
                    }

                    return (
                      <div
                        key={msg.id}
                        className={`flex items-end space-x-2 ${
                          isConsierge ? 'justify-end' : 'justify-start'
                        }`}
                      >
                        {!isConsierge && showAvatar && (
                          <Avatar 
                            size={24}
                            icon={<UserOutlined />}
                            className="bg-gray-400 flex-shrink-0"
                          />
                        )}
                        {!isConsierge && !showAvatar && (
                          <div className="w-6" />
                        )}

                        <div className={`flex flex-col ${isConsierge ? 'items-end' : 'items-start'}`}>
                          <div
                            className={`relative max-w-[280px] px-3 py-2 rounded-2xl shadow-sm ${
                              isConsierge
                                ? 'bg-gradient-to-r from-blue-500 to-indigo-500 text-white'
                                : 'bg-white text-gray-900 border border-gray-200'
                            } ${
                              isConsierge 
                                ? 'rounded-br-md' 
                                : 'rounded-bl-md'
                            }`}
                          >
                            <div className="text-sm leading-relaxed whitespace-pre-wrap">
                              {msg.content}
                            </div>
                            <div
                              className={`text-xs mt-1 ${
                                isConsierge
                                  ? 'text-blue-100'
                                  : 'text-gray-500'
                              }`}
                            >
                              {formatMessageTime(msg.created_at)}
                            </div>
                          </div>
                        </div>

                        {isConsierge && showAvatar && (
                          <Avatar 
                            size={24}
                            icon={<UserOutlined />}
                            className="bg-gradient-to-r from-blue-500 to-indigo-500 flex-shrink-0"
                          />
                        )}
                        {isConsierge && !showAvatar && (
                          <div className="w-6" />
                        )}
                      </div>
                    );
                  })}
                  <div ref={messagesEndRef} />
                </div>
              )}
            </div>

            {/* Мобильное поле ввода */}
            {selectedRoom.status === 'active' ? (
              <div className="border-t bg-white p-3">
                <div className="flex items-end space-x-2">
                  <div className="flex-1">
                    <TextArea
                      value={newMessage}
                      onChange={handleInputChange}
                      onKeyPress={handleKeyPress}
                      placeholder="Сообщение..."
                      autoSize={{ minRows: 1, maxRows: 3 }}
                      className="resize-none border-gray-300 rounded-xl focus:border-blue-500"
                      style={{ 
                        fontSize: '14px',
                        lineHeight: '1.4'
                      }}
                    />
                  </div>
                  <Button
                    type="primary"
                    icon={<SendOutlined />}
                    onClick={handleSendMessage}
                    loading={sendMessageMutation.isPending}
                    disabled={!newMessage.trim()}
                    className="bg-gradient-to-r from-blue-500 to-indigo-500 border-0 rounded-xl shadow-md min-w-[50px] h-[36px] flex items-center justify-center"
                  />
                </div>
              </div>
            ) : (
              <div className="border-t bg-gray-50 p-4 text-center">
                <div className="flex items-center justify-center space-x-2 text-gray-500">
                  <CloseCircleOutlined />
                  <Text type="secondary">
                    {selectedRoom.status === 'closed' ? 'Чат закрыт' : 'Чат неактивен'}
                  </Text>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="h-full bg-white rounded-lg border shadow-sm">
            {/* Мобильный заголовок списка чатов */}
            <div className="px-4 py-3 border-b bg-gradient-to-r from-blue-50 to-indigo-50">
              <div className="flex items-center justify-between">
                <div>
                  <Title level={4} className="mb-1 text-gray-800">Чаты с гостями</Title>
                  <Text type="secondary" className="text-sm">
                    {roomsData?.data?.rooms?.length || 0} активных чата
                  </Text>
                </div>
                <Badge 
                  count={roomsData?.data?.rooms?.reduce((acc, room) => acc + (room.unread_count || 0), 0) || 0}
                  showZero={false}
                />
              </div>
            </div>

            {/* Мобильный список чатов */}
            <div className="h-[calc(100%-80px)] overflow-y-auto">
              {roomsLoading ? (
                <div className="flex justify-center items-center h-32">
                  <Spin size="large" />
                </div>
              ) : roomsData?.data?.rooms?.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full px-4 py-8">
                  <MessageOutlined className="text-4xl text-gray-300 mb-3" />
                  <Text type="secondary" className="text-center">
                    Нет активных чатов
                  </Text>
                </div>
              ) : (
                <div className="space-y-1 p-2">
                  {roomsData?.data?.rooms?.map((room) => (
                    <div
                      key={room.id}
                      className="p-3 mx-2 rounded-xl cursor-pointer transition-all duration-200 bg-gray-50 hover:bg-gray-100 active:bg-gray-200"
                      onClick={() => handleRoomSelect(room)}
                    >
                      <div className="flex items-start space-x-3">
                        <div className="relative flex-shrink-0">
                          <Avatar 
                            size={40}
                            icon={<UserOutlined />}
                            className="bg-blue-500"
                          />
                          {room.unread_count > 0 && (
                            <Badge 
                              count={room.unread_count} 
                              size="small"
                              className="absolute -top-1 -right-1"
                            />
                          )}
                        </div>

                        <div className="flex-1 min-w-0">
                          <div className="flex items-center justify-between mb-1">
                            <Text className="font-semibold text-gray-900 truncate">
                              {room.renter?.user?.first_name} {room.renter?.user?.last_name}
                            </Text>
                            <Tag color={getStatusColor(room.status)} size="small">
                              {getStatusText(room.status)}
                            </Tag>
                          </div>

                          <div className="text-sm text-gray-600 space-y-1">
                            <div className="flex items-center">
                              <HomeOutlined className="mr-1 text-xs" />
                              <span className="truncate">
                                {room.apartment?.description || `Квартира #${room.apartment_id}`}
                              </span>
                            </div>
                            
                            {room.booking && (
                              <div className="flex items-center">
                                <CalendarOutlined className="mr-1 text-xs" />
                                <span>
                                  {dayjs(room.booking.start_date).format('DD.MM')} - {dayjs(room.booking.end_date).format('DD.MM')}
                                </span>
                              </div>
                            )}

                            {room.last_message_at && (
                              <div className="text-xs text-gray-400">
                                Последнее: {formatMessageTime(room.last_message_at)}
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ConciergeChatPage;