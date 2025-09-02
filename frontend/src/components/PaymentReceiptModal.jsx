import React, { useState } from 'react';
import { Modal, Descriptions, Card, Button, Space, Typography, Alert, Spin, message } from 'antd';
import { 
  FileTextOutlined, 
  CreditCardOutlined, 
  CalendarOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  PrinterOutlined,
  DownloadOutlined
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { bookingsAPI } from '../lib/api.js';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

const PaymentReceiptModal = ({ visible, onClose, bookingId }) => {
  const [printing, setPrinting] = useState(false);

  // Получение данных чека
  const { data: receiptData, isLoading, error } = useQuery({
    queryKey: ['payment-receipt', bookingId],
    queryFn: () => bookingsAPI.getPaymentReceipt(bookingId),
    enabled: visible && !!bookingId,
    retry: false
  });

  const handlePrint = () => {
    setPrinting(true);
    
    // Создаем окно для печати
    const printWindow = window.open('', '_blank');
    
    if (!printWindow) {
      message.error('Не удалось открыть окно для печати');
      setPrinting(false);
      return;
    }

    const printContent = generatePrintableReceipt(receiptData);
    
    printWindow.document.write(printContent);
    printWindow.document.close();
    printWindow.focus();
    
    setTimeout(() => {
      printWindow.print();
      printWindow.close();
      setPrinting(false);
    }, 500);
  };

  const handleDownload = () => {
    if (!receiptData) return;
    
    const receiptText = generateTextReceipt(receiptData);
    const blob = new Blob([receiptText], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `receipt_booking_${receiptData.booking_number || bookingId}.txt`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    URL.revokeObjectURL(url);
    message.success('Чек скачан');
  };

  const generatePrintableReceipt = (receipt) => {
    if (!receipt) return '';

    return `
      <!DOCTYPE html>
      <html>
        <head>
          <meta charset="utf-8">
          <title>Чек об оплате - ${receipt.booking_number}</title>
          <style>
            body {
              font-family: Arial, sans-serif;
              max-width: 600px;
              margin: 0 auto;
              padding: 20px;
              line-height: 1.6;
            }
            .header {
              text-align: center;
              margin-bottom: 30px;
              border-bottom: 2px solid #000;
              padding-bottom: 20px;
            }
            .receipt-info {
              margin-bottom: 20px;
            }
            .row {
              display: flex;
              justify-content: space-between;
              margin-bottom: 10px;
            }
            .row strong {
              font-weight: bold;
            }
            .total-section {
              border-top: 2px solid #000;
              padding-top: 20px;
              margin-top: 20px;
            }
            .total {
              font-size: 18px;
              font-weight: bold;
            }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>ЧЕК ОБ ОПЛАТЕ</h1>
            <h2>RENTI.KZ</h2>
          </div>
          
          <div class="receipt-info">
            <div class="row">
              <span>Номер бронирования:</span>
              <strong>${receipt.booking_number}</strong>
            </div>
            <div class="row">
              <span>Дата платежа:</span>
              <strong>${dayjs(receipt.payment_date).format('DD.MM.YYYY HH:mm')}</strong>
            </div>
            <div class="row">
              <span>ID платежа:</span>
              <strong>${receipt.payment_id}</strong>
            </div>
            <div class="row">
              <span>ID заказа:</span>
              <strong>${receipt.order_id}</strong>
            </div>
            <div class="row">
              <span>Способ оплаты:</span>
              <strong>${receipt.payment_method === 'bankcard' ? 'Банковская карта' : receipt.payment_method}</strong>
            </div>
            <div class="row">
              <span>Карта:</span>
              <strong>${receipt.card_pan || '****-****-****-****'}</strong>
            </div>
          </div>

          <div class="receipt-info">
            <h3>Детали бронирования:</h3>
            <div class="row">
              <span>Квартира:</span>
              <strong>${receipt.booking_details?.apartment_address}</strong>
            </div>
            <div class="row">
              <span>Период:</span>
              <strong>${dayjs(receipt.booking_details?.start_date).format('DD.MM.YYYY HH:mm')} - ${dayjs(receipt.booking_details?.end_date).format('DD.MM.YYYY HH:mm')}</strong>
            </div>
            <div class="row">
              <span>Длительность:</span>
              <strong>${receipt.booking_details?.duration_hours} ч.</strong>
            </div>
            <div class="row">
              <span>Тип аренды:</span>
              <strong>${receipt.booking_details?.rental_type === 'daily' ? 'Посуточно' : receipt.booking_details?.rental_type}</strong>
            </div>
          </div>

          <div class="total-section">
            <div class="row">
              <span>Базовая стоимость:</span>
              <strong>${receipt.amounts?.total_price?.toLocaleString()} ₸</strong>
            </div>
            <div class="row">
              <span>Сервисный сбор:</span>
              <strong>${receipt.amounts?.service_fee?.toLocaleString()} ₸</strong>
            </div>
            <div class="row total">
              <span>ИТОГО К ОПЛАТЕ:</span>
              <strong>${receipt.amounts?.final_price?.toLocaleString()} ₸</strong>
            </div>
          </div>

          <div style="margin-top: 40px; text-align: center; color: #666;">
            <p>Спасибо за использование RENTI.KZ!</p>
            <p>Дата формирования чека: ${dayjs().format('DD.MM.YYYY HH:mm')}</p>
          </div>
        </body>
      </html>
    `;
  };

  const generateTextReceipt = (receipt) => {
    if (!receipt) return '';

    return `
===========================================
               ЧЕК ОБ ОПЛАТЕ
                RENTI.KZ
===========================================

Номер бронирования: ${receipt.booking_number}
Дата платежа: ${dayjs(receipt.payment_date).format('DD.MM.YYYY HH:mm')}
ID платежа: ${receipt.payment_id}
ID заказа: ${receipt.order_id}
Способ оплаты: ${receipt.payment_method === 'bankcard' ? 'Банковская карта' : receipt.payment_method}
Карта: ${receipt.card_pan || '****-****-****-****'}

-------------------------------------------
              ДЕТАЛИ БРОНИРОВАНИЯ
-------------------------------------------

Квартира: ${receipt.booking_details?.apartment_address}
Период: ${dayjs(receipt.booking_details?.start_date).format('DD.MM.YYYY HH:mm')} - ${dayjs(receipt.booking_details?.end_date).format('DD.MM.YYYY HH:mm')}
Длительность: ${receipt.booking_details?.duration_hours} ч.
Тип аренды: ${receipt.booking_details?.rental_type === 'daily' ? 'Посуточно' : receipt.booking_details?.rental_type}

-------------------------------------------
                СУММА К ОПЛАТЕ
-------------------------------------------

Базовая стоимость: ${receipt.amounts?.total_price?.toLocaleString()} ₸
Сервисный сбор: ${receipt.amounts?.service_fee?.toLocaleString()} ₸

ИТОГО К ОПЛАТЕ: ${receipt.amounts?.final_price?.toLocaleString()} ₸

===========================================
       Спасибо за использование RENTI.KZ!
    Дата формирования: ${dayjs().format('DD.MM.YYYY HH:mm')}
===========================================
    `;
  };

  const getPaymentStatusColor = (status) => {
    const colors = {
      'paid': 'success',
      'success': 'success',
      'pending': 'processing',
      'failed': 'error',
      'refunded': 'warning'
    };
    return colors[status] || 'default';
  };

  const getPaymentStatusText = (status) => {
    const texts = {
      'paid': 'Оплачено',
      'success': 'Успешно',
      'pending': 'В обработке',
      'failed': 'Ошибка',
      'refunded': 'Возвращен'
    };
    return texts[status] || status;
  };

  return (
    <Modal
      title={
        <div className="flex items-center space-x-2">
          <FileTextOutlined />
          <span>Чек об оплате</span>
        </div>
      }
      open={visible}
      onCancel={onClose}
      footer={[
        <Button key="close" onClick={onClose}>
          Закрыть
        </Button>,
        receiptData && (
          <Button
            key="download"
            icon={<DownloadOutlined />}
            onClick={handleDownload}
          >
            Скачать
          </Button>
        ),
        receiptData && (
          <Button
            key="print"
            type="primary"
            icon={<PrinterOutlined />}
            loading={printing}
            onClick={handlePrint}
          >
            Печать
          </Button>
        )
      ]}
      width={700}
      style={{ top: 20 }}
    >
      {isLoading && (
        <div className="text-center py-8">
          <Spin size="large" />
          <div className="mt-4">Загружаем данные чека...</div>
        </div>
      )}

      {error && (
        <Alert
          message="Ошибка загрузки чека"
          description={error.response?.data?.error || 'Не удалось загрузить данные чека об оплате'}
          type="error"
          showIcon
        />
      )}

      {receiptData && !isLoading && (
        <div className="space-y-6">
          <Card>
            <div className="text-center mb-4">
              <Title level={3}>ЧЕКА ОБ ОПЛАТЕ</Title>
              <Text type="secondary">RENTI.KZ</Text>
            </div>

            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="Номер бронирования">
                <Text strong className="font-mono">
                  #{receiptData.booking_number}
                </Text>
              </Descriptions.Item>
              
              <Descriptions.Item label="Дата и время платежа">
                <Space>
                  <CalendarOutlined />
                  <Text className="font-mono">
                    {dayjs(receiptData.payment_date).format('DD.MM.YYYY HH:mm')}
                  </Text>
                </Space>
              </Descriptions.Item>

              <Descriptions.Item label="ID платежа">
                <Text className="font-mono">
                  {receiptData.payment_id}
                </Text>
              </Descriptions.Item>

              <Descriptions.Item label="ID заказа">
                <Text className="font-mono">
                  {receiptData.order_id}
                </Text>
              </Descriptions.Item>

              <Descriptions.Item label="Статус платежа">
                <Space>
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                  <Text strong style={{ color: '#52c41a' }}>
                    {getPaymentStatusText(receiptData.status)}
                  </Text>
                </Space>
              </Descriptions.Item>

              <Descriptions.Item label="Способ оплаты">
                <Space>
                  <CreditCardOutlined />
                  <Text>
                    {receiptData.payment_method === 'bankcard' 
                      ? 'Банковская карта' 
                      : receiptData.payment_method
                    }
                  </Text>
                </Space>
              </Descriptions.Item>

              {receiptData.card_pan && (
                <Descriptions.Item label="Номер карты">
                  <Text className="font-mono">
                    {receiptData.card_pan}
                  </Text>
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>

          <Card title="Детали бронирования" size="small">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Квартира">
                <Text strong>
                  {receiptData.booking_details?.apartment_address}
                </Text>
              </Descriptions.Item>
              
              <Descriptions.Item label="Период аренды">
                <Text className="font-mono">
                  {dayjs(receiptData.booking_details?.start_date).format('DD.MM.YYYY HH:mm')} 
                  {' - '}
                  {dayjs(receiptData.booking_details?.end_date).format('DD.MM.YYYY HH:mm')}
                </Text>
              </Descriptions.Item>

              <Descriptions.Item label="Длительность">
                <Text>
                  {receiptData.booking_details?.duration_hours} часов
                </Text>
              </Descriptions.Item>

              <Descriptions.Item label="Тип аренды">
                <Text>
                  {receiptData.booking_details?.rental_type === 'daily' ? 'Посуточно' : receiptData.booking_details?.rental_type}
                </Text>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Card title="Детали оплаты" size="small">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Базовая стоимость">
                <Space>
                  <DollarOutlined />
                  <Text className="font-mono">
                    {receiptData.amounts?.total_price?.toLocaleString()} ₸
                  </Text>
                </Space>
              </Descriptions.Item>
              
              <Descriptions.Item label="Сервисный сбор">
                <Space>
                  <DollarOutlined />
                  <Text className="font-mono">
                    {receiptData.amounts?.service_fee?.toLocaleString()} ₸
                  </Text>
                </Space>
              </Descriptions.Item>

              <Descriptions.Item label="ИТОГО К ОПЛАТЕ">
                <Text strong className="font-mono text-lg" style={{ color: '#52c41a' }}>
                  {receiptData.amounts?.final_price?.toLocaleString()} ₸
                </Text>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <div className="text-center text-gray-500 text-sm">
            <div>Спасибо за использование RENTI.KZ!</div>
            <div>Дата формирования чека: {dayjs().format('DD.MM.YYYY HH:mm')}</div>
          </div>
        </div>
      )}
    </Modal>
  );
};

export default PaymentReceiptModal;
