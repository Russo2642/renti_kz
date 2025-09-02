import { useState, useEffect, useRef, useCallback } from 'react';
import { message } from 'antd';
import useAuthStore from '../store/useAuthStore';

const useWebSocket = (roomId) => {
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState([]);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const prevRoomIdRef = useRef(null);
  const token = useAuthStore((state) => state.accessToken);

  const connect = useCallback(() => {
    if (!roomId || !token) return;

    try {
      // WebSocket URL для подключения к чат комнате с токеном в query параметре
      // Динамически определяем протокол на основе текущего протокола страницы
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      const wsUrl = `${protocol}//${host}/api/ws/chat/${roomId}?token=${encodeURIComponent(token)}`;
      
      const ws = new WebSocket(wsUrl);
      
      ws.onopen = () => {
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          
          // Обрабатываем разные типы WebSocket сообщений
          switch (message.type) {
            case 'message':
              // Новое сообщение в чате
              setMessages(prev => [...prev, message.data]);
              break;
            case 'user_joined':
              break;
            case 'user_left':
              break;
            case 'typing':
              // Можно добавить индикатор печати
              break;
            default:
              console.log('Unknown WebSocket message type:', message.type);
          }
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      ws.onclose = (event) => {
        setIsConnected(false);
        
        // Автоматическое переподключение через 3 секунды
        if (event.code !== 1000) { // 1000 = нормальное закрытие
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, 3000);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        message.error('Ошибка подключения к чату');
      };

      wsRef.current = ws;

    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      message.error('Не удалось подключиться к чату');
    }
  }, [roomId, token]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close(1000, 'Closed by user');
      wsRef.current = null;
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const sendMessage = useCallback((messageData) => {
    if (wsRef.current && isConnected) {
      wsRef.current.send(JSON.stringify({
        type: 'message',
        data: messageData
      }));
      return true;
    }
    return false;
  }, [isConnected]);

  const sendTypingIndicator = useCallback(() => {
    if (wsRef.current && isConnected) {
      wsRef.current.send(JSON.stringify({
        type: 'typing',
        data: { timestamp: Date.now() }
      }));
    }
  }, [isConnected]);

  // Подключение при изменении roomId
  useEffect(() => {
    // Очищаем сообщения только при смене комнаты (не при переподключении)
    if (prevRoomIdRef.current !== null && prevRoomIdRef.current !== roomId) {
      setMessages([]);
    }
    prevRoomIdRef.current = roomId;

    if (roomId) {
      connect();
    }
    
    return () => {
      disconnect();
    };
  }, [roomId, connect, disconnect]);

  // Очистка при размонтировании компонента
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    messages,
    sendMessage,
    sendTypingIndicator,
    connect,
    disconnect
  };
};

export default useWebSocket;