import { create } from 'zustand';
import dayjs from 'dayjs';

const useDashboardStore = create((set, get) => ({
  // Текущий период фильтрации
  period: 'month', // day, week, month
  
  // Диапазон дат
  dateRange: {
    date_from: dayjs().subtract(6, 'month').format('YYYY-MM-DD'),
    date_to: dayjs().format('YYYY-MM-DD'),
  },
  
  // Автообновление дат при смене периода
  autoUpdateDates: true,

  // Установка периода с автообновлением дат
  setPeriod: (newPeriod) => {
    const { autoUpdateDates } = get();
    
    if (autoUpdateDates) {
      let date_from, date_to;
      const now = dayjs();
      
      switch (newPeriod) {
        case 'day':
          date_from = now.subtract(7, 'day').format('YYYY-MM-DD');
          date_to = now.format('YYYY-MM-DD');
          break;
        case 'week':
          date_from = now.subtract(4, 'week').format('YYYY-MM-DD');
          date_to = now.format('YYYY-MM-DD');
          break;
        case 'month':
          date_from = now.subtract(6, 'month').format('YYYY-MM-DD');
          date_to = now.format('YYYY-MM-DD');
          break;
        default:
          date_from = now.subtract(6, 'month').format('YYYY-MM-DD');
          date_to = now.format('YYYY-MM-DD');
      }
      
      set({ 
        period: newPeriod, 
        dateRange: { date_from, date_to } 
      });
    } else {
      set({ period: newPeriod });
    }
  },

  // Установка диапазона дат
  setDateRange: (date_from, date_to) => {
    set({ 
      dateRange: { date_from, date_to },
      autoUpdateDates: false // Отключаем автообновление при ручной установке
    });
  },

  // Включение/выключение автообновления дат
  setAutoUpdateDates: (auto) => {
    set({ autoUpdateDates: auto });
  },

  // Получение параметров для API запроса
  getApiParams: () => {
    const { period, dateRange } = get();
    return {
      period,
      ...dateRange,
    };
  },

  // Сброс к значениям по умолчанию
  reset: () => {
    set({
      period: 'month',
      dateRange: {
        date_from: dayjs().subtract(6, 'month').format('YYYY-MM-DD'),
        date_to: dayjs().format('YYYY-MM-DD'),
      },
      autoUpdateDates: true,
    });
  },
}));

export default useDashboardStore; 