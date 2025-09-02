import React, { memo } from 'react';
import { Card, Statistic, Tooltip } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';

const KPICard = memo(({ 
  title, 
  value, 
  suffix = '', 
  prefix, 
  growthRate, 
  loading = false,
  formatter,
  tooltipTitle,
  className = '',
  size = 'default', // default, small, large
  onClick = null, // функция для обработки клика
  clickable = false // указывает, что карточка кликабельна
}) => {
  const formatGrowthRate = (rate) => {
    if (rate === null || rate === undefined || isNaN(rate)) return null;
    return `${rate > 0 ? '+' : ''}${rate.toFixed(1)}%`;
  };

  const getGrowthColor = (rate) => {
    if (rate > 0) return '#52c41a'; // green
    if (rate < 0) return '#ff4d4f'; // red
    return '#8c8c8c'; // gray
  };

  const getGrowthIcon = (rate) => {
    if (rate > 0) return <ArrowUpOutlined />;
    if (rate < 0) return <ArrowDownOutlined />;
    return null;
  };

  const cardContent = (
    <Card
      loading={loading}
      className={`dashboard-kpi-card ${className} ${clickable ? 'cursor-pointer hover:shadow-lg transition-shadow' : ''}`}
      styles={size === 'small' ? { body: { padding: '16px' } } : undefined}
      onClick={clickable && onClick ? onClick : undefined}
    >
      <Statistic
        title={title}
        value={value}
        suffix={suffix}
        prefix={prefix}
        formatter={formatter}
        valueStyle={{ 
          fontSize: size === 'large' ? '32px' : size === 'small' ? '20px' : '24px',
          fontWeight: 600
        }}
      />
      
      {growthRate !== null && growthRate !== undefined && !isNaN(growthRate) && (
        <div 
          className="mt-2 text-sm"
          style={{ color: getGrowthColor(growthRate) }}
        >
          {getGrowthIcon(growthRate)}
          <span className="ml-1">
            {formatGrowthRate(growthRate)} за период
          </span>
        </div>
      )}
    </Card>
  );

  return tooltipTitle ? (
    <Tooltip title={tooltipTitle}>
      {cardContent}
    </Tooltip>
  ) : cardContent;
});

KPICard.displayName = 'KPICard';

export default KPICard; 