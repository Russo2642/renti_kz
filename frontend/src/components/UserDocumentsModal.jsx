import React, { useState } from 'react';
import { Modal, Row, Col, Card, Image, Descriptions, Tag, Button, Space, Typography } from 'antd';
import { 
  UserOutlined, 
  IdcardOutlined, 
  CheckCircleOutlined, 
  CloseCircleOutlined,
  EyeOutlined 
} from '@ant-design/icons';

const { Title, Text } = Typography;

const UserDocumentsModal = ({ visible, onClose, booking }) => {
  const [selectedImage, setSelectedImage] = useState(null);
  const [imagePreviewVisible, setImagePreviewVisible] = useState(false);

  if (!booking || !booking.renter) {
    return null;
  }

  const { renter } = booking;
  const { user, document_type, document_url, photo_with_doc_url, verification_status } = renter;

  const getVerificationStatusColor = (status) => {
    const colors = {
      'approved': 'success',
      'pending': 'processing',
      'rejected': 'error',
      'unverified': 'default'
    };
    return colors[status] || 'default';
  };

  const getVerificationStatusText = (status) => {
    const texts = {
      'approved': 'Одобрено',
      'pending': 'На проверке',
      'rejected': 'Отклонено',
      'unverified': 'Не проверено'
    };
    return texts[status] || status;
  };

  const getVerificationStatusIcon = (status) => {
    const icons = {
      'approved': <CheckCircleOutlined />,
      'rejected': <CloseCircleOutlined />,
    };
    return icons[status] || null;
  };

  const getDocumentTypeText = (type) => {
    const types = {
      'passport': 'Паспорт',
      'udv': 'УДВ (Удостоверение личности)'
    };
    return types[type] || type;
  };

  const handleImageClick = (imageUrl) => {
    setSelectedImage(imageUrl);
    setImagePreviewVisible(true);
  };

  const hasDocuments = document_url && (document_url.page1 || document_url.page2);

  return (
    <>
      <Modal
        title={
          <div className="flex items-center space-x-2">
            <UserOutlined />
            <span>Документы пользователя</span>
          </div>
        }
        open={visible}
        onCancel={onClose}
        footer={[
          <Button key="close" onClick={onClose}>
            Закрыть
          </Button>
        ]}
        width={800}
        style={{ top: 20 }}
      >
        <div className="space-y-6">
          {/* Информация о пользователе */}
          <Card size="small">
            <Descriptions column={2} size="small">
              <Descriptions.Item label="ФИО">
                <Text strong>
                  {user?.first_name} {user?.last_name}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="Телефон">
                {user?.phone}
              </Descriptions.Item>
              <Descriptions.Item label="Email">
                {user?.email || 'Не указан'}
              </Descriptions.Item>
              <Descriptions.Item label="ИИН">
                {user?.iin || 'Не указан'}
              </Descriptions.Item>
              <Descriptions.Item label="Тип документа">
                <Space>
                  <IdcardOutlined />
                  {getDocumentTypeText(document_type)}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="Статус верификации">
                <Tag 
                  color={getVerificationStatusColor(verification_status)}
                  icon={getVerificationStatusIcon(verification_status)}
                >
                  {getVerificationStatusText(verification_status)}
                </Tag>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          {/* Документы */}
          {hasDocuments ? (
            <div>
              <Title level={4}>Сканы документов</Title>
              <Row gutter={[16, 16]}>
                {/* Первая страница */}
                {document_url.page1 && (
                  <Col xs={24} sm={12} md={8}>
                    <Card
                      size="small"
                      title="Страница 1"
                      hoverable
                      cover={
                        <div 
                          className="cursor-pointer relative group"
                          onClick={() => handleImageClick(document_url.page1)}
                        >
                          <Image
                            alt="Документ страница 1"
                            src={document_url.page1}
                            preview={false}
                            style={{ height: 200, objectFit: 'cover' }}
                            fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMIAAADDCAYAAADQvc6UAAABRWlDQ1BJQ0MgUHJvZmlsZQAAKJFjYGASSSwoyGFhYGDIzSspCnJ3UoiIjFJgf8LAwSDCIMogwMCcmFxc4BgQ4ANUwgCjUcG3awyMIPqyLsis7PPOq3QdDFcvjV3jOD1boQVTPQrgSkktTgbSf4A4LbmgqISBgTEFyFYuLykAsTuAbJEioKOA7DkgdjqEvQHEToKwj4DVhAQ5A9k3gGyB5IxEoBmML4BsnSQk8XQkNtReEOBxcfXxUQg1Mjc0dyHgXNJBSWpFCYh2zi+oLMpMzyhRcASGUqqCZ16yno6CkYGRAQMDKMwhqj/fAIcloxgHQqxAjIHBEugw5sUIsSQpBobtQPdLciLEVJYzMPBHMDBsayhILEqEO4DxG0txmrERhM29nYGBddr//5/DGRjYNRkY/l7////39v///y4Dmn+LgeHANwDrkl1AuO+pmgAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAwqADAAQAAAABAAAAwwAAAAD9b/HnAAAHlklEQVR4Ae3dP3Ik1RnG4W+FgYxN"
                          />
                          <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-20 transition-all duration-200 flex items-center justify-center">
                            <EyeOutlined className="text-white text-2xl opacity-0 group-hover:opacity-100 transition-opacity" />
                          </div>
                        </div>
                      }
                    />
                  </Col>
                )}

                {/* Вторая страница (для УДВ) */}
                {document_url.page2 && (
                  <Col xs={24} sm={12} md={8}>
                    <Card
                      size="small"
                      title="Страница 2"
                      hoverable
                      cover={
                        <div 
                          className="cursor-pointer relative group"
                          onClick={() => handleImageClick(document_url.page2)}
                        >
                          <Image
                            alt="Документ страница 2"
                            src={document_url.page2}
                            preview={false}
                            style={{ height: 200, objectFit: 'cover' }}
                            fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMIAAADDCAYAAADQvc6UAAABRWlDQ1BJQ0MgUHJvZmlsZQAAKJFjYGASSSwoyGFhYGDIzSspCnJ3UoiIjFJgf8LAwSDCIMogwMCcmFxc4BgQ4ANUwgCjUcG3awyMIPqyLsis7PPOq3QdDFcvjV3jOD1boQVTPQrgSkktTgbSf4A4LbmgqISBgTEFyFYuLykAsTuAbJEioKOA7DkgdjqEvQHEToKwj4DVhAQ5A9k3gGyB5IxEoBmML4BsnSQk8XQkNtReEOBxcfXxUQg1Mjc0dyHgXNJBSWpFCYh2zi+oLMpMzyhRcASGUqqCZ16yno6CkYGRAQMDKMwhqj/fAIcloxgHQqxAjIHBEugw5sUIsSQpBobtQPdLciLEVJYzMPBHMDBsayhILEqEO4DxG0txmrERhM29nYGBddr//5/DGRjYNRkY/l7////39v///y4Dmn+LgeHANwDrkl1AuO+pmgAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAwqADAAQAAAABAAAAwwAAAAD9b/HnAAAHlklEQVR4Ae3dP3Ik1RnG4W+FgYxN"
                          />
                          <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-20 transition-all duration-200 flex items-center justify-center">
                            <EyeOutlined className="text-white text-2xl opacity-0 group-hover:opacity-100 transition-opacity" />
                          </div>
                        </div>
                      }
                    />
                  </Col>
                )}

                {/* Фото с документом */}
                {photo_with_doc_url && (
                  <Col xs={24} sm={12} md={8}>
                    <Card
                      size="small"
                      title="Фото с документом"
                      hoverable
                      cover={
                        <div 
                          className="cursor-pointer relative group"
                          onClick={() => handleImageClick(photo_with_doc_url)}
                        >
                          <Image
                            alt="Фото с документом"
                            src={photo_with_doc_url}
                            preview={false}
                            style={{ height: 200, objectFit: 'cover' }}
                            fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMIAAADDCAYAAADQvc6UAAABRWlDQ1BJQ0MgUHJvZmlsZQAAKJFjYGASSSwoyGFhYGDIzSspCnJ3UoiIjFJgf8LAwSDCIMogwMCcmFxc4BgQ4ANUwgCjUcG3awyMIPqyLsis7PPOq3QdDFcvjV3jOD1boQVTPQrgSkktTgbSf4A4LbmgqISBgTEFyFYuLykAsTuAbJEioKOA7DkgdjqEvQHEToKwj4DVhAQ5A9k3gGyB5IxEoBmML4BsnSQk8XQkNtReEOBxcfXxUQg1Mjc0dyHgXNJBSWpFCYh2zi+oLMpMzyhRcASGUqqCZ16yno6CkYGRAQMDKMwhqj/fAIcloxgHQqxAjIHBEugw5sUIsSQpBobtQPdLciLEVJYzMPBHMDBsayhILEqEO4DxG0txmrERhM29nYGBddr//5/DGRjYNRkY/l7////39v///y4Dmn+LgeHANwDrkl1AuO+pmgAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAwqADAAQAAAABAAAAwwAAAAD9b/HnAAAHlklEQVR4Ae3dP3Ik1RnG4W+FgYxN"
                          />
                          <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-20 transition-all duration-200 flex items-center justify-center">
                            <EyeOutlined className="text-white text-2xl opacity-0 group-hover:opacity-100 transition-opacity" />
                          </div>
                        </div>
                      }
                    />
                  </Col>
                )}
              </Row>
            </div>
          ) : (
            <Card>
              <div className="text-center py-8">
                <IdcardOutlined className="text-4xl text-gray-400 mb-4" />
                <Title level={4} type="secondary">
                  Документы не загружены
                </Title>
                <Text type="secondary">
                  Пользователь не загрузил документы для верификации
                </Text>
              </div>
            </Card>
          )}
        </div>
      </Modal>

      {/* Модал предварительного просмотра изображения */}
      <Modal
        open={imagePreviewVisible}
        footer={null}
        onCancel={() => setImagePreviewVisible(false)}
        width="90%"
        style={{ top: 20 }}
        centered
      >
        {selectedImage && (
          <Image
            alt="Предварительный просмотр"
            src={selectedImage}
            style={{ width: '100%' }}
          />
        )}
      </Modal>
    </>
  );
};

export default UserDocumentsModal;
