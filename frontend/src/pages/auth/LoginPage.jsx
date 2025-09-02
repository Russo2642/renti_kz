import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, App } from 'antd';
import { UserOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import useAuthStore from '../../store/useAuthStore.js';

const { Title, Text, Link } = Typography;

const LoginPage = () => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { login } = useAuthStore();
  const { message } = App.useApp();

  // Функция для форматирования номера телефона
  const formatPhoneNumber = (value) => {
    // Убираем все нецифровые символы
    const phoneNumber = value.replace(/\D/g, '');
    
    // Если номер начинается с 8, заменяем на 7
    const normalizedNumber = phoneNumber.startsWith('8') ? '7' + phoneNumber.slice(1) : phoneNumber;
    
    // Форматируем в +7 (777) 731 82 42
    if (normalizedNumber.length === 0) return '';
    if (normalizedNumber.length <= 1) return `+${normalizedNumber}`;
    if (normalizedNumber.length <= 4) return `+${normalizedNumber.slice(0, 1)} (${normalizedNumber.slice(1)}`;
    if (normalizedNumber.length <= 7) return `+${normalizedNumber.slice(0, 1)} (${normalizedNumber.slice(1, 4)}) ${normalizedNumber.slice(4)}`;
    if (normalizedNumber.length <= 9) return `+${normalizedNumber.slice(0, 1)} (${normalizedNumber.slice(1, 4)}) ${normalizedNumber.slice(4, 7)} ${normalizedNumber.slice(7)}`;
    return `+${normalizedNumber.slice(0, 1)} (${normalizedNumber.slice(1, 4)}) ${normalizedNumber.slice(4, 7)} ${normalizedNumber.slice(7, 9)} ${normalizedNumber.slice(9, 11)}`;
  };

  // Функция для извлечения чистого номера телефона
  const getCleanPhoneNumber = (value) => {
    return value.replace(/\D/g, '');
  };

  // Обработчик изменения номера телефона
  const handlePhoneChange = (e) => {
    const value = e.target.value;
    const formatted = formatPhoneNumber(value);
    form.setFieldsValue({ phone: formatted });
  };

  const handleLogin = async (values) => {
    setLoading(true);
    try {
      // Получаем чистый номер телефона без форматирования
      const cleanPhone = getCleanPhoneNumber(values.phone);
      const result = await login(cleanPhone, values.password);
      
      if (result.success) {
        message.success('Добро пожаловать!');
        // Редиректим в зависимости от роли пользователя
        const redirectPath = result.user?.role === 'owner' ? '/owner/apartments' : '/admin/dashboard';
        navigate(redirectPath);
      } else {
        message.error(result.error || 'Ошибка авторизации');
      }
    } catch (error) {
      message.error('Ошибка авторизации');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-blue-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Логотип и заголовок */}
        <div className="text-center mb-8">
          <div className="mx-auto w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center mb-4 shadow-lg">
            <span className="text-white text-2xl font-bold">R</span>
          </div>
          <Title level={2} className="!mb-2 !text-gray-800">
            Renti.kz
          </Title>
          <Text className="text-gray-600">
            Административная панель
          </Text>
        </div>

        {/* Форма входа */}
        <Card 
          className="shadow-xl border-0 rounded-2xl"
          styles={{ body: { padding: '2rem' } }}
        >
          <div className="text-center mb-6">
            <Title level={3} className="!mb-2">
              Вход в систему
            </Title>
            <Text className="text-gray-500">
              Введите свои учетные данные для входа
            </Text>
          </div>

          <Form
            form={form}
            name="login"
            onFinish={handleLogin}
            layout="vertical"
            size="large"
            autoComplete="off"
          >
            <Form.Item
              name="phone"
              label="Номер телефона"
              rules={[
                { required: true, message: 'Введите номер телефона' },
                { 
                  validator: (_, value) => {
                    const cleanNumber = getCleanPhoneNumber(value || '');
                    if (cleanNumber.length === 11 && cleanNumber.startsWith('7')) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('Введите корректный номер телефона'));
                  }
                }
              ]}
            >
              <Input
                prefix={<PhoneOutlined className="text-gray-400" />}
                placeholder="+7 (777) 777 77 77"
                maxLength={18}
                className="rounded-lg h-12"
                onChange={handlePhoneChange}
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="Пароль"
              rules={[
                { required: true, message: 'Введите пароль' },
                { min: 6, message: 'Пароль должен содержать минимум 6 символов' }
              ]}
            >
              <Input.Password
                prefix={<LockOutlined className="text-gray-400" />}
                placeholder="Введите пароль"
                className="rounded-lg h-12"
              />
            </Form.Item>

            <Form.Item className="!mb-6">
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                className="w-full h-12 rounded-lg text-base font-medium shadow-lg"
              >
                {loading ? 'Вход...' : 'Войти'}
              </Button>
            </Form.Item>
          </Form>
        </Card>

        {/* Футер */}
        <div className="text-center mt-8">
          <Text className="text-xs text-gray-400">
            © 2025 Renti.kz. Все права защищены.
          </Text>
        </div>
      </div>
    </div>
  );
};

export default LoginPage; 