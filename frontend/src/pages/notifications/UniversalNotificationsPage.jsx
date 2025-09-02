import React, { useState, useEffect } from 'react';
import { Typography, Tabs, Space, Button } from 'antd';
import { BellOutlined, SettingOutlined, ReloadOutlined } from '@ant-design/icons';
import NotificationList from '../../components/NotificationList.jsx';
import NotificationSettings from '../../components/NotificationSettings.jsx';
import useNotifications from '../../hooks/useNotifications.js';
import './NotificationsPage.css';

const { Title, Text } = Typography;

const UniversalNotificationsPage = () => {
  const [activeTab, setActiveTab] = useState('notifications');
  const [isDesktop, setIsDesktop] = useState(false);
  
  const { refresh, isLoading, count } = useNotifications({
    autoRefresh: true,
    refreshInterval: 30000,
    enableRealtime: true,
    onNewNotification: (notification) => {
      console.log('Новое уведомление:', notification);
    }
  });

  useEffect(() => {
    const checkScreenSize = () => {
      setIsDesktop(window.innerWidth > 992);
    };
    
    checkScreenSize();
    window.addEventListener('resize', checkScreenSize);
    
    return () => window.removeEventListener('resize', checkScreenSize);
  }, []);

  const tabItems = [
    {
      key: 'notifications',
      label: (
        <span className="flex items-center">
          <BellOutlined className="mr-2" />
          Мои уведомления
          {count > 0 && (
            <span className="notification-badge">
              {count}
            </span>
          )}
        </span>
      ),
      children: (
        <div className="notification-tab-content space-y-4 md:space-y-6">
          {/* Заголовок и действия */}
          <div className="flex flex-col md:flex-row md:justify-between md:items-center space-y-3 md:space-y-0 px-1">
            <div>
              <Text type="secondary" className="text-sm md:text-base">
                Просматривайте и управляйте своими уведомлениями
              </Text>
            </div>
            
            <Space wrap size="small">
              <Button 
                icon={<ReloadOutlined />} 
                onClick={refresh}
                loading={isLoading}
                size="middle"
                className="w-full md:w-auto"
              >
                Обновить
              </Button>
            </Space>
          </div>

          {/* Список уведомлений */}
          <div className="px-1">
            <NotificationList
              title={`Все уведомления${count > 0 ? ` (${count} непрочитанных)` : ''}`}
              showPagination={true}
              pageSize={20}
              showBulkActions={true}
              showStats={true}
              cardProps={{
                className: "shadow-sm"
              }}
            />
          </div>
        </div>
      )
    },
    {
      key: 'settings',
      label: (
        <span className="flex items-center">
          <SettingOutlined className="mr-2" />
          Настройки уведомлений
        </span>
      ),
      children: (
        <div className="notification-tab-content px-1">
          <NotificationSettings />
        </div>
      )
    }
  ];

  return (
    <div className="space-y-4 md:space-y-6 px-2 md:px-4 lg:px-6">
      {/* Заголовок */}
      <div className="notifications-header">
        <div className="flex flex-col space-y-2 md:space-y-4">
          <div>
            <Title level={2} className="!mb-2 flex items-center">
              <BellOutlined className="mr-2 md:mr-3 text-blue-500" />
              <span className="text-lg md:text-xl lg:text-2xl">Уведомления</span>
            </Title>
            <Text type="secondary" className="text-sm md:text-base">
              Управление вашими уведомлениями и настройками
            </Text>
          </div>
        </div>
      </div>

      {/* Табы */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        size="large"
        className="notification-tabs"
        tabBarStyle={{ 
          marginBottom: 0,
          paddingLeft: 0,
          paddingRight: 0,
          ...(isDesktop && {
            width: '100%'
          })
        }}
        tabBarExtraContent={null}
        style={{
          ...(isDesktop && {
            '--ant-tabs-tab-padding': '16px 32px',
            '--ant-tabs-tab-min-width': '200px'
          })
        }}
      />
    </div>
  );
};

export default UniversalNotificationsPage;
