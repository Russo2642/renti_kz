import React, { memo, useMemo } from 'react';
import { Card } from 'antd';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';

const DashboardPieChart = memo(({
  title,
  data = {},
  height = 300,
  loading = false,
  colors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#13c2c2', '#eb2f96'],
  showLabels = true,
  showLegend = true,
  showPercentage = true,
  innerRadius = 0, // для создания donut chart
  formatTooltip,
  formatValue,
  className = ''
}) => {
  // Преобразуем объект данных в массив для Recharts
  const chartData = useMemo(() => {
    if (!data || typeof data !== 'object') return [];
    
    const total = Object.values(data).reduce((sum, value) => sum + value, 0);
    
    return Object.entries(data).map(([key, value], index) => ({
      name: key,
      value: value,
      percentage: total > 0 ? ((value / total) * 100).toFixed(1) : 0,
      color: colors[index % colors.length]
    }));
  }, [data, colors]);

  const defaultTooltipFormatter = (value, name, props) => [
    `${formatValue ? formatValue(value) : value.toLocaleString()} ${showPercentage ? `(${props.payload.percentage}%)` : ''}`,
    name
  ];

  const tooltipFormatter = formatTooltip || defaultTooltipFormatter;

  // Кастомная функция для отображения меток
  const renderLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, value, percentage, name }) => {
    if (!showLabels) return null;
    
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    // Показываем метку только если процент больше 5%
    if (parseFloat(percentage) < 5) return null;

    return (
      <text 
        x={x} 
        y={y} 
        fill="white" 
        textAnchor={x > cx ? 'start' : 'end'} 
        dominantBaseline="central"
        fontSize="12"
        fontWeight="bold"
      >
        {showPercentage ? `${percentage}%` : formatValue ? formatValue(value) : value}
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
    <Card title={title} className={`dashboard-pie-chart ${className}`}>
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={renderLabel}
            outerRadius={Math.min(height * 0.35, 120)}
            innerRadius={innerRadius}
            fill="#8884d8"
            dataKey="value"
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.color} />
            ))}
          </Pie>
          
          <Tooltip
            formatter={tooltipFormatter}
            contentStyle={{
              border: 'none',
              borderRadius: '8px',
              boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)'
            }}
          />
          
          {showLegend && (
            <Legend
              verticalAlign="bottom"
              height={36}
              formatter={(value, entry) => (
                <span style={{ color: entry.color }}>
                  {value}
                  {showPercentage && ` (${entry.payload.percentage}%)`}
                </span>
              )}
            />
          )}
        </PieChart>
      </ResponsiveContainer>
    </Card>
  );
});

DashboardPieChart.displayName = 'DashboardPieChart';

export default DashboardPieChart; 