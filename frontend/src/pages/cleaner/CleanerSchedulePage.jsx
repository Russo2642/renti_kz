import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  Card, Typography, Form, TimePicker, Button, message, Spin, 
  Row, Col, Switch, Space, Divider, Tag, Alert
} from 'antd';
import {
  CalendarOutlined,
  ClockCircleOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import { cleanerAPI } from '../../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

const CleanerSchedulePage = () => {
  const [form] = Form.useForm();
  const [hasChanges, setHasChanges] = useState(false);
  const queryClient = useQueryClient();

  // Получение текущего расписания
  const { data: profile, isLoading } = useQuery({
    queryKey: ['cleaner-profile'],
    queryFn: () => cleanerAPI.getProfile()
  });

  // Эффект для обновления формы при загрузке данных
  React.useEffect(() => {
    const formValues = {};
    const daysKeys = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
    
    // Инициализируем все дни как выходные по умолчанию
    daysKeys.forEach(day => {
      formValues[day] = {
        enabled: false,
        start_time: null,
        end_time: null,
      };
    });
    
    // Если есть данные расписания, обновляем соответствующие дни
    if (profile?.data?.schedule) {
      const schedule = profile.data.schedule;
      
      Object.entries(schedule).forEach(([day, hoursArray]) => {
        if (hoursArray && Array.isArray(hoursArray) && hoursArray.length > 0) {
          const firstHours = hoursArray[0]; // Берем первый интервал
          if (firstHours && firstHours.start && firstHours.end) {
            formValues[day] = {
              enabled: true,
              start_time: dayjs(firstHours.start, 'HH:mm'),
              end_time: dayjs(firstHours.end, 'HH:mm'),
            };
          }
        }
      });
    }
    
    form.setFieldsValue(formValues);
  }, [profile, form]);

  // Мутация для обновления расписания
  const updateScheduleMutation = useMutation({
    mutationFn: cleanerAPI.updateSchedule,
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaner-profile']);
      setHasChanges(false);
      message.success('Расписание успешно обновлено');
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Ошибка обновления расписания');
    }
  });

  const onFinish = (values) => {
    const schedule = {};
    
    Object.entries(values).forEach(([day, dayData]) => {
      if (dayData && dayData.enabled && dayData.start_time && dayData.end_time) {
        schedule[day] = [{
          start: dayData.start_time.format('HH:mm'),
          end: dayData.end_time.format('HH:mm'),
        }];
      } else {
        schedule[day] = [];
      }
    });

    console.log('Отправляемое расписание:', schedule); // Для отладки
    updateScheduleMutation.mutate(schedule);
  };

  const onValuesChange = () => {
    setHasChanges(true);
  };

  const daysOfWeek = [
    { key: 'monday', label: 'Понедельник', short: 'Пн' },
    { key: 'tuesday', label: 'Вторник', short: 'Вт' },
    { key: 'wednesday', label: 'Среда', short: 'Ср' },
    { key: 'thursday', label: 'Четверг', short: 'Чт' },
    { key: 'friday', label: 'Пятница', short: 'Пт' },
    { key: 'saturday', label: 'Суббота', short: 'Сб' },
    { key: 'sunday', label: 'Воскресенье', short: 'Вс' }
  ];

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <Title level={2}>Мое расписание</Title>
          <Text type="secondary">
            Настройте свое рабочее расписание для планирования уборок
          </Text>
        </div>
        <CalendarOutlined className="text-2xl text-blue-500" />
      </div>

      <Alert
        message="Информация о расписании"
        description="Установите время, когда вы доступны для уборки квартир. Это поможет системе правильно планировать задачи."
        type="info"
        showIcon
        className="mb-6"
      />

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          onValuesChange={onValuesChange}
        >
          <Row gutter={[16, 24]}>
            {daysOfWeek.map((day) => (
              <Col xs={24} sm={12} lg={8} key={day.key}>
                <Card size="small" className="h-full">
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <Text strong className="text-lg">{day.label}</Text>
                      <Form.Item
                        name={[day.key, 'enabled']}
                        valuePropName="checked"
                        className="mb-0"
                      >
                        <Switch 
                          size="small"
                          checkedChildren="Работаю"
                          unCheckedChildren="Выходной"
                        />
                      </Form.Item>
                    </div>

                    <Form.Item dependencies={[[day.key, 'enabled']]}>
                      {({ getFieldValue }) => {
                        const enabled = getFieldValue([day.key, 'enabled']);
                        return enabled ? (
                          <div className="space-y-3">
                            <Form.Item
                              name={[day.key, 'start_time']}
                              label="Начало работы"
                              rules={[
                                { required: true, message: 'Укажите время начала' }
                              ]}
                            >
                              <TimePicker
                                format="HH:mm"
                                placeholder="09:00"
                                className="w-full"
                              />
                            </Form.Item>

                            <Form.Item
                              name={[day.key, 'end_time']}
                              label="Конец работы"
                              rules={[
                                { required: true, message: 'Укажите время окончания' }
                              ]}
                            >
                              <TimePicker
                                format="HH:mm"
                                placeholder="18:00"
                                className="w-full"
                              />
                            </Form.Item>
                          </div>
                        ) : (
                          <div className="text-center py-4">
                            <Text type="secondary">Выходной день</Text>
                          </div>
                        );
                      }}
                    </Form.Item>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>

          <Divider />

          {/* Текущее расписание */}
          <div className="mb-6">
            <Title level={4}>Текущее расписание</Title>
            <div className="flex flex-wrap gap-2">
              {daysOfWeek.map((day) => {
                const scheduleArray = profile?.data?.schedule?.[day.key];
                const hasSchedule = scheduleArray && Array.isArray(scheduleArray) && scheduleArray.length > 0;
                const firstSchedule = hasSchedule ? scheduleArray[0] : null;
                
                return (
                  <Tag 
                    key={day.key}
                    color={hasSchedule ? 'blue' : 'default'}
                    className="mb-2"
                  >
                    <ClockCircleOutlined className="mr-1" />
                    {day.short}: {hasSchedule && firstSchedule ? `${firstSchedule.start} - ${firstSchedule.end}` : 'Выходной'}
                  </Tag>
                );
              })}
            </div>
          </div>

          <div className="flex justify-end space-x-2">
            <Button 
              onClick={() => {
                form.resetFields();
                setHasChanges(false);
              }}
              disabled={!hasChanges}
            >
              Отменить изменения
            </Button>
            <Button 
              type="primary" 
              htmlType="submit"
              loading={updateScheduleMutation.isPending}
              icon={<SaveOutlined />}
              disabled={!hasChanges}
            >
              Сохранить расписание
            </Button>
          </div>
        </Form>
      </Card>

      {/* Быстрые шаблоны */}
      <Card title="Быстрые шаблоны">
        <Space wrap>
          <Button
            size="small"
            onClick={() => {
              const workdaySchedule = {};
              ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'].forEach(day => {
                workdaySchedule[day] = {
                  enabled: true,
                  start_time: dayjs('09:00', 'HH:mm'),
                  end_time: dayjs('18:00', 'HH:mm'),
                };
              });
              ['saturday', 'sunday'].forEach(day => {
                workdaySchedule[day] = {
                  enabled: false,
                  start_time: null,
                  end_time: null,
                };
              });
              form.setFieldsValue(workdaySchedule);
              setHasChanges(true);
            }}
          >
            Рабочие дни (Пн-Пт, 9:00-18:00)
          </Button>
          
          <Button
            size="small"
            onClick={() => {
              const fullWeekSchedule = {};
              daysOfWeek.forEach(day => {
                fullWeekSchedule[day.key] = {
                  enabled: true,
                  start_time: dayjs('08:00', 'HH:mm'),
                  end_time: dayjs('20:00', 'HH:mm'),
                };
              });
              form.setFieldsValue(fullWeekSchedule);
              setHasChanges(true);
            }}
          >
            Вся неделя (8:00-20:00)
          </Button>
          
          <Button
            size="small"
            onClick={() => {
              const weekendSchedule = {};
              ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'].forEach(day => {
                weekendSchedule[day] = {
                  enabled: false,
                  start_time: null,
                  end_time: null,
                };
              });
              ['saturday', 'sunday'].forEach(day => {
                weekendSchedule[day] = {
                  enabled: true,
                  start_time: dayjs('10:00', 'HH:mm'),
                  end_time: dayjs('16:00', 'HH:mm'),
                };
              });
              form.setFieldsValue(weekendSchedule);
              setHasChanges(true);
            }}
          >
            Только выходные (10:00-16:00)
          </Button>
          
          <Button
            size="small"
            danger
            onClick={() => {
              const emptySchedule = {};
              daysOfWeek.forEach(day => {
                emptySchedule[day.key] = {
                  enabled: false,
                  start_time: null,
                  end_time: null,
                };
              });
              form.setFieldsValue(emptySchedule);
              setHasChanges(true);
            }}
          >
            Очистить все
          </Button>
        </Space>
      </Card>
    </div>
  );
};

export default CleanerSchedulePage;
