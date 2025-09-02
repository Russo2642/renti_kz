import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  Table, 
  Button, 
  Space, 
  Card, 
  Typography, 
  Modal, 
  Form, 
  Input, 
  message, 
  Popconfirm,
  Tooltip
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  EyeOutlined
} from '@ant-design/icons';
import { apartmentTypesAPI } from '../../lib/api';

const { Title } = Typography;

const ApartmentTypesPage = () => {
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingType, setEditingType] = useState(null);
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  // Получение всех типов
  const { data: apartmentTypes, isLoading } = useQuery({
    queryKey: ['apartmentTypes', 'admin'],
    queryFn: apartmentTypesAPI.adminGetAll,
  });

  // Мутация создания
  const createMutation = useMutation({
    mutationFn: apartmentTypesAPI.create,
    onSuccess: () => {
      message.success('Тип квартиры создан успешно');
      setIsModalVisible(false);
      form.resetFields();
      queryClient.invalidateQueries(['apartmentTypes']);
    },
    onError: (error) => {
      message.error(error.error || 'Ошибка создания типа квартиры');
    },
  });

  // Мутация обновления
  const updateMutation = useMutation({
    mutationFn: ({ id, data }) => apartmentTypesAPI.update(id, data),
    onSuccess: () => {
      message.success('Тип квартиры обновлен успешно');
      setIsModalVisible(false);
      setEditingType(null);
      form.resetFields();
      queryClient.invalidateQueries(['apartmentTypes']);
    },
    onError: (error) => {
      message.error(error.error || 'Ошибка обновления типа квартиры');
    },
  });

  // Мутация удаления
  const deleteMutation = useMutation({
    mutationFn: apartmentTypesAPI.delete,
    onSuccess: () => {
      message.success('Тип квартиры удален успешно');
      queryClient.invalidateQueries(['apartmentTypes']);
    },
    onError: (error) => {
      message.error(error.error || 'Ошибка удаления типа квартиры');
    },
  });



  // Обработчики
  const handleCreate = () => {
    setEditingType(null);
    setIsModalVisible(true);
    form.resetFields();
  };

  const handleEdit = (record) => {
    setEditingType(record);
    setIsModalVisible(true);
    form.setFieldsValue(record);
  };

  const handleSubmit = async (values) => {
    try {
      if (editingType) {
        await updateMutation.mutateAsync({ 
          id: editingType.id, 
          data: values 
        });
      } else {
        await createMutation.mutateAsync(values);
      }
    } catch (error) {
      console.error('Submit error:', error);
    }
  };

  const handleDelete = (id) => {
    deleteMutation.mutate(id);
  };



  // Колонки таблицы
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      render: (name) => <strong>{name}</strong>,
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      ellipsis: {
        showTitle: false,
      },
      render: (description) => (
        <Tooltip placement="topLeft" title={description}>
          {description}
        </Tooltip>
      ),
    },

    {
      title: 'Действия',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="Редактировать">
            <Button 
              type="text" 
              icon={<EditOutlined />} 
              size="small"
              onClick={() => handleEdit(record)}
            />
          </Tooltip>

          <Tooltip title="Удалить">
            <Popconfirm
              title="Удалить тип квартиры?"
              description="Это действие нельзя отменить"
              onConfirm={() => handleDelete(record.id)}
              okText="Да"
              cancelText="Нет"
            >
              <Button 
                type="text" 
                icon={<DeleteOutlined />} 
                size="small"
                danger
              />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Title level={2}>Типы квартир</Title>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            Добавить тип
          </Button>
        </div>

        <Table
          dataSource={apartmentTypes}
          columns={columns}
          loading={isLoading}
          rowKey="id"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} из ${total} типов`,
          }}
        />
      </Card>

      {/* Модальное окно создания/редактирования */}
      <Modal
        title={editingType ? 'Редактировать тип квартиры' : 'Создать тип квартиры'}
        open={isModalVisible}
        onCancel={() => {
          setIsModalVisible(false);
          setEditingType(null);
          form.resetFields();
        }}
        width={600}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            label="Название"
            name="name"
            rules={[
              { required: true, message: 'Введите название типа' },
              { min: 2, max: 100, message: 'От 2 до 100 символов' }
            ]}
          >
            <Input placeholder="Эконом" />
          </Form.Item>

          <Form.Item
            label="Описание"
            name="description"
            rules={[{ max: 500, message: 'Максимум 500 символов' }]}
          >
            <Input.TextArea 
              rows={3}
              placeholder="Описание типа квартиры"
            />
          </Form.Item>

          <Form.Item style={{ marginTop: '24px', marginBottom: 0 }}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={() => {
                setIsModalVisible(false);
                setEditingType(null);
                form.resetFields();
              }}>
                Отмена
              </Button>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={createMutation.isLoading || updateMutation.isLoading}
              >
                {editingType ? 'Обновить' : 'Создать'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ApartmentTypesPage;
