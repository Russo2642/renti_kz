import React, { memo, useMemo } from 'react';
import { Card } from 'antd';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';

const DashboardAreaChart = memo(({
  title,
  data = [],
  dataKeys = [],
  height = 300,
  loading = false,
  colors = ['#1890ff', '#52c41a', '#faad14', '#f5222d'],
  stacked = false,
  formatTooltip,
  formatXAxisLabel,
  formatYAxisLabel,
  showGrid = true,
  showLegend = true,
  fillOpacity = 0.6,
  className = ''
}) => {
  // Оптимизируем данные для графика
  const chartData = useMemo(() => {
    if (!data || data.length === 0) return [];
    return data.map(item => ({
      ...item,
      // Форматируем дату для отображения на оси X
      displayDate: formatXAxisLabel ? formatXAxisLabel(item.date) : item.date
    }));
  }, [data, formatXAxisLabel]);

  // Создаем конфигурацию областей
  const areas = useMemo(() => {
    return dataKeys.map((key, index) => ({
      dataKey: key.dataKey || key,
      name: key.name || key,
      color: key.color || colors[index % colors.length],
      fillOpacity: key.fillOpacity || fillOpacity,
      strokeWidth: key.strokeWidth || 2
    }));
  }, [dataKeys, colors, fillOpacity]);

  const defaultTooltipFormatter = (value, name) => [
    `${typeof value === 'number' ? value.toLocaleString() : value}`,
    name
  ];

  const tooltipFormatter = formatTooltip || defaultTooltipFormatter;

  // Создаем градиенты для каждой области
  const gradients = useMemo(() => {
    return areas.map((area, index) => {
      const gradientId = `gradient-${index}`;
      return (
        <defs key={gradientId}>
          <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={area.color} stopOpacity={0.8} />
            <stop offset="95%" stopColor={area.color} stopOpacity={0.1} />
          </linearGradient>
        </defs>
      );
    });
  }, [areas]);

  if (loading) {
    return (
      <Card title={title} className={className}>
        <div className="flex items-center justify-center" style={{ height }}>
          <div className="animate-pulse text-gray-400">Загрузка данных...</div>
        </div>
      </Card>
    );
  }

  if (!chartData || chartData.length === 0) {
    return (
      <Card title={title} className={className}>
        <div className="flex items-center justify-center" style={{ height }}>
          <div className="text-gray-400">Нет данных для отображения</div>
        </div>
      </Card>
    );
  }

  return (
    <Card title={title} className={`dashboard-area-chart ${className}`}>
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          {gradients}
          
          {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />}
          
          <XAxis 
            dataKey="displayDate"
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12, fill: '#8c8c8c' }}
          />
          
          <YAxis
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12, fill: '#8c8c8c' }}
            tickFormatter={formatYAxisLabel}
          />
          
          <Tooltip
            formatter={tooltipFormatter}
            labelStyle={{ color: '#262626' }}
            contentStyle={{
              border: 'none',
              borderRadius: '8px',
              boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)'
            }}
          />
          
          {showLegend && (
            <Legend
              wrapperStyle={{ paddingTop: '20px' }}
              iconType="rect"
            />
          )}
          
          {areas.map((area, index) => (
            <Area
              key={area.dataKey}
              type="monotone"
              dataKey={area.dataKey}
              stackId={stacked ? "1" : undefined}
              stroke={area.color}
              strokeWidth={area.strokeWidth}
              fill={`url(#gradient-${index})`}
              fillOpacity={1}
              name={area.name}
            />
          ))}
        </AreaChart>
      </ResponsiveContainer>
    </Card>
  );
});

DashboardAreaChart.displayName = 'DashboardAreaChart';

export default DashboardAreaChart; 