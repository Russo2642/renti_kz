import React, { useState, useEffect } from 'react';
import { Select, Space } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { locationsAPI } from '../lib/api.js';

const { Option } = Select;

const LocationFilter = ({
  showCity = true,
  showDistrict = true,
  showMicrodistrict = true,
  cityId = null,
  districtId = null,
  microdistrictId = null,
  onCityChange = null,
  onDistrictChange = null,
  onMicrodistrictChange = null,
  size = 'middle',
  className = '',
  layout = 'horizontal', // 'horizontal' or 'vertical'
  allowClear = true,
  showLabels = true, // новый пропс для отображения лейблов
  placeholder = {
    city: 'Выберите город',
    district: 'Выберите район',
    microdistrict: 'Выберите микрорайон'
  }
}) => {
  const [selectedCity, setSelectedCity] = useState(cityId);
  const [selectedDistrict, setSelectedDistrict] = useState(districtId);
  const [selectedMicrodistrict, setSelectedMicrodistrict] = useState(microdistrictId);

  // Получение списка городов
  const { data: cities, isLoading: citiesLoading, error: citiesError } = useQuery({
    queryKey: ['cities'],
    queryFn: locationsAPI.getCities,
    enabled: showCity,
    staleTime: 5 * 60 * 1000, // 5 минут
    onError: (error) => {
      console.error('Ошибка при получении списка городов:', error);
      console.error('Детали ошибки:', {
        message: error.message,
        response: error.response?.data,
        status: error.response?.status,
        config: error.config
      });
    },
    onSuccess: (data) => {
      if (Array.isArray(data)) {
        if (data.length > 0) {
        }
      } else if (data?.data && Array.isArray(data.data)) {
        if (data.data.length > 0) {
        }
      }
      // Добавляем глобальную функцию для тестирования
      window.testCitiesAPI = async () => {
        try {
          const response = await fetch('http://localhost:8080/api/locations/cities');
          const result = await response.json();
          return result;
        } catch (error) {
          console.error('🚨 Ошибка прямого теста:', error);
          return null;
        }
      };
    }
  });

  // Получение списка районов для выбранного города
  const { data: districts, isLoading: districtsLoading } = useQuery({
    queryKey: ['districts', selectedCity],
    queryFn: () => locationsAPI.getDistrictsByCity(selectedCity),
    enabled: showDistrict && !!selectedCity,
    staleTime: 5 * 60 * 1000, // 5 минут
  });

  // Получение списка микрорайонов для выбранного района
  const { data: microdistricts, isLoading: microdistrictsLoading } = useQuery({
    queryKey: ['microdistricts', selectedDistrict],
    queryFn: () => locationsAPI.getMicrodistrictsByDistrict(selectedDistrict),
    enabled: showMicrodistrict && !!selectedDistrict,
    staleTime: 5 * 60 * 1000, // 5 минут
  });

  // Обновление состояния при изменении пропсов
  useEffect(() => {
    setSelectedCity(cityId);
  }, [cityId]);

  useEffect(() => {
    setSelectedDistrict(districtId);
  }, [districtId]);

  useEffect(() => {
    setSelectedMicrodistrict(microdistrictId);
  }, [microdistrictId]);

  // Обработчики изменений
  const handleCityChange = (value) => {
    setSelectedCity(value);
    setSelectedDistrict(null);
    setSelectedMicrodistrict(null);
    
    if (onCityChange) onCityChange(value);
    if (onDistrictChange) onDistrictChange(null);
    if (onMicrodistrictChange) onMicrodistrictChange(null);
  };

  const handleDistrictChange = (value) => {
    setSelectedDistrict(value);
    setSelectedMicrodistrict(null);
    
    if (onDistrictChange) onDistrictChange(value);
    if (onMicrodistrictChange) onMicrodistrictChange(null);
  };

  const handleMicrodistrictChange = (value) => {
    setSelectedMicrodistrict(value);
    
    if (onMicrodistrictChange) onMicrodistrictChange(value);
  };

  const Container = layout === 'horizontal' ? Space : 'div';
  const containerProps = layout === 'horizontal' ? { wrap: true } : { className: `space-y-4 ${className}` };

  return (
    <Container {...containerProps}>
      {showCity && (
        <div className={layout === 'horizontal' ? '' : 'w-full'}>
          {showLabels && (
            <label className="block text-sm font-medium text-gray-700 mb-1">Город</label>
          )}
          <Select
            placeholder={citiesLoading ? "Загрузка городов..." : citiesError ? "Ошибка загрузки городов" : placeholder.city}
            className="w-full"
            allowClear={allowClear}
            size={size}
            value={selectedCity || undefined}
            onChange={handleCityChange}
            loading={citiesLoading}
            status={citiesError ? 'error' : undefined}
            notFoundContent={citiesLoading ? 'Загрузка...' : citiesError ? 'Ошибка загрузки' : 'Нет доступных городов'}
          >
            {(() => {
              // Определяем список городов в зависимости от структуры ответа
              let cityList = [];
              if (Array.isArray(cities)) {
                cityList = cities;
              } else if (cities?.data && Array.isArray(cities.data)) {
                cityList = cities.data;
              }
              
              return cityList.length > 0 ? (
                cityList.map(city => (
                  <Option key={city.id} value={city.id}>
                    {city.name}
                  </Option>
                ))
              ) : (
                !citiesLoading && !citiesError && (
                  <Option disabled>Нет доступных городов</Option>
                )
              );
            })()}
          </Select>
        </div>
      )}

      {showDistrict && (
        <div className={layout === 'horizontal' ? '' : 'w-full'}>
          {showLabels && (
            <label className="block text-sm font-medium text-gray-700 mb-1">Район</label>
          )}
          <Select
            placeholder={placeholder.district}
            className="w-full"
            allowClear={allowClear}
            size={size}
            value={selectedDistrict || undefined}
            onChange={handleDistrictChange}
            loading={districtsLoading}
            disabled={!selectedCity}
          >
            {districts?.data?.map(district => (
              <Option key={district.id} value={district.id}>
                {district.name}
              </Option>
            ))}
          </Select>
        </div>
      )}

      {showMicrodistrict && (
        <div className={layout === 'horizontal' ? '' : 'w-full'}>
          {showLabels && (
            <label className="block text-sm font-medium text-gray-700 mb-1">Микрорайон</label>
          )}
          <Select
            placeholder={placeholder.microdistrict}
            className="w-full"
            allowClear={allowClear}
            size={size}
            value={selectedMicrodistrict || undefined}
            onChange={handleMicrodistrictChange}
            loading={microdistrictsLoading}
            disabled={!selectedDistrict}
          >
            {microdistricts?.data?.map(microdistrict => (
              <Option key={microdistrict.id} value={microdistrict.id}>
                {microdistrict.name}
              </Option>
            ))}
          </Select>
        </div>
      )}
    </Container>
  );
};

export default LocationFilter; 