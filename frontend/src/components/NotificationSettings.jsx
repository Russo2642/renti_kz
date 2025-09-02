import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Switch,
  Select,
  Button,
  Typography,
  Space,
  Divider,
  message,
  Alert,
  Row,
  Col,
  InputNumber,
  Checkbox
} from 'antd';
import {
  BellOutlined,
  MobileOutlined,
  MailOutlined,
  SettingOutlined,
  NotificationOutlined
} from '@ant-design/icons';
import { useMutation } from '@tanstack/react-query';
import { notificationsAPI } from '../lib/api.js';
import { NOTIFICATION_TYPES, getNotificationTypeText, DEVICE_TYPES } from '../utils/notificationTypes.js';
import useAuthStore from '../store/useAuthStore.js';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;

const NotificationSettings = () => {
  const [form] = Form.useForm();
  const [isSupported, setIsSupported] = useState(false);
  const [permission, setPermission] = useState('default');
  const { user } = useAuthStore();

  // Проверяем поддержку push-уведомлений
  useEffect(() => {
    if ('serviceWorker' in navigator && 'PushManager' in window) {
      setIsSupported(true);
      // Проверяем текущее разрешение
      setPermission(Notification.permission);
    }
  }, []);

  // Мутация для регистрации устройства
  const registerDeviceMutation = useMutation({
    mutationFn: (data) => notificationsAPI.registerDevice(data),
    onSuccess: () => {
      message.success('Устройство успешно зарегистрировано для получения уведомлений');
    },
    onError: (error) => {
      console.error('Ошибка регистрации устройства:', error);
      message.error('Ошибка при регистрации устройства');
    }
  });

  // Запрос разрешения на push-уведомления
  const requestNotificationPermission = async () => {
    if (!isSupported) {
      message.error('Ваш браузер не поддерживает push-уведомления');
      return;
    }

    try {
      const permission = await Notification.requestPermission();
      setPermission(permission);

      if (permission === 'granted') {
        // Регистрируем service worker и получаем токен
        const registration = await navigator.serviceWorker.register('/sw.js');
        
        // Здесь нужно получить токен от Firebase или другого push-сервиса
        // Для примера используем заглушку
        const mockToken = `web_token_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        
        await registerDeviceMutation.mutateAsync({
          device_token: mockToken,
          device_type: DEVICE_TYPES.WEB,
          app_version: '1.0.0',
          os_version: navigator.userAgent
        });

        message.success('Push-уведомления включены!');
      } else {
        message.warning('Разрешение на уведомления не предоставлено');
      }
    } catch (error) {
      console.error('Ошибка при запросе разрешения:', error);
      message.error('Ошибка при настройке уведомлений');
    }
  };

  // Группировка типов уведомлений по категориям
  const notificationCategories = {
    bookings: {
      title: 'Бронирования',
      types: [
        NOTIFICATION_TYPES.NEW_BOOKING,
        NOTIFICATION_TYPES.BOOKING_APPROVED,
        NOTIFICATION_TYPES.BOOKING_REJECTED,
        NOTIFICATION_TYPES.BOOKING_CANCELED,
        NOTIFICATION_TYPES.BOOKING_COMPLETED,
        NOTIFICATION_TYPES.BOOKING_STARTING_SOON,
        NOTIFICATION_TYPES.BOOKING_ENDING
      ]
    },
    extensions: {
      title: 'Продления',
      types: [
        NOTIFICATION_TYPES.EXTENSION_REQUEST,
        NOTIFICATION_TYPES.EXTENSION_APPROVED,
        NOTIFICATION_TYPES.EXTENSION_REJECTED
      ]
    },
    payments: {
      title: 'Платежи',
      types: [
        NOTIFICATION_TYPES.PAYMENT_REQUIRED
      ]
    },
    apartments: {
      title: 'Квартиры',
      types: [
        NOTIFICATION_TYPES.APARTMENT_CREATED,
        NOTIFICATION_TYPES.APARTMENT_APPROVED,
        NOTIFICATION_TYPES.APARTMENT_REJECTED,
        NOTIFICATION_TYPES.APARTMENT_UPDATED,
        NOTIFICATION_TYPES.APARTMENT_STATUS_CHANGED
      ]
    },
    technical: {
      title: 'Техническое',
      types: [
        NOTIFICATION_TYPES.PASSWORD_READY,
        NOTIFICATION_TYPES.LOCK_ISSUE,
        NOTIFICATION_TYPES.CHECKOUT_REMINDER
      ]
    }
  };

  // Фильтрация категорий по роли пользователя
  const getRelevantCategories = () => {
    if (!user?.role) return notificationCategories;

    switch (user.role) {
      case 'admin':
        return notificationCategories;
      case 'owner':
        return {
          bookings: notificationCategories.bookings,
          extensions: notificationCategories.extensions,
          payments: notificationCategories.payments,
          apartments: notificationCategories.apartments
        };
      case 'renter':
        return {
          bookings: {
            ...notificationCategories.bookings,
            types: notificationCategories.bookings.types.filter(type => 
              ![NOTIFICATION_TYPES.NEW_BOOKING].includes(type)
            )
          },
          extensions: notificationCategories.extensions,
          payments: notificationCategories.payments,
          technical: notificationCategories.technical
        };
      case 'concierge':
        return {
          bookings: {
            ...notificationCategories.bookings,
            types: [
              NOTIFICATION_TYPES.NEW_BOOKING,
              NOTIFICATION_TYPES.BOOKING_STARTING_SOON,
              NOTIFICATION_TYPES.BOOKING_ENDING
            ]
          },
          technical: notificationCategories.technical
        };
      case 'cleaner':
        return {
          bookings: {
            ...notificationCategories.bookings,
            types: [
              NOTIFICATION_TYPES.BOOKING_COMPLETED,
              NOTIFICATION_TYPES.SESSION_FINISHED
            ]
          },
          technical: notificationCategories.technical
        };
      default:
        return notificationCategories;
    }
  };

  const relevantCategories = getRelevantCategories();

  const handleSaveSettings = (values) => {
    console.log('Настройки уведомлений:', values);
    message.success('Настройки сохранены');
  };

  return (
    <div className="space-y-4 md:space-y-6">
      <Card className="px-2 md:px-4">
        <div className="flex items-center space-x-2 md:space-x-3 mb-4 md:mb-6">
          <SettingOutlined className="text-xl md:text-2xl text-blue-500" />
          <div>
            <Title level={3} className="!mb-1 text-lg md:text-xl">
              Настройки уведомлений
            </Title>
            <Text type="secondary" className="text-sm md:text-base">
              Управляйте способами получения уведомлений и их типами
            </Text>
          </div>
        </div>

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveSettings}
          initialValues={{
            enablePush: permission === 'granted',
            enableEmail: true,
            quietHours: {
              enabled: false,
              start: 22,
              end: 8
            }
          }}
        >
          {/* Push-уведомления */}
          <Card size="small" className="mb-3 md:mb-4">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-2 sm:space-y-0">
              <div className="flex items-center space-x-2 md:space-x-3">
                <BellOutlined className="text-base md:text-lg text-blue-500" />
                <div>
                  <Text strong className="text-sm md:text-base">Push-уведомления</Text>
                  <br />
                  <Text type="secondary" className="text-xs md:text-sm">
                    Мгновенные уведомления в браузере
                  </Text>
                </div>
              </div>
              
              {isSupported ? (
                permission === 'granted' ? (
                  <div className="flex items-center justify-between sm:justify-end space-x-2">
                    <Text type="success" className="text-xs md:text-sm">Включено</Text>
                    <Form.Item name="enablePush" valuePropName="checked" className="!mb-0">
                      <Switch size="small" />
                    </Form.Item>
                  </div>
                ) : (
                  <Button 
                    type="primary" 
                    size="small"
                    onClick={requestNotificationPermission}
                    loading={registerDeviceMutation.isLoading}
                    className="w-full sm:w-auto"
                  >
                    Включить
                  </Button>
                )
              ) : (
                <Text type="secondary" className="text-xs md:text-sm">
                  Не поддерживается
                </Text>
              )}
            </div>
          </Card>

          {/* Email уведомления */}
          <Card size="small" className="mb-3 md:mb-4">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-2 sm:space-y-0">
              <div className="flex items-center space-x-2 md:space-x-3">
                <MailOutlined className="text-base md:text-lg text-green-500" />
                <div>
                  <Text strong className="text-sm md:text-base">Email уведомления</Text>
                  <br />
                  <Text type="secondary" className="text-xs md:text-sm">
                    Уведомления на электронную почту
                  </Text>
                </div>
              </div>
              
              <Form.Item name="enableEmail" valuePropName="checked" className="!mb-0">
                <Switch size="small" />
              </Form.Item>
            </div>
          </Card>

          <Divider />

          {/* Типы уведомлений */}
          <Title level={4} className="flex items-center text-base md:text-lg">
            <NotificationOutlined className="mr-2 text-base md:text-lg" />
            Типы уведомлений
          </Title>

          <Row gutter={[8, 8]} className="md:gutter-16">
            {Object.entries(relevantCategories).map(([categoryKey, category]) => (
              <Col xs={24} sm={24} md={12} lg={12} xl={12} key={categoryKey}>
                <Card 
                  size="small" 
                  title={
                    <span className="text-sm md:text-base">{category.title}</span>
                  } 
                  className="h-full"
                >
                  <Space direction="vertical" className="w-full" size="small">
                    {category.types.map(type => (
                      <div key={type} className="flex items-center justify-between py-1">
                        <Text className="text-xs md:text-sm flex-1 pr-2">
                          {getNotificationTypeText(type)}
                        </Text>
                        <Form.Item 
                          name={['notificationTypes', type]} 
                          valuePropName="checked" 
                          className="!mb-0 flex-shrink-0"
                          initialValue={true}
                        >
                          <Switch size="small" />
                        </Form.Item>
                      </div>
                    ))}
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>

          <Divider />

          {/* Тихие часы */}
          <Card size="small" className="px-2 md:px-4">
            <Title level={5} className="flex items-center text-sm md:text-base">
              <MobileOutlined className="mr-2 text-sm md:text-base" />
              Тихие часы
            </Title>
            <Paragraph type="secondary" className="text-xs md:text-sm">
              Не получать уведомления в определенное время
            </Paragraph>
            
            <Form.Item name={['quietHours', 'enabled']} valuePropName="checked">
              <Checkbox className="text-sm md:text-base">Включить тихие часы</Checkbox>
            </Form.Item>

            <Form.Item shouldUpdate={(prev, curr) => prev.quietHours?.enabled !== curr.quietHours?.enabled}>
              {({ getFieldValue }) => {
                const enabled = getFieldValue(['quietHours', 'enabled']);
                return enabled ? (
                  <Row gutter={[8, 8]} className="md:gutter-16">
                    <Col xs={12} sm={12} md={12}>
                      <Form.Item 
                        label={<span className="text-xs md:text-sm">Начало</span>}
                        name={['quietHours', 'start']}
                        rules={[{ required: true, message: 'Укажите время начала' }]}
                      >
                        <InputNumber 
                          min={0} 
                          max={23} 
                          formatter={value => `${value}:00`}
                          parser={value => value.replace(':00', '')}
                          className="w-full"
                          size="small"
                        />
                      </Form.Item>
                    </Col>
                    <Col xs={12} sm={12} md={12}>
                      <Form.Item 
                        label={<span className="text-xs md:text-sm">Конец</span>}
                        name={['quietHours', 'end']}
                        rules={[{ required: true, message: 'Укажите время окончания' }]}
                      >
                        <InputNumber 
                          min={0} 
                          max={23} 
                          formatter={value => `${value}:00`}
                          parser={value => value.replace(':00', '')}
                          className="w-full"
                          size="small"
                        />
                      </Form.Item>
                    </Col>
                  </Row>
                ) : null;
              }}
            </Form.Item>
          </Card>

          {/* Кнопки действий */}
          <div className="flex flex-col sm:flex-row sm:justify-end space-y-2 sm:space-y-0 sm:space-x-2 pt-3 md:pt-4">
            <Button 
              onClick={() => form.resetFields()}
              className="w-full sm:w-auto"
              size="middle"
            >
              Сбросить
            </Button>
            <Button 
              type="primary" 
              htmlType="submit"
              className="w-full sm:w-auto"
              size="middle"
            >
              Сохранить настройки
            </Button>
          </div>
        </Form>
      </Card>

      {/* Информация */}
      <Alert
        message={<span className="text-sm md:text-base">Информация о уведомлениях</span>}
        description={
          <div className="space-y-1 md:space-y-2">
            <div className="text-xs md:text-sm">• Push-уведомления работают только при открытом браузере</div>
            <div className="text-xs md:text-sm">• Email уведомления отправляются на адрес, указанный в профиле</div>
            <div className="text-xs md:text-sm">• Критические уведомления (безопасность, платежи) всегда доставляются</div>
            <div className="text-xs md:text-sm">• Тихие часы не влияют на критические уведомления</div>
          </div>
        }
        type="info"
        showIcon
        className="text-xs md:text-sm"
      />
    </div>
  );
};

export default NotificationSettings;
