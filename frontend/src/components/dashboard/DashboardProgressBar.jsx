import React, { memo, useMemo } from 'react';
import { Card, Progress, Row, Col, Typography } from 'antd';
import { ThunderboltOutlined, HomeOutlined } from '@ant-design/icons';

const { Text, Title } = Typography;

const DashboardProgressBar = memo(({
  title,
  data = {},
  loading = false,
  type = 'battery', // battery, occupancy, custom
  showPercentage = true,
  size = 'default', // small, default, large
  layout = 'vertical', // vertical, horizontal
  className = ''
}) => {
  // Преобразуем данные в структурированный формат
  const processedData = useMemo(() => {
    if (!data || typeof data !== 'object') return [];

    return Object.entries(data).map(([key, value]) => {
      let percentage = 0;
      let status = 'normal';
      let color = '#1890ff';
      let label = key;

      if (type === 'battery') {
        // Для батареи: low, medium, high, critical, unknown
        switch (key.toLowerCase()) {
          case 'critical':
            percentage = 15;
            status = 'exception';
            color = '#ff4d4f';
            label = 'Критический';
            break;
          case 'low':
            percentage = 30;
            status = 'exception';
            color = '#faad14';
            label = 'Низкий';
            break;
          case 'medium':
            percentage = 60;
            status = 'normal';
            color = '#1890ff';
            label = 'Средний';
            break;
          case 'high':
            percentage = 85;
            status = 'success';
            color = '#52c41a';
            label = 'Высокий';
            break;
          case 'unknown':
            percentage = 0;
            status = 'normal';
            color = '#8c8c8c';
            label = 'Неизвестно';
            break;
          default:
            percentage = 50;
            status = 'normal';
            color = '#1890ff';
            label = key;
        }
      } else if (type === 'occupancy') {
        // Для заполняемости - value является процентом
        percentage = Math.min(Math.max(value, 0), 100);
        if (percentage >= 80) {
          status = 'success';
          color = '#52c41a';
        } else if (percentage >= 50) {
          status = 'normal';
          color = '#1890ff';
        } else if (percentage >= 20) {
          status = 'normal';
          color = '#faad14';
        } else {
          status = 'exception';
          color = '#ff4d4f';
        }
        label = key;
      } else {
        // Кастомный тип - value является процентом
        percentage = Math.min(Math.max(value, 0), 100);
        if (percentage >= 70) {
          status = 'success';
          color = '#52c41a';
        } else if (percentage >= 40) {
          status = 'normal';
          color = '#1890ff';
        } else {
          status = 'exception';
          color = '#ff4d4f';
        }
        label = key;
      }

      return {
        key,
        label,
        value,
        percentage,
        status,
        color
      };
    });
  }, [data, type]);

  const getIcon = (type) => {
    switch (type) {
      case 'battery':
        return <ThunderboltOutlined />;
      case 'occupancy':
        return <HomeOutlined />;
      default:
        return null;
    }
  };

  const getProgressSize = () => {
    switch (size) {
      case 'small':
        return 6;
      case 'large':
        return 10;
      default:
        return 8;
    }
  };

  if (loading) {
    return (
      <Card title={title} className={className}>
        <div className="flex items-center justify-center p-8">
          <div className="animate-pulse text-gray-400">Загрузка данных...</div>
        </div>
      </Card>
    );
  }

  if (!processedData || processedData.length === 0) {
    return (
      <Card title={title} className={className}>
        <div className="flex items-center justify-center p-8">
          <div className="text-gray-400">Нет данных для отображения</div>
        </div>
      </Card>
    );
  }

  return (
    <Card title={title} className={`dashboard-progress-bar ${className}`}>
      <div className="space-y-4">
        {processedData.map((item) => (
          <div key={item.key} className="progress-item">
            {layout === 'horizontal' ? (
              <Row align="middle" gutter={[16, 8]}>
                <Col flex="120px">
                  <Text strong className="flex items-center">
                    {getIcon(type) && <span className="mr-2">{getIcon(type)}</span>}
                    {item.label}
                  </Text>
                </Col>
                <Col flex="auto">
                  <Progress
                    percent={item.percentage}
                    status={item.status}
                    strokeColor={item.color}
                    size={getProgressSize()}
                    showInfo={showPercentage}
                    format={(percent) => 
                      type === 'battery' 
                        ? `${item.value}` 
                        : `${percent}%`
                    }
                  />
                </Col>
                {type !== 'battery' && (
                  <Col flex="60px">
                    <Text type="secondary">
                      {item.value} {type === 'occupancy' ? 'шт.' : ''}
                    </Text>
                  </Col>
                )}
              </Row>
            ) : (
              <div className="text-center">
                <div className="mb-2">
                  <Text strong className="flex items-center justify-center">
                    {getIcon(type) && <span className="mr-2">{getIcon(type)}</span>}
                    {item.label}
                  </Text>
                  {type !== 'battery' && (
                    <Text type="secondary" className="text-sm">
                      {item.value} {type === 'occupancy' ? 'объектов' : 'единиц'}
                    </Text>
                  )}
                </div>
                <Progress
                  type="circle"
                  percent={item.percentage}
                  status={item.status}
                  strokeColor={item.color}
                  size={[
                    size === 'small' ? 60 : size === 'large' ? 100 : 80,
                    getProgressSize()
                  ]}
                  format={(percent) => 
                    type === 'battery' 
                      ? `${item.value}` 
                      : `${percent}%`
                  }
                />
              </div>
            )}
          </div>
        ))}
      </div>
    </Card>
  );
});

DashboardProgressBar.displayName = 'DashboardProgressBar';

export default DashboardProgressBar; 