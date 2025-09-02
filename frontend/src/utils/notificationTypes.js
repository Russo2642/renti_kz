// Типы уведомлений из backend
export const NOTIFICATION_TYPES = {
  // Бронирования
  BOOKING_APPROVED: 'booking_approved',
  BOOKING_REJECTED: 'booking_rejected',
  BOOKING_CANCELED: 'booking_canceled',
  BOOKING_COMPLETED: 'booking_completed',
  BOOKING_STARTING_SOON: 'booking_starting_soon',
  BOOKING_ENDING: 'booking_ending',
  NEW_BOOKING: 'new_booking',
  SESSION_FINISHED: 'session_finished',
  
  // Пароли и замки
  PASSWORD_READY: 'password_ready',
  LOCK_ISSUE: 'lock_issue',
  
  // Продления
  EXTENSION_REQUEST: 'extension_request',
  EXTENSION_APPROVED: 'extension_approved',
  EXTENSION_REJECTED: 'extension_rejected',
  
  // Напоминания
  CHECKOUT_REMINDER: 'checkout_reminder',
  
  // Платежи
  PAYMENT_REQUIRED: 'payment_required',
  
  // Квартиры
  APARTMENT_CREATED: 'apartment_created',
  APARTMENT_APPROVED: 'apartment_approved',
  APARTMENT_REJECTED: 'apartment_rejected',
  APARTMENT_UPDATED: 'apartment_updated',
  APARTMENT_STATUS_CHANGED: 'apartment_status_changed'
};

// Приоритеты уведомлений
export const NOTIFICATION_PRIORITIES = {
  LOW: 'low',
  NORMAL: 'normal',
  HIGH: 'high',
  URGENT: 'urgent'
};

// Типы устройств
export const DEVICE_TYPES = {
  IOS: 'ios',
  ANDROID: 'android',
  WEB: 'web'
};

// Цвета для типов уведомлений
export const getNotificationTypeColor = (type) => {
  const colors = {
    [NOTIFICATION_TYPES.BOOKING_APPROVED]: 'green',
    [NOTIFICATION_TYPES.BOOKING_REJECTED]: 'red',
    [NOTIFICATION_TYPES.BOOKING_CANCELED]: 'orange',
    [NOTIFICATION_TYPES.BOOKING_COMPLETED]: 'blue',
    [NOTIFICATION_TYPES.BOOKING_STARTING_SOON]: 'orange',
    [NOTIFICATION_TYPES.BOOKING_ENDING]: 'orange',
    [NOTIFICATION_TYPES.NEW_BOOKING]: 'blue',
    [NOTIFICATION_TYPES.SESSION_FINISHED]: 'gray',
    [NOTIFICATION_TYPES.PASSWORD_READY]: 'cyan',
    [NOTIFICATION_TYPES.LOCK_ISSUE]: 'red',
    [NOTIFICATION_TYPES.EXTENSION_REQUEST]: 'purple',
    [NOTIFICATION_TYPES.EXTENSION_APPROVED]: 'green',
    [NOTIFICATION_TYPES.EXTENSION_REJECTED]: 'red',
    [NOTIFICATION_TYPES.CHECKOUT_REMINDER]: 'orange',
    [NOTIFICATION_TYPES.PAYMENT_REQUIRED]: 'red',
    [NOTIFICATION_TYPES.APARTMENT_CREATED]: 'green',
    [NOTIFICATION_TYPES.APARTMENT_APPROVED]: 'green',
    [NOTIFICATION_TYPES.APARTMENT_REJECTED]: 'red',
    [NOTIFICATION_TYPES.APARTMENT_UPDATED]: 'blue',
    [NOTIFICATION_TYPES.APARTMENT_STATUS_CHANGED]: 'orange'
  };
  return colors[type] || 'default';
};

// Текст для типов уведомлений
export const getNotificationTypeText = (type) => {
  const texts = {
    [NOTIFICATION_TYPES.BOOKING_APPROVED]: 'Бронь подтверждена',
    [NOTIFICATION_TYPES.BOOKING_REJECTED]: 'Бронь отклонена',
    [NOTIFICATION_TYPES.BOOKING_CANCELED]: 'Бронь отменена',
    [NOTIFICATION_TYPES.BOOKING_COMPLETED]: 'Бронь завершена',
    [NOTIFICATION_TYPES.BOOKING_STARTING_SOON]: 'Бронь скоро начнется',
    [NOTIFICATION_TYPES.BOOKING_ENDING]: 'Бронь заканчивается',
    [NOTIFICATION_TYPES.NEW_BOOKING]: 'Новая бронь',
    [NOTIFICATION_TYPES.SESSION_FINISHED]: 'Сессия завершена',
    [NOTIFICATION_TYPES.PASSWORD_READY]: 'Пароль готов',
    [NOTIFICATION_TYPES.LOCK_ISSUE]: 'Проблема с замком',
    [NOTIFICATION_TYPES.EXTENSION_REQUEST]: 'Запрос продления',
    [NOTIFICATION_TYPES.EXTENSION_APPROVED]: 'Продление одобрено',
    [NOTIFICATION_TYPES.EXTENSION_REJECTED]: 'Продление отклонено',
    [NOTIFICATION_TYPES.CHECKOUT_REMINDER]: 'Напоминание о выселении',
    [NOTIFICATION_TYPES.PAYMENT_REQUIRED]: 'Требуется оплата',
    [NOTIFICATION_TYPES.APARTMENT_CREATED]: 'Квартира создана',
    [NOTIFICATION_TYPES.APARTMENT_APPROVED]: 'Квартира одобрена',
    [NOTIFICATION_TYPES.APARTMENT_REJECTED]: 'Квартира отклонена',
    [NOTIFICATION_TYPES.APARTMENT_UPDATED]: 'Квартира обновлена',
    [NOTIFICATION_TYPES.APARTMENT_STATUS_CHANGED]: 'Статус квартиры изменен'
  };
  return texts[type] || type;
};

// Цвета для приоритетов
export const getPriorityColor = (priority) => {
  const colors = {
    [NOTIFICATION_PRIORITIES.LOW]: '#52c41a',
    [NOTIFICATION_PRIORITIES.NORMAL]: '#1890ff',
    [NOTIFICATION_PRIORITIES.HIGH]: '#fa8c16',
    [NOTIFICATION_PRIORITIES.URGENT]: '#f5222d'
  };
  return colors[priority] || '#1890ff';
};

// Текст для приоритетов
export const getPriorityText = (priority) => {
  const texts = {
    [NOTIFICATION_PRIORITIES.LOW]: 'Низкий',
    [NOTIFICATION_PRIORITIES.NORMAL]: 'Обычный',
    [NOTIFICATION_PRIORITIES.HIGH]: 'Высокий',
    [NOTIFICATION_PRIORITIES.URGENT]: 'Критический'
  };
  return texts[priority] || priority;
};

// Иконки для типов уведомлений
export const getNotificationTypeIcon = (type) => {
  // Возвращаем название иконки из ant design
  const icons = {
    [NOTIFICATION_TYPES.BOOKING_APPROVED]: 'CheckCircleOutlined',
    [NOTIFICATION_TYPES.BOOKING_REJECTED]: 'CloseCircleOutlined',
    [NOTIFICATION_TYPES.BOOKING_CANCELED]: 'StopOutlined',
    [NOTIFICATION_TYPES.BOOKING_COMPLETED]: 'CheckOutlined',
    [NOTIFICATION_TYPES.BOOKING_STARTING_SOON]: 'ClockCircleOutlined',
    [NOTIFICATION_TYPES.BOOKING_ENDING]: 'ExclamationCircleOutlined',
    [NOTIFICATION_TYPES.NEW_BOOKING]: 'PlusCircleOutlined',
    [NOTIFICATION_TYPES.SESSION_FINISHED]: 'CheckCircleOutlined',
    [NOTIFICATION_TYPES.PASSWORD_READY]: 'KeyOutlined',
    [NOTIFICATION_TYPES.LOCK_ISSUE]: 'LockOutlined',
    [NOTIFICATION_TYPES.EXTENSION_REQUEST]: 'QuestionCircleOutlined',
    [NOTIFICATION_TYPES.EXTENSION_APPROVED]: 'CheckCircleOutlined',
    [NOTIFICATION_TYPES.EXTENSION_REJECTED]: 'CloseCircleOutlined',
    [NOTIFICATION_TYPES.CHECKOUT_REMINDER]: 'ClockCircleOutlined',
    [NOTIFICATION_TYPES.PAYMENT_REQUIRED]: 'DollarCircleOutlined',
    [NOTIFICATION_TYPES.APARTMENT_CREATED]: 'HomeOutlined',
    [NOTIFICATION_TYPES.APARTMENT_APPROVED]: 'CheckCircleOutlined',
    [NOTIFICATION_TYPES.APARTMENT_REJECTED]: 'CloseCircleOutlined',
    [NOTIFICATION_TYPES.APARTMENT_UPDATED]: 'EditOutlined',
    [NOTIFICATION_TYPES.APARTMENT_STATUS_CHANGED]: 'InfoCircleOutlined'
  };
  return icons[type] || 'BellOutlined';
};

// Фильтрация уведомлений по роли пользователя
export const filterNotificationsByRole = (notifications, userRole) => {
  if (!userRole || !notifications) return notifications;

  // Определяем какие типы уведомлений релевантны для каждой роли
  const roleRelevantTypes = {
    admin: Object.values(NOTIFICATION_TYPES), // Админ видит все
    owner: [
      NOTIFICATION_TYPES.NEW_BOOKING,
      NOTIFICATION_TYPES.BOOKING_APPROVED,
      NOTIFICATION_TYPES.BOOKING_REJECTED,
      NOTIFICATION_TYPES.BOOKING_CANCELED,
      NOTIFICATION_TYPES.BOOKING_COMPLETED,
      NOTIFICATION_TYPES.SESSION_FINISHED,
      NOTIFICATION_TYPES.EXTENSION_REQUEST,
      NOTIFICATION_TYPES.PAYMENT_REQUIRED,
      NOTIFICATION_TYPES.APARTMENT_APPROVED,
      NOTIFICATION_TYPES.APARTMENT_REJECTED,
      NOTIFICATION_TYPES.APARTMENT_STATUS_CHANGED
    ],
    renter: [
      NOTIFICATION_TYPES.BOOKING_APPROVED,
      NOTIFICATION_TYPES.BOOKING_REJECTED,
      NOTIFICATION_TYPES.BOOKING_CANCELED,
      NOTIFICATION_TYPES.BOOKING_COMPLETED,
      NOTIFICATION_TYPES.BOOKING_STARTING_SOON,
      NOTIFICATION_TYPES.BOOKING_ENDING,
      NOTIFICATION_TYPES.PASSWORD_READY,
      NOTIFICATION_TYPES.LOCK_ISSUE,
      NOTIFICATION_TYPES.EXTENSION_APPROVED,
      NOTIFICATION_TYPES.EXTENSION_REJECTED,
      NOTIFICATION_TYPES.CHECKOUT_REMINDER,
      NOTIFICATION_TYPES.PAYMENT_REQUIRED
    ],
    concierge: [
      NOTIFICATION_TYPES.NEW_BOOKING,
      NOTIFICATION_TYPES.BOOKING_STARTING_SOON,
      NOTIFICATION_TYPES.BOOKING_ENDING,
      NOTIFICATION_TYPES.LOCK_ISSUE,
      NOTIFICATION_TYPES.CHECKOUT_REMINDER
    ],
    cleaner: [
      NOTIFICATION_TYPES.BOOKING_COMPLETED,
      NOTIFICATION_TYPES.SESSION_FINISHED,
      NOTIFICATION_TYPES.CHECKOUT_REMINDER
    ]
  };

  const relevantTypes = roleRelevantTypes[userRole] || [];
  
  return notifications.filter(notification => 
    relevantTypes.includes(notification.type)
  );
};

// Группировка уведомлений по дате
export const groupNotificationsByDate = (notifications) => {
  if (!notifications || notifications.length === 0) return {};

  const groups = {};
  
  notifications.forEach(notification => {
    const date = new Date(notification.created_at).toDateString();
    if (!groups[date]) {
      groups[date] = [];
    }
    groups[date].push(notification);
  });

  return groups;
};

// Сортировка уведомлений по приоритету и дате
export const sortNotificationsByPriority = (notifications) => {
  if (!notifications || notifications.length === 0) return [];

  const priorityOrder = {
    [NOTIFICATION_PRIORITIES.URGENT]: 4,
    [NOTIFICATION_PRIORITIES.HIGH]: 3,
    [NOTIFICATION_PRIORITIES.NORMAL]: 2,
    [NOTIFICATION_PRIORITIES.LOW]: 1
  };

  return [...notifications].sort((a, b) => {
    // Сначала по приоритету
    const priorityDiff = (priorityOrder[b.priority] || 2) - (priorityOrder[a.priority] || 2);
    if (priorityDiff !== 0) return priorityDiff;

    // Потом по дате (новые сначала)
    return new Date(b.created_at) - new Date(a.created_at);
  });
};

// Проверка, нужно ли показывать уведомление пользователю с определенной ролью
export const shouldShowNotificationForRole = (notification, userRole) => {
  const filtered = filterNotificationsByRole([notification], userRole);
  return filtered.length > 0;
};

export default {
  NOTIFICATION_TYPES,
  NOTIFICATION_PRIORITIES,
  DEVICE_TYPES,
  getNotificationTypeColor,
  getNotificationTypeText,
  getPriorityColor,
  getPriorityText,
  getNotificationTypeIcon,
  filterNotificationsByRole,
  groupNotificationsByDate,
  sortNotificationsByPriority,
  shouldShowNotificationForRole
};
