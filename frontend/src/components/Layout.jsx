import React, { useState, useEffect } from 'react';
import { Layout as AntLayout, Menu, Avatar, Dropdown, Badge, Button, Typography, Drawer } from 'antd';
import { useLocation, useNavigate } from 'react-router-dom';
import {
  DashboardOutlined,
  UserOutlined,
  HomeOutlined,
  CalendarOutlined,
  BellOutlined,
  LockOutlined,
  CustomerServiceOutlined,
  MessageOutlined,
  LogoutOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BarChartOutlined,
  ToolOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import useAuthStore from '../store/useAuthStore.js';
import { notificationsAPI } from '../lib/api.js';
import NotificationCenter from './NotificationCenter.jsx';

const { Header, Sider, Content } = AntLayout;
const { Title } = Typography;

const Layout = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();

  // Проверка размера экрана
  useEffect(() => {
    const checkScreenSize = () => {
      const mobile = window.innerWidth < 992;
      setIsMobile(mobile);
      if (mobile) {
        setCollapsed(true); // На мобильных меню закрыто по умолчанию
      }
    };

    checkScreenSize();
    window.addEventListener('resize', checkScreenSize);
    return () => window.removeEventListener('resize', checkScreenSize);
  }, []);

  // Убираем отдельный запрос уведомлений, используем NotificationCenter

  // Меню для админов и модераторов
  const adminMenuItems = [
    {
      key: '/admin/dashboard',
      icon: <DashboardOutlined />,
      label: 'Дашборд',
    },
    {
      key: '/admin/users',
      icon: <UserOutlined />,
      label: 'Пользователи',
    },
    {
      key: '/admin/apartments',
      icon: <HomeOutlined />,
      label: 'Квартиры',
    },
    {
      key: '/admin/bookings',
      icon: <CalendarOutlined />,
      label: 'Бронирования',
    },
    {
      key: '/admin/locks',
      icon: <LockOutlined />,
      label: 'Замки',
    },
    {
      key: '/admin/concierges',
      icon: <CustomerServiceOutlined />,
      label: 'Консьержи',
    },
    {
      key: '/admin/cleaners',
      icon: <ToolOutlined />,
      label: 'Уборщицы',
    },
    {
      key: '/admin/notifications',
      icon: <BellOutlined />,
      label: 'Уведомления',
    },
  ];

  // Меню для владельцев
  const ownerMenuItems = [
    // {
    //   key: '/owner/dashboard',
    //   icon: <DashboardOutlined />,
    //   label: 'Дашборд',
    // },
    {
      key: '/owner/apartments',
      icon: <HomeOutlined />,
      label: 'Мои квартиры',
    },
    {
      key: '/owner/bookings',
      icon: <CalendarOutlined />,
      label: 'Бронирования',
    },
    {
      key: '/owner/statistics',
      icon: <BarChartOutlined />,
      label: 'Статистика',
    },
  ];

  // Меню для консьержей
  const conciergeMenuItems = [
    {
      key: '/concierge/dashboard',
      icon: <DashboardOutlined />,
      label: 'Дашборд',
    },
    {
      key: '/concierge/apartments',
      icon: <HomeOutlined />,
      label: 'Мои квартиры',
    },
    {
      key: '/concierge/bookings',
      icon: <CalendarOutlined />,
      label: 'Бронирования',
    },
    {
      key: '/concierge/chat',
      icon: <MessageOutlined />,
      label: 'Чат с гостями',
    },
  ];

  // Меню для уборщиц
  const cleanerMenuItems = [
    {
      key: '/cleaner/dashboard',
      icon: <DashboardOutlined />,
      label: 'Дашборд',
    },
    {
      key: '/cleaner/apartments',
      icon: <HomeOutlined />,
      label: 'Мои квартиры',
    },
    {
      key: '/cleaner/apartments-for-cleaning',
      icon: <ToolOutlined />,
      label: 'Уборка квартир',
    },
    {
      key: '/cleaner/schedule',
      icon: <CalendarOutlined />,
      label: 'Мое расписание',
    },
  ];

  const getMenuItems = () => {
    switch (user?.role) {
      case 'owner':
        return ownerMenuItems;
      case 'concierge':
        return conciergeMenuItems;
      case 'cleaner':
        return cleanerMenuItems;
      default:
        return adminMenuItems;
    }
  };

  const menuItems = getMenuItems();

  // Выход из системы
  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  // Меню профиля
  const profileMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Профиль',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: 'Настройки',
      onClick: () => navigate('/settings'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Выйти',
      onClick: handleLogout,
    },
  ];

  const getRoleTitle = (role) => {
    switch (role) {
      case 'admin':
        return 'Администратор';
      case 'moderator':
        return 'Модератор';
      case 'owner':
        return 'Владелец';
      case 'concierge':
        return 'Консьерж';
      case 'cleaner':
        return 'Уборщица';
      default:
        return 'Пользователь';
    }
  };

  const getPageTitle = () => {
    const path = location.pathname;
    
    if (path.includes('/dashboard')) return 'Дашборд';
    if (path.includes('/users')) return 'Пользователи';
    if (path.includes('/apartments')) return 'Квартиры';
    if (path.includes('/bookings')) return 'Бронирования';
    if (path.includes('/locks')) return 'Замки';
    if (path.includes('/notifications')) return 'Уведомления';
    if (path.includes('/statistics')) return 'Статистика';
    if (path.includes('/chat')) return 'Чат с гостями';
    if (path.includes('/concierges')) return 'Консьержи';
    if (path.includes('/cleaners')) return 'Уборщицы';
    
    if (user?.role === 'concierge') return 'Консьерж панель';
    if (user?.role === 'cleaner') return 'Панель уборщицы';
    if (user?.role === 'owner') return 'Панель владельца';
    return 'Админ панель';
  };

  return (
    <AntLayout className="min-h-screen w-full">
      {/* Боковая панель для десктопа */}
      {!isMobile && (
        <Sider
          trigger={null}
          collapsible
          collapsed={collapsed}
          className="!bg-white shadow-lg"
          width={260}
          collapsedWidth={80}
          style={{ minHeight: '100vh' }}
        >
          {/* Логотип */}
          <div className="flex items-center justify-center py-4 md:py-6 px-4 border-b border-gray-100">
            {!collapsed ? (
              <div className="text-center">
                <Title level={3} className="!mb-1 !text-blue-600">
                  Renti.kz
                </Title>
                <p className="text-xs text-gray-500 uppercase tracking-wider">
                  {getRoleTitle(user?.role)}
                </p>
              </div>
            ) : (
              <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold">R</span>
              </div>
            )}
          </div>

          {/* Навигационное меню */}
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            className="border-none"
            items={menuItems}
            onClick={({ key }) => navigate(key)}
          />
        </Sider>
      )}

      {/* Мобильное меню как Drawer */}
      <Drawer
        title={
          <div className="text-center">
            <Title level={3} className="!mb-1 !text-blue-600">
              Renti.kz
            </Title>
            <p className="text-xs text-gray-500 uppercase tracking-wider">
              {getRoleTitle(user?.role)}
            </p>
          </div>
        }
        placement="left"
        onClose={() => setCollapsed(true)}
        open={!collapsed && isMobile}
        styles={{ body: { padding: 0 } }}
        width={260}
        className="lg:hidden"
      >
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          className="border-none"
          items={menuItems}
          onClick={({ key }) => {
            navigate(key);
            setCollapsed(true); // Закрываем drawer после клика
          }}
        />
      </Drawer>

      {/* Основная область */}
      <AntLayout className="w-full" style={{ minHeight: '100vh' }}>
        {/* Шапка */}
        <Header className="flex items-center justify-between !px-4 md:!px-6 !bg-white border-b border-gray-100" style={{ position: 'sticky', top: 0, zIndex: 1000 }}>
          <div className="flex items-center space-x-2 md:space-x-4">
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              className="!w-10 !h-10"
            />
            
            <div className="hidden sm:block">
              <Title level={3} className="!mb-0">
                {getPageTitle()}
              </Title>
              <p className="text-sm text-gray-500 !mb-0">
                Добро пожаловать, {user?.first_name} {user?.last_name}
              </p>
            </div>
            
            <div className="block sm:hidden">
              <Title level={4} className="!mb-0">
                {getPageTitle()}
              </Title>
            </div>
          </div>

          <div className="flex items-center space-x-2 md:space-x-4">
            {/* Уведомления */}
            <NotificationCenter 
              maxItems={5}
              showAllLink={true}
              placement="bottomRight"
            />

            {/* Профиль */}
            <Dropdown
              menu={{ items: profileMenuItems }}
              placement="bottomRight"
              arrow
            >
              <div className="flex items-center space-x-2 cursor-pointer hover:bg-gray-50 rounded-lg px-2 md:px-3 py-2">
                <Avatar 
                  size="default" 
                  icon={<UserOutlined />}
                  className="bg-blue-500"
                >
                  {user?.first_name?.[0]}{user?.last_name?.[0]}
                </Avatar>
                <div className="text-left hidden md:block">
                  <div className="text-sm font-medium text-gray-900">
                    {user?.first_name} {user?.last_name}
                  </div>
                  <div className="text-xs text-gray-500">
                    {getRoleTitle(user?.role)}
                  </div>
                </div>
              </div>
            </Dropdown>
          </div>
        </Header>

        {/* Контент */}
        <Content className="flex-1 w-full" style={{ padding: '12px 16px', minHeight: 'calc(100vh - 64px)', background: '#f5f5f5' }}>
          <div className="fade-in w-full h-full max-w-full">
            {children}
          </div>
        </Content>
      </AntLayout>
    </AntLayout>
  );
};

export default Layout; 