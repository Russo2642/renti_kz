import React, { memo, useMemo } from 'react';
import { Card } from 'antd';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';

const DashboardLineChart = memo(({
  title,
  data = [],
  dataKeys = [],
  height = 300,
  loading = false,
  colors = ['#1890ff', '#52c41a', '#faad14', '#f5222d'],
  formatTooltip,
  formatXAxisLabel,
  formatYAxisLabel,
  showGrid = true,
  showLegend = true,
  strokeWidth = 2,
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

  // Создаем конфигурацию линий
  const lines = useMemo(() => {
    return dataKeys.map((key, index) => ({
      dataKey: key.dataKey || key,
      name: key.name || key,
      color: key.color || colors[index % colors.length],
      strokeWidth: key.strokeWidth || strokeWidth,
      dot: key.dot !== undefined ? key.dot : false
    }));
  }, [dataKeys, colors, strokeWidth]);

  const defaultTooltipFormatter = (value, name) => [
    `${typeof value === 'number' ? value.toLocaleString() : value}`,
    name
  ];

  const tooltipFormatter = formatTooltip || defaultTooltipFormatter;

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
    <Card title={title} className={`dashboard-line-chart ${className}`}>
      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
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
              iconType="line"
            />
          )}
          
          {lines.map(line => (
            <Line
              key={line.dataKey}
              type="monotone"
              dataKey={line.dataKey}
              stroke={line.color}
              strokeWidth={line.strokeWidth}
              name={line.name}
              dot={line.dot}
              activeDot={{ r: 4, fill: line.color }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </Card>
  );
});

DashboardLineChart.displayName = 'DashboardLineChart';

export default DashboardLineChart; 