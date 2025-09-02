import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Table, Button, Card, Tag, Modal, Form, Input, Select, Space,
  Typography, Badge, Tooltip, message, Popconfirm, Switch
} from 'antd';
import {
  SettingOutlined, EditOutlined, DeleteOutlined, PlusOutlined,
  SaveOutlined, CheckCircleOutlined, ExclamationCircleOutlined
} from '@ant-design/icons';
import { platformSettingsAPI } from '../../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const PlatformSettingsPage = () => {
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [selectedSetting, setSelectedSetting] = useState(null);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();
  const queryClient = useQueryClient();

  // Получение всех настроек
  const { data: settingsData, isLoading } = useQuery({
    queryKey: ['platform-settings'],
    queryFn: platformSettingsAPI.getAllSettings
  });

  // Получение процента сервисного сбора
  const { data: serviceFeeData } = useQuery({
    queryKey: ['service-fee'],
    queryFn: platformSettingsAPI.getServiceFee
  });

  // Мутация для создания настройки
  const createMutation = useMutation({
    mutationFn: platformSettingsAPI.createSetting,
    onSuccess: () => {
      queryClient.invalidateQueries(['platform-settings']);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('Настройка создана');
    },
    onError: (error) => {
      message.error('Ошибка при создании настройки');
    }
  });

  // Мутация для обновления настройки
  const updateMutation = useMutation({
    mutationFn: ({ key, ...data }) => platformSettingsAPI.updateSetting(key, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['platform-settings']);
      queryClient.invalidateQueries(['service-fee']);
      setEditModalVisible(false);
      editForm.resetFields();
      message.success('Настройка обновлена');
    },
    onError: (error) => {
      message.error('Ошибка при обновлении настройки');
    }
  });

  // Мутация для удаления настройки
  const deleteMutation = useMutation({
    mutationFn: platformSettingsAPI.deleteSetting,
    onSuccess: () => {
      queryClient.invalidateQueries(['platform-settings']);
      message.success('Настройка удалена');
    },
    onError: (error) => {
      message.error('Ошибка при удалении настройки');
    }
  });

  const handleCreate = (values) => {
    createMutation.mutate(values);
  };

  const handleUpdate = (values) => {
    updateMutation.mutate({
      key: selectedSetting.setting_key,
      ...values
    });
  };

  const handleDelete = (key) => {
    deleteMutation.mutate(key);
  };

  const getDataTypeColor = (dataType) => {
    const colors = {
      'string': 'blue',
      'integer': 'green',
      'decimal': 'orange',
      'boolean': 'purple'
    };
    return colors[dataType] || 'default';
  };

  const getDataTypeText = (dataType) => {
    const texts = {
      'string': 'Строка',
      'integer': 'Целое число',
      'decimal': 'Десятичное',
      'boolean': 'Логическое'
    };
    return texts[dataType] || dataType;
  };

  const formatValue = (value, dataType) => {
    if (dataType === 'boolean') {
      return value === 'true' ? <Badge status="success" text="Да" /> : <Badge status="error" text="Нет" />;
    }
    return value;
  };

  const columns = [
    {
      title: 'Ключ настройки',
      dataIndex: 'setting_key',
      key: 'setting_key',
      render: (key) => (
        <Text strong>{key}</Text>
      ),
    },
    {
      title: 'Значение',
      dataIndex: 'setting_value',
      key: 'setting_value',
      render: (value, record) => formatValue(value, record.data_type),
    },
    {
      title: 'Тип данных',
      dataIndex: 'data_type',
      key: 'data_type',
      render: (dataType) => (
        <Tag color={getDataTypeColor(dataType)}>
          {getDataTypeText(dataType)}
        </Tag>
      ),
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      render: (description) => description || '—',
    },
    {
      title: 'Статус',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? 'Активна' : 'Неактивна'}
        </Tag>
      ),
    },
    {
      title: 'Обновлено',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (date) => date ? dayjs(date).format('DD.MM.YYYY HH:mm') : '—',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedSetting(record);
                editForm.setFieldsValue({
                  setting_value: record.setting_value,
                  description: record.description,
                  data_type: record.data_type,
                  is_active: record.is_active
                });
                setEditModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Удалить">
            <Popconfirm
              title="Удалить настройку?"
              description="Это действие нельзя отменить"
              onConfirm={() => handleDelete(record.setting_key)}
              okText="Да"
              cancelText="Нет"
            >
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div className="p-6">
      <div className="mb-6 flex justify-between items-center">
        <div>
          <Title level={2}>Настройки платформы</Title>
          <Text type="secondary">
            Управление глобальными настройками системы
          </Text>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
        >
          Добавить настройку
        </Button>
      </div>

      {/* Ключевые настройки */}
      <Card className="mb-6" title="Основные настройки">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="text-center p-4 bg-blue-50 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">
              {serviceFeeData?.data?.service_fee_percentage || 0}%
            </div>
            <div className="text-gray-600">Сервисный сбор</div>
          </div>
          <div className="text-center p-4 bg-green-50 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {settingsData?.data?.filter(s => s.is_active).length || 0}
            </div>
            <div className="text-gray-600">Активных настроек</div>
          </div>
          <div className="text-center p-4 bg-orange-50 rounded-lg">
            <div className="text-2xl font-bold text-orange-600">
              {settingsData?.data?.length || 0}
            </div>
            <div className="text-gray-600">Всего настроек</div>
          </div>
        </div>
      </Card>

      {/* Таблица настроек */}
      <Card>
        <Table
          columns={columns}
          dataSource={settingsData?.data || []}
          loading={isLoading}
          rowKey="setting_key"
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `Всего ${total} настроек`,
          }}
        />
      </Card>

      {/* Модальное окно создания */}
      <Modal
        title="Создание новой настройки"
        open={createModalVisible}
        onOk={() => form.submit()}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        okText="Создать"
        cancelText="Отмена"
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreate}
        >
          <Form.Item
            name="setting_key"
            label="Ключ настройки"
            rules={[
              { required: true, message: 'Введите ключ настройки' },
              { pattern: /^[a-z_]+$/, message: 'Используйте только строчные буквы и подчеркивания' }
            ]}
          >
            <Input placeholder="service_fee_percentage" />
          </Form.Item>
          
          <Form.Item
            name="setting_value"
            label="Значение"
            rules={[{ required: true, message: 'Введите значение' }]}
          >
            <Input placeholder="15" />
          </Form.Item>
          
          <Form.Item
            name="data_type"
            label="Тип данных"
            rules={[{ required: true, message: 'Выберите тип данных' }]}
          >
            <Select>
              <Option value="string">Строка</Option>
              <Option value="integer">Целое число</Option>
              <Option value="decimal">Десятичное число</Option>
              <Option value="boolean">Логическое значение</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="description"
            label="Описание"
          >
            <TextArea rows={3} placeholder="Описание назначения настройки" />
          </Form.Item>
          
          <Form.Item
            name="is_active"
            label="Статус"
            valuePropName="checked"
            initialValue={true}
          >
            <Switch checkedChildren="Активна" unCheckedChildren="Неактивна" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Модальное окно редактирования */}
      <Modal
        title="Редактирование настройки"
        open={editModalVisible}
        onOk={() => editForm.submit()}
        onCancel={() => {
          setEditModalVisible(false);
          editForm.resetFields();
          setSelectedSetting(null);
        }}
        okText="Сохранить"
        cancelText="Отмена"
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdate}
        >
          <div className="mb-4 p-3 bg-gray-50 rounded">
            <Text strong>Ключ настройки: </Text>
            <Text code>{selectedSetting?.setting_key}</Text>
          </div>
          
          <Form.Item
            name="setting_value"
            label="Значение"
            rules={[{ required: true, message: 'Введите значение' }]}
          >
            <Input />
          </Form.Item>
          
          <Form.Item
            name="data_type"
            label="Тип данных"
            rules={[{ required: true, message: 'Выберите тип данных' }]}
          >
            <Select>
              <Option value="string">Строка</Option>
              <Option value="integer">Целое число</Option>
              <Option value="decimal">Десятичное число</Option>
              <Option value="boolean">Логическое значение</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="description"
            label="Описание"
          >
            <TextArea rows={3} />
          </Form.Item>
          
          <Form.Item
            name="is_active"
            label="Статус"
            valuePropName="checked"
          >
            <Switch checkedChildren="Активна" unCheckedChildren="Неактивна" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PlatformSettingsPage; 