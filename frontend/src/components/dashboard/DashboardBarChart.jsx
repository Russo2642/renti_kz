import React, { memo, useMemo } from 'react';
import { Card } from 'antd';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell
} from 'recharts';

const DashboardBarChart = memo(({
  title,
  data = {},
  height = 300,
  loading = false,
  colors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1'],
  formatTooltip,
  formatValue,
  showValues = false,
  className = ''
}) => {
  // Преобразуем объект данных в массив для Recharts
  const chartData = useMemo(() => {
    if (!data || typeof data !== 'object') {
      return [];
    }
    
    const entries = Object.entries(data);
    
    const result = entries.map(([key, value], index) => ({
      name: key,
      value: value,
      color: colors[index % colors.length]
    }));
    
    return result;
  }, [data, colors]);

  const defaultTooltipFormatter = (value, name) => [
    formatValue ? formatValue(value) : value.toLocaleString(),
    'Количество'
  ];

  const tooltipFormatter = formatTooltip || defaultTooltipFormatter;

  const CustomizedLabel = ({ x, y, width, height, value }) => {
    if (!showValues) return null;
    
    return (
      <text 
        x={x + width / 2} 
        y={y - 5} 
        fill="#666" 
        textAnchor="middle"
        dominantBaseline="auto"
        fontSize="12"
      >
        {formatValue ? formatValue(value) : value}
      </text>
    );
  };

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
    <Card title={title} className={`dashboard-bar-chart ${className}`}>
      <ResponsiveContainer width="100%" height={height}>
        <BarChart 
          data={chartData} 
          margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          
          <XAxis 
            dataKey="name"
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12, fill: '#8c8c8c' }}
            interval={0}
            angle={chartData.length > 5 ? -45 : 0}
            textAnchor={chartData.length > 5 ? 'end' : 'middle'}
            height={chartData.length > 5 ? 60 : 30}
          />
          <YAxis
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12, fill: '#8c8c8c' }}
            tickFormatter={formatValue}
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
          
          <Bar 
            dataKey="value" 
            radius={[4, 4, 0, 0]}
            label={showValues ? <CustomizedLabel /> : false}
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.color} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </Card>
  );
});

DashboardBarChart.displayName = 'DashboardBarChart';

export default DashboardBarChart; 