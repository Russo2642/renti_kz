// Service Worker для push-уведомлений Renti.kz

const CACHE_NAME = 'renti-notifications-v1';

// Установка service worker
self.addEventListener('install', (event) => {
  console.log('Service Worker установлен');
  self.skipWaiting();
});

// Активация service worker
self.addEventListener('activate', (event) => {
  console.log('Service Worker активирован');
  event.waitUntil(self.clients.claim());
});

// Обработка push-уведомлений
self.addEventListener('push', (event) => {
  console.log('Получено push-уведомление:', event);

  let notificationData = {
    title: 'Renti.kz',
    body: 'У вас новое уведомление',
    icon: '/favicon.ico',
    badge: '/favicon.ico',
    tag: 'renti-notification',
    requireInteraction: false,
    actions: [
      {
        action: 'open',
        title: 'Открыть',
        icon: '/favicon.ico'
      },
      {
        action: 'close',
        title: 'Закрыть'
      }
    ],
    data: {
      url: '/'
    }
  };

  // Парсим данные из push-сообщения
  if (event.data) {
    try {
      const pushData = event.data.json();
      
      notificationData = {
        ...notificationData,
        title: pushData.title || notificationData.title,
        body: pushData.body || notificationData.body,
        icon: pushData.icon || notificationData.icon,
        tag: pushData.tag || `renti-${Date.now()}`,
        requireInteraction: pushData.priority === 'urgent',
        data: {
          url: pushData.url || '/',
          notificationId: pushData.notification_id,
          bookingId: pushData.booking_id,
          apartmentId: pushData.apartment_id,
          type: pushData.type
        }
      };

      // Настройка действий в зависимости от типа уведомления
      if (pushData.type) {
        switch (pushData.type) {
          case 'booking_approved':
          case 'booking_rejected':
          case 'new_booking':
            notificationData.data.url = `/bookings/${pushData.booking_id || ''}`;
            break;
          case 'apartment_approved':
          case 'apartment_rejected':
          case 'apartment_status_changed':
            notificationData.data.url = `/apartments/${pushData.apartment_id || ''}`;
            break;
          case 'payment_required':
            notificationData.data.url = `/payments`;
            break;
          case 'password_ready':
            notificationData.data.url = `/locks`;
            break;
          default:
            notificationData.data.url = '/notifications';
        }
      }

      // Цвет и приоритет уведомления
      if (pushData.priority === 'urgent') {
        notificationData.requireInteraction = true;
        notificationData.silent = false;
      } else if (pushData.priority === 'low') {
        notificationData.silent = true;
      }

    } catch (error) {
      console.error('Ошибка парсинга push-данных:', error);
    }
  }

  // Показываем уведомление
  event.waitUntil(
    self.registration.showNotification(notificationData.title, notificationData)
  );
});

// Обработка клика по уведомлению
self.addEventListener('notificationclick', (event) => {
  console.log('Клик по уведомлению:', event);

  const notification = event.notification;
  const action = event.action;
  const data = notification.data || {};

  notification.close();

  if (action === 'close') {
    return;
  }

  // Открываем соответствующую страницу
  const urlToOpen = data.url || '/';
  
  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true })
      .then((clients) => {
        // Ищем уже открытое окно с нашим сайтом
        const existingClient = clients.find(client => 
          client.url.includes(self.location.origin)
        );

        if (existingClient) {
          // Фокусируемся на существующем окне и переходим на нужную страницу
          existingClient.focus();
          return existingClient.navigate(urlToOpen);
        } else {
          // Открываем новое окно
          return self.clients.openWindow(urlToOpen);
        }
      })
  );

  // Отправляем информацию о клике обратно в приложение
  if (data.notificationId) {
    event.waitUntil(
      self.clients.matchAll().then((clients) => {
        clients.forEach(client => {
          client.postMessage({
            type: 'NOTIFICATION_CLICKED',
            notificationId: data.notificationId,
            action: action || 'open'
          });
        });
      })
    );
  }
});

// Обработка закрытия уведомления
self.addEventListener('notificationclose', (event) => {
  console.log('Уведомление закрыто:', event);
  
  const data = event.notification.data || {};
  
  // Отправляем информацию о закрытии
  if (data.notificationId) {
    event.waitUntil(
      self.clients.matchAll().then((clients) => {
        clients.forEach(client => {
          client.postMessage({
            type: 'NOTIFICATION_CLOSED',
            notificationId: data.notificationId
          });
        });
      })
    );
  }
});

// Обработка сообщений от главного потока
self.addEventListener('message', (event) => {
  console.log('Сообщение в Service Worker:', event.data);

  if (event.data && event.data.type) {
    switch (event.data.type) {
      case 'SKIP_WAITING':
        self.skipWaiting();
        break;
      case 'GET_VERSION':
        event.ports[0].postMessage({ version: CACHE_NAME });
        break;
    }
  }
});

// Фоновая синхронизация (если поддерживается)
if ('sync' in self.registration) {
  self.addEventListener('sync', (event) => {
    console.log('Фоновая синхронизация:', event.tag);
    
    if (event.tag === 'background-sync-notifications') {
      event.waitUntil(
        // Здесь можно синхронизировать непрочитанные уведомления
        fetch('/api/notifications/unread')
          .then(response => response.json())
          .then(data => {
            console.log('Синхронизированы уведомления:', data);
          })
          .catch(error => {
            console.error('Ошибка синхронизации:', error);
          })
      );
    }
  });
}

console.log('Service Worker загружен для Renti.kz');
