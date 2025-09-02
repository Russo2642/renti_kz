import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  Card, Row, Col, Typography, Tag, Button, Space, Modal, Form, 
  Input, message, Spin, Empty, Descriptions, Badge, Tooltip, 
  Popconfirm, Select
} from 'antd';
import {
  HomeOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  PauseCircleOutlined,
  ToolOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { cleanerAPI } from '../../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

const CleanerCleaningPage = () => {
  const [startCleaningModalVisible, setStartCleaningModalVisible] = useState(false);
  const [completeCleaningModalVisible, setCompleteCleaningModalVisible] = useState(false);
  const [selectedApartment, setSelectedApartment] = useState(null);
  const [startForm] = Form.useForm();
  const [completeForm] = Form.useForm();
  const queryClient = useQueryClient();

  // Получение квартир для уборки
  const { data: apartmentsForCleaning, isLoading } = useQuery({
    queryKey: ['cleaner-apartments-for-cleaning'],
    queryFn: () => cleanerAPI.getApartmentsForCleaning()
  });

  // Мутация для начала уборки
  const startCleaningMutation = useMutation({
    mutationFn: cleanerAPI.startCleaning,
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaner-apartments-for-cleaning']);
      queryClient.invalidateQueries(['cleaner-stats']);
      setStartCleaningModalVisible(false);
      startForm.resetFields();
      message.success('Уборка начата');
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Ошибка начала уборки');
    }
  });

  // Мутация для завершения уборки
  const completeCleaningMutation = useMutation({
    mutationFn: cleanerAPI.completeCleaning,
    onSuccess: () => {
      queryClient.invalidateQueries(['cleaner-apartments-for-cleaning']);
      queryClient.invalidateQueries(['cleaner-stats']);
      setCompleteCleaningModalVisible(false);
      completeForm.resetFields();
      message.success('Уборка завершена, квартира освобождена');
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Ошибка завершения уборки');
    }
  });

  const getCleaningStatusColor = (status) => {
    const colors = {
      'needs_cleaning': 'orange',
      'cleaning_in_progress': 'blue',
      'cleaned': 'green',
      'cleaning_overdue': 'red'
    };
    return colors[status] || 'default';
  };

  const getCleaningStatusText = (status) => {
    const texts = {
      'needs_cleaning': 'Нужна уборка',
      'cleaning_in_progress': 'Убирается',
      'cleaned': 'Убрано',
      'cleaning_overdue': 'Просрочено'
    };
    return texts[status] || status;
  };

  const getCleaningStatusIcon = (status) => {
    const icons = {
      'needs_cleaning': <ClockCircleOutlined />,
      'cleaning_in_progress': <PlayCircleOutlined />,
      'cleaned': <CheckCircleOutlined />,
      'cleaning_overdue': <ExclamationCircleOutlined />
    };
    return icons[status] || <ToolOutlined />;
  };

  const handleStartCleaning = (apartment) => {
    setSelectedApartment(apartment);
    startForm.setFieldsValue({
      apartment_id: apartment.id
    });
    setStartCleaningModalVisible(true);
  };

  const handleCompleteCleaning = (apartment) => {
    setSelectedApartment(apartment);
    completeForm.setFieldsValue({
      apartment_id: apartment.id
    });
    setCompleteCleaningModalVisible(true);
  };

  const onStartCleaning = (values) => {
    startCleaningMutation.mutate(values);
  };

  const onCompleteCleaning = (values) => {
    completeCleaningMutation.mutate(values);
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  const apartmentData = apartmentsForCleaning?.data || [];

  if (!apartmentData || apartmentData.length === 0) {
    return (
      <Card>
        <Empty 
          description="Нет квартир для уборки"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <Title level={2}>Квартиры для уборки</Title>
        <Text type="secondary">
          Всего квартир: {apartmentData.length}
        </Text>
      </div>

      <Row gutter={[16, 16]}>
        {apartmentData.map((apartment) => (
          <Col xs={24} lg={12} key={apartment.id}>
            <Card
              className="h-full"
              title={
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <HomeOutlined />
                    <span>
                      {apartment.street}, д. {apartment.building}
                      {apartment.apartment_number && `, кв. ${apartment.apartment_number}`}
                    </span>
                  </div>
                  <Tag 
                    icon={getCleaningStatusIcon(apartment.cleaning_status)}
                    color={getCleaningStatusColor(apartment.cleaning_status)}
                  >
                    {getCleaningStatusText(apartment.cleaning_status)}
                  </Tag>
                </div>
              }
              extra={
                <Space>
                  {apartment.cleaning_status === 'needs_cleaning' || apartment.cleaning_status === 'cleaning_overdue' ? (
                    <Button 
                      type="primary"
                      icon={<PlayCircleOutlined />}
                      size="small"
                      onClick={() => handleStartCleaning(apartment)}
                    >
                      Начать уборку
                    </Button>
                  ) : apartment.cleaning_status === 'cleaning_in_progress' ? (
                    <Button 
                      type="primary"
                      icon={<CheckCircleOutlined />}
                      size="small"
                      onClick={() => handleCompleteCleaning(apartment)}
                    >
                      Завершить уборку
                    </Button>
                  ) : (
                    <Button 
                      disabled
                      icon={<CheckCircleOutlined />}
                      size="small"
                    >
                      Убрано
                    </Button>
                  )}
                </Space>
              }
            >
              <div className="space-y-4">
                <Descriptions size="small" column={1}>
                  <Descriptions.Item label="Тип">
                    {apartment.apartment_type?.name || 'Не указан'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Район">
                    {apartment.district?.name || 'Не указан'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Площадь">
                    {apartment.area ? `${apartment.area} м²` : 'Не указана'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Комнат">
                    {apartment.rooms || 'Не указано'}
                  </Descriptions.Item>
                </Descriptions>

                {/* Информация о последней уборке */}
                {apartment.last_cleaning_date && (
                  <div>
                    <Text strong>Последняя уборка: </Text>
                    <Text>
                      {dayjs(apartment.last_cleaning_date).format('DD.MM.YYYY HH:mm')}
                    </Text>
                  </div>
                )}

                {/* Информация о следующей уборке */}
                {apartment.next_cleaning_date && (
                  <div>
                    <Text strong>Следующая уборка: </Text>
                    <Text>
                      {dayjs(apartment.next_cleaning_date).format('DD.MM.YYYY')}
                    </Text>
                    {dayjs(apartment.next_cleaning_date).isBefore(dayjs(), 'day') && (
                      <Tag color="red" className="ml-2">Просрочено</Tag>
                    )}
                  </div>
                )}

                {/* Информация о текущей уборке */}
                {apartment.current_cleaning && (
                  <div className="bg-blue-50 p-3 rounded">
                    <Text strong>Текущая уборка:</Text>
                    <div className="mt-2 space-y-1">
                      <div>
                        <Text>Начата: </Text>
                        <Text>{dayjs(apartment.current_cleaning.started_at).format('DD.MM.YYYY HH:mm')}</Text>
                      </div>
                      {apartment.current_cleaning.notes && (
                        <div>
                          <Text>Заметки: </Text>
                          <Text>{apartment.current_cleaning.notes}</Text>
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {/* Статус квартиры */}
                <div>
                  <Badge
                    status={apartment.is_free ? 'success' : 'error'}
                    text={apartment.is_free ? 'Свободна' : 'Занята'}
                  />
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* Модал начала уборки */}
      <Modal
        title="Начать уборку"
        open={startCleaningModalVisible}
        onCancel={() => {
          setStartCleaningModalVisible(false);
          startForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        {selectedApartment && (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded">
              <Text strong>Квартира: </Text>
              <Text>
                {selectedApartment.street}, д. {selectedApartment.building}
                {selectedApartment.apartment_number && `, кв. ${selectedApartment.apartment_number}`}
              </Text>
            </div>

            <Form
              form={startForm}
              layout="vertical"
              onFinish={onStartCleaning}
            >
              <Form.Item name="apartment_id" hidden>
                <Input />
              </Form.Item>

              <Form.Item
                name="cleaning_type"
                label="Тип уборки"
                rules={[{ required: true, message: 'Выберите тип уборки' }]}
              >
                <Select placeholder="Выберите тип уборки">
                  <Option value="regular">Обычная уборка</Option>
                  <Option value="deep">Генеральная уборка</Option>
                  <Option value="checkout">Уборка после выезда</Option>
                  <Option value="checkin">Подготовка к заезду</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="notes"
                label="Заметки (необязательно)"
              >
                <TextArea 
                  rows={3}
                  placeholder="Дополнительные заметки о состоянии квартиры или особенностях уборки"
                />
              </Form.Item>

              <div className="flex justify-end space-x-2 pt-4 border-t">
                <Button onClick={() => setStartCleaningModalVisible(false)}>
                  Отмена
                </Button>
                <Button 
                  type="primary" 
                  htmlType="submit"
                  loading={startCleaningMutation.isPending}
                  icon={<PlayCircleOutlined />}
                >
                  Начать уборку
                </Button>
              </div>
            </Form>
          </div>
        )}
      </Modal>

      {/* Модал завершения уборки */}
      <Modal
        title="Завершить уборку"
        open={completeCleaningModalVisible}
        onCancel={() => {
          setCompleteCleaningModalVisible(false);
          completeForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        {selectedApartment && (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded">
              <Text strong>Квартира: </Text>
              <Text>
                {selectedApartment.street}, д. {selectedApartment.building}
                {selectedApartment.apartment_number && `, кв. ${selectedApartment.apartment_number}`}
              </Text>
            </div>

            <Form
              form={completeForm}
              layout="vertical"
              onFinish={onCompleteCleaning}
            >
              <Form.Item name="apartment_id" hidden>
                <Input />
              </Form.Item>

              <Form.Item
                name="quality_rating"
                label="Оценка качества уборки"
                rules={[{ required: true, message: 'Выберите оценку качества' }]}
              >
                <Select placeholder="Оцените качество уборки">
                  <Option value={5}>Отлично (5)</Option>
                  <Option value={4}>Хорошо (4)</Option>
                  <Option value={3}>Удовлетворительно (3)</Option>
                  <Option value={2}>Плохо (2)</Option>
                  <Option value={1}>Очень плохо (1)</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="completion_notes"
                label="Заметки о выполненной работе"
                rules={[{ required: true, message: 'Добавьте заметки о выполненной работе' }]}
              >
                <TextArea 
                  rows={4}
                  placeholder="Опишите выполненную работу, обнаруженные проблемы или особенности"
                />
              </Form.Item>

              <Form.Item
                name="issues_found"
                label="Обнаруженные проблемы (необязательно)"
              >
                <TextArea 
                  rows={2}
                  placeholder="Поломки, недостающие предметы или другие проблемы"
                />
              </Form.Item>

              <div className="flex justify-end space-x-2 pt-4 border-t">
                <Button onClick={() => setCompleteCleaningModalVisible(false)}>
                  Отмена
                </Button>
                <Button 
                  type="primary" 
                  htmlType="submit"
                  loading={completeCleaningMutation.isPending}
                  icon={<CheckCircleOutlined />}
                >
                  Завершить уборку
                </Button>
              </div>
            </Form>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default CleanerCleaningPage;
