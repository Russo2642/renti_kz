import axios from 'axios';
import { toast } from 'react-hot-toast';

const API_BASE_URL = 'https://api.renti.kz/api';

// Создаем экземпляр axios
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor для добавления токена авторизации
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Interceptor для обработки ответов
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  async (error) => {
    const { response, config } = error;
    
    if (response?.status === 401) {
      // Проверяем, является ли это запросом на логин
      if (config?.url?.includes('/auth/login')) {
        // Для ошибок логина не показываем toast и не делаем редирект
        // Позволяем LoginPage обработать ошибку самостоятельно
        return Promise.reject(error);
      }
      
      // Токен истек, пытаемся обновить
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        try {
          const refreshResponse = await axios.post(`${API_BASE_URL}/auth/refresh`, {
            refresh_token: refreshToken
          });
          
          const { access_token, refresh_token: newRefreshToken } = refreshResponse.data.data;
          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', newRefreshToken);
          
          // Повторяем оригинальный запрос
          error.config.headers.Authorization = `Bearer ${access_token}`;
          return api.request(error.config);
        } catch (refreshError) {
          // Не удалось обновить токен, перенаправляем на логин
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          window.location.href = '/admin/login';
          return Promise.reject(refreshError);
        }
      } else {
        // Нет refresh токена, перенаправляем на логин
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = '/admin/login';
      }
    }
    
    // Не показываем toast для ошибок логина - пусть LoginPage сам обработает
    if (!config?.url?.includes('/auth/login')) {
      const errorMessage = response?.data?.error || 'Произошла ошибка';
      toast.error(errorMessage);
    }
    
    return Promise.reject(error);
  }
);

// API функции для авторизации
export const authAPI = {
  login: (phone, password) => api.post('/auth/login', { phone, password }),
  register: (userData) => api.post('/auth/register', userData),
  logout: (accessToken) => api.post('/auth/logout', { access_token: accessToken }),
  getCurrentUser: () => api.get('/users/me'),
  checkPhone: (phone) => api.get(`/auth/check-phone/${phone}`),
};

// API функции для пользователей
export const usersAPI = {
  getMe: () => api.get('/users/me'),
  updateProfile: (data) => api.put('/users/me', data),
  changePassword: (data) => api.put('/users/password', data),
  deleteAccount: () => api.delete('/users/me'),
  updateRenterVerificationStatus: (renterId, status) => 
    api.put(`/admin/renters/${renterId}/verification-status`, { status }),
  deleteUserByAdmin: (userId) => api.delete(`/admin/users/${userId}`),
  
  // Админские функции
  adminGetAllUsers: (params) => api.get('/admin/users', { params }),
  adminGetUsersStatistics: () => api.get('/admin/users/statistics'),
  adminGetUserById: (userId) => api.get(`/admin/users/${userId}`),
  adminUpdateUserRole: (userId, role) => api.put(`/admin/users/${userId}/role`, { role }),
  adminUpdateUserStatus: (userId, isActive, reason) => 
    api.put(`/admin/users/${userId}/status`, { is_active: isActive, reason }),
  adminGetUserBookingHistory: (userId, params) => api.get(`/admin/users/${userId}/booking-history`, { params }),
};

// API функции для квартир
export const apartmentsAPI = {
  getAll: (params) => api.get('/apartments', { params }),
  getById: (id) => api.get(`/apartments/${id}`),
  create: (data) => api.post('/apartments', data),
  update: (id, data) => api.put(`/apartments/${id}`, data),
  delete: (id) => api.delete(`/apartments/${id}`),
  getPhotos: (id) => api.get(`/apartments/${id}/photos`),
  addPhotos: (id, formData) => api.post(`/apartments/${id}/photos`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  deletePhoto: (photoId) => api.delete(`/apartments/photos/${photoId}`),
  getLocation: (id) => api.get(`/apartments/${id}/location`),
  addLocation: (id, location) => api.post(`/apartments/${id}/location`, location),
  updateLocation: (id, location) => api.put(`/apartments/${id}/location`, location),
  getDocuments: (id) => api.get(`/apartments/${id}/documents`),
  addDocuments: (id, formData) => api.post(`/apartments/${id}/documents`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  deleteDocument: (documentId) => api.delete(`/apartments/documents/${documentId}`),
  getDashboardStats: () => api.get('/apartments/dashboard'),
  updateStatus: (id, status, comment) => api.put(`/admin/apartments/${id}/status`, { status, comment }),
  getMyApartments: (params) => api.get('/users/apartments/me', { params }),
  getOwnerStatistics: (params) => api.get('/apartments/owner/statistics', { params }),
  
  // Админские функции
  adminGetAllApartments: (params) => api.get('/admin/apartments', { params }),
  adminGetApartmentsStatistics: () => api.get('/admin/apartments/statistics'),
  adminGetApartmentById: (id) => api.get(`/admin/apartments/${id}`),
  adminDeleteApartment: (id) => api.delete(`/admin/apartments/${id}`),
  adminGetFullDashboardStats: () => api.get('/admin/dashboard/statistics'),
  adminGetDashboardStatistics: (params) => api.get('/admin/dashboard/statistics', { params }),
  adminGetApartmentBookingsHistory: (apartmentId, params) => api.get(`/admin/apartments/${apartmentId}/bookings-history`, { params }),
  
  // Админские функции для управления счетчиками
  adminUpdateCounters: (id, data) => api.put(`/admin/apartments/${id}/counters`, data),
  adminResetCounters: (id) => api.post(`/admin/apartments/${id}/counters/reset`),
  updateApartmentType: (id, apartmentTypeId) => api.put(`/admin/apartments/${id}/apartment-type`, { apartment_type_id: apartmentTypeId }),
};

// API функции для бронирований
export const bookingsAPI = {
  create: (data) => api.post('/bookings', data),
  createNegotiated: (data) => api.post('/bookings/negotiated', data),
  confirm: (id, data) => api.post(`/bookings/${id}/confirm`, data),
  getMyBookings: (params) => api.get('/users/my-bookings', { params }),
  getOwnerBookings: (params) => api.get('/users/property-bookings', { params }),
  getById: (id) => api.get(`/bookings/${id}`),
  getByNumber: (number) => api.get(`/bookings/number/${number}`),
  approve: (id, data) => api.post(`/bookings/${id}/approve`, data),
  reject: (id, data) => api.post(`/bookings/${id}/reject`, data),
  cancel: (id, data) => api.post(`/bookings/${id}/cancel`, data),
  finish: (id) => api.post(`/bookings/${id}/finish`),
  extend: (id, data) => api.post(`/bookings/${id}/extend`, data),
  getExtensions: (id) => api.get(`/bookings/${id}/extensions`),
  approveExtension: (id, extensionId) => api.post(`/bookings/${id}/extensions/${extensionId}/approve`),
  rejectExtension: (id, extensionId) => api.post(`/bookings/${id}/extensions/${extensionId}/reject`),
  checkAvailability: (apartmentId, params) => api.get(`/apartments/${apartmentId}/availability`, { params }),
  getAvailableSlots: (apartmentId, params) => api.get(`/apartments/${apartmentId}/available-slots`, { params }),
  getPaymentReceipt: (id) => api.get(`/bookings/${id}/receipt`),
  
  // Админские функции
  adminGetAllBookings: (params) => api.get('/admin/bookings', { params }),
  adminGetBookingsStatistics: () => api.get('/admin/bookings/statistics'),
  adminGetBookingById: (id) => api.get(`/admin/bookings/${id}`),
  adminUpdateBookingStatus: (id, status, reason) => 
    api.put(`/admin/bookings/${id}/status`, { status, reason }),
  adminCancelBooking: (id) => api.delete(`/admin/bookings/${id}`),
};

// API функции для замков
export const locksAPI = {
  create: (data) => api.post('/locks', data),
  getAll: () => api.get('/locks'),
  getById: (id) => api.get(`/locks/${id}`),
  getByUniqueId: (uniqueId) => api.get(`/locks/unique/${uniqueId}`),
  getByApartmentId: (apartmentId) => api.get(`/locks/apartment/${apartmentId}`),
  update: (id, data) => api.put(`/locks/${id}`, data),
  delete: (id) => api.delete(`/locks/${id}`),
  generatePassword: (uniqueId, data) => api.post(`/locks/password/${uniqueId}/generate`, data),
  getOwnerPassword: (uniqueId) => api.get(`/locks/password/${uniqueId}/owner`),
  getMyActivePasswords: () => api.get('/locks/password/my'),
  getPasswordForBooking: (bookingId) => api.get(`/locks/password/booking/${bookingId}`),
  deactivatePassword: (bookingId) => api.delete(`/locks/password/booking/${bookingId}`),
  getStatus: (uniqueId) => api.get(`/locks/status/${uniqueId}`),
  getHistory: (uniqueId) => api.get(`/locks/history/${uniqueId}`),
  
  // Функции auto-update
  getAutoUpdateStatus: (uniqueId) => api.get(`/locks/unique/${uniqueId}/auto-update/status`),
  enableAutoUpdate: (uniqueId) => api.post(`/locks/unique/${uniqueId}/auto-update/enable`),
  disableAutoUpdate: (uniqueId) => api.post(`/locks/unique/${uniqueId}/auto-update/disable`),
  syncWithTuya: (uniqueId) => api.post(`/locks/unique/${uniqueId}/sync`),
  configureWebhooks: (uniqueId) => api.post(`/locks/unique/${uniqueId}/configure-webhooks`),
  
  // Админские функции
  adminGetAllLocks: (params) => api.get('/admin/locks', { params }),
  adminGetLockById: (id) => api.get(`/admin/locks/${id}`),
  adminBindLockToApartment: (id, apartmentId) => 
    api.post(`/admin/locks/${id}/bind-apartment`, { apartment_id: apartmentId }),
  adminUnbindLockFromApartment: (id) => api.delete(`/admin/locks/${id}/unbind-apartment`),
  adminEmergencyResetLock: (id) => api.put(`/admin/locks/${id}/emergency-reset`),
  adminGetLocksStatistics: () => api.get('/admin/locks/statistics'),
  
  // Админские функции для управления паролями
  adminGeneratePassword: (uniqueId, data) => 
    api.post(`/admin/locks/by-unique-id/${uniqueId}/passwords`, data),
  adminGetAllLockPasswords: (uniqueId) => 
    api.get(`/admin/locks/by-unique-id/${uniqueId}/passwords`),
  adminDeactivatePassword: (passwordId) => 
    api.post(`/admin/passwords/${passwordId}/deactivate`),
};

// API функции для уведомлений
export const notificationsAPI = {
  getUnread: () => api.get('/notifications/unread'),
  getAll: (params) => api.get('/notifications', { params }),
  getCount: () => api.get('/notifications/count'),
  markAsRead: (id) => api.post(`/notifications/${id}/read`),
  markMultipleAsRead: (ids) => api.post('/notifications/read-multiple', { notification_ids: ids }),
  markAllAsRead: () => api.post('/notifications/read-all'),
  delete: (id) => api.delete(`/notifications/${id}`),
  deleteRead: () => api.delete('/notifications/read'),
  deleteAll: () => api.delete('/notifications/all'),
  registerDevice: (data) => api.post('/devices/register', data),
  updateDeviceHeartbeat: (token) => api.post('/devices/heartbeat', { device_token: token }),
  removeDevice: (token) => api.delete(`/devices/${token}`),
};

// API функции для чатов
export const chatAPI = {
  createRoom: (data) => api.post('/chat/rooms', data),
  getRooms: (params) => api.get('/chat/rooms', { params }),
  getRoom: (roomId) => api.get(`/chat/rooms/${roomId}`),
  sendMessage: (roomId, data) => api.post(`/chat/rooms/${roomId}/messages`, data),
  getMessages: (roomId, params) => api.get(`/chat/rooms/${roomId}/messages`, { params }),
  updateMessage: (messageId, data) => api.put(`/chat/messages/${messageId}`, data),
  deleteMessage: (messageId) => api.delete(`/chat/messages/${messageId}`),
  markAsRead: (roomId) => api.post(`/chat/rooms/${roomId}/read`),
  getUnreadCount: (roomId) => api.get(`/chat/rooms/${roomId}/unread-count`),
  activateChat: (roomId) => api.post(`/chat/rooms/${roomId}/activate`),
  closeChat: (roomId) => api.post(`/chat/rooms/${roomId}/close`),
};

// API функции для договоров
export const contractsAPI = {
  getContractHTML: (contractId) => api.get(`/contracts/${contractId}/html`),
  getByBookingId: (bookingId) => api.get(`/contracts/booking/${bookingId}`),
  getByApartmentId: (apartmentId) => api.get(`/contracts/apartment/${apartmentId}`),
};

// API функции для консьержей (админская панель)
export const conciergesAPI = {
  create: (data) => api.post('/admin/concierges', data),
  getAll: (params) => api.get('/admin/concierges', { params }),
  getById: (id) => api.get(`/admin/concierges/${id}`),
  update: (id, data) => api.put(`/admin/concierges/${id}`, data),
  delete: (id) => api.delete(`/admin/concierges/${id}`),
  getByApartment: (apartmentId) => api.get(`/admin/concierges/apartment/${apartmentId}`),
  assign: (data) => api.post('/admin/concierges/assign', data),
  removeFromApartment: (conciergeId, apartmentId) => api.delete(`/admin/concierges/${conciergeId}/apartments/${apartmentId}/remove`),
  getByOwner: (ownerId) => api.get(`/admin/concierges/owner/${ownerId}`),
  isUserConcierge: (userId) => api.get(`/admin/concierges/user/${userId}/status`),
};

// API функции для интерфейса консьержа
export const conciergeAPI = {
  // Получение профиля консьержа
  getProfile: () => api.get('/concierge/profile'),
  
  // Получение квартир консьержа
  getApartments: () => api.get('/concierge/apartments'),
  
  // Получение броней для квартир консьержа
  getBookings: (params) => api.get('/concierge/bookings', { params }),
  
  // Получение чат-комнат консьержа
  getChatRooms: (params) => api.get('/concierge/chat/rooms', { params }),
  
  // Получение статистики
  getStats: () => api.get('/concierge/stats'),
  
  // Обновление расписания
  updateSchedule: (data) => api.put('/concierge/schedule', data),
};

// API функции для уборщиц (админская панель)
export const cleanersAPI = {
  create: (data) => api.post('/admin/cleaners', data),
  getAll: (params) => api.get('/admin/cleaners', { params }),
  getById: (id) => api.get(`/admin/cleaners/${id}`),
  update: (id, data) => api.put(`/admin/cleaners/${id}`, data),
  delete: (id) => api.delete(`/admin/cleaners/${id}`),
  assignToApartment: (data) => api.post('/admin/cleaners/assign', data),
  removeFromApartment: (data) => api.post('/admin/cleaners/remove', data),
  updateSchedulePatch: (cleanerId, scheduleData) => api.patch(`/admin/cleaners/${cleanerId}/schedule`, scheduleData),
  updateApartments: async (cleanerId, apartmentIds) => {
    // Получаем список всех уборщиц (в нем есть квартиры)
    const cleanersResponse = await api.get('/admin/cleaners');
    
    // Находим нужную уборщицу в списке
    const cleaners = Array.isArray(cleanersResponse.data) ? cleanersResponse.data : 
                    cleanersResponse.data.data?.cleaners || 
                    cleanersResponse.data.data || [];
    const cleaner = cleaners.find(c => c.id === cleanerId);
    
    // Получаем текущие квартиры
    const currentApartmentIds = cleaner?.apartments?.map(apt => apt.id) || 
                               cleaner?.assigned_apartments?.map(apt => apt.id) || [];
    
    // Определяем, какие квартиры нужно удалить и добавить
    const toRemove = currentApartmentIds.filter(id => !apartmentIds.includes(id));
    const toAdd = apartmentIds.filter(id => !currentApartmentIds.includes(id));
    
    // Удаляем квартиры
    for (const apartmentId of toRemove) {
      await api.post('/admin/cleaners/remove', {
        cleaner_id: cleanerId,
        apartment_id: apartmentId
      });
    }
    
    // Добавляем квартиры
    for (const apartmentId of toAdd) {
      await api.post('/admin/cleaners/assign', {
        cleaner_id: cleanerId,
        apartment_id: apartmentId
      });
    }
    
    return { success: true };
  },
  getApartmentsNeedingCleaning: () => api.get('/admin/cleaners/apartments-needing-cleaning'),
};

// API функции для интерфейса уборщицы
export const cleanerAPI = {
  // Получение профиля уборщицы (с расписанием)
  getProfile: () => api.get('/cleaner/profile'),
  
  // Получение квартир уборщицы
  getApartments: () => api.get('/cleaner/apartments'),
  
  // Получение квартир для уборки
  getApartmentsForCleaning: () => api.get('/cleaner/apartments-for-cleaning'),
  
  // Получение статистики уборщицы
  getStats: () => api.get('/cleaner/stats'),
  
  // Начать уборку
  startCleaning: (data) => api.post('/cleaner/start-cleaning', data),
  
  // Завершить уборку
  completeCleaning: (data) => api.post('/cleaner/complete-cleaning', data),
  
  // Обновление расписания
  updateSchedule: (data) => api.put('/cleaner/schedule', data),
  
  // Частичное обновление расписания
  updateSchedulePatch: (data) => api.patch('/cleaner/schedule/patch', data),
};

// API функции для локаций
export const locationsAPI = {
  getRegions: () => api.get('/locations/regions'),
  getRegionById: (id) => api.get(`/locations/regions/${id}`),
  getCities: () => api.get('/locations/cities'),
  getCitiesByRegion: (regionId) => api.get(`/locations/regions/${regionId}/cities`),
  getCityById: (id) => api.get(`/locations/cities/${id}`),
  getDistricts: () => api.get('/locations/districts'),
  getDistrictsByCity: (cityId) => api.get(`/locations/cities/${cityId}/districts`),
  getDistrictById: (id) => api.get(`/locations/districts/${id}`),
  getMicrodistricts: () => api.get('/locations/microdistricts'),
  getMicrodistrictsByDistrict: (districtId) => api.get(`/locations/districts/${districtId}/microdistricts`),
  getMicrodistrictById: (id) => api.get(`/locations/microdistricts/${id}`),
};

// API функции для словарей
export const dictionariesAPI = {
  getConditions: () => api.get('/dictionaries/conditions'),
  getHouseRules: () => api.get('/dictionaries/house-rules'),
  getAmenities: () => api.get('/dictionaries/amenities'),
};

// API функции для избранного
export const favoritesAPI = {
  getAll: () => api.get('/favorites'),
  add: (apartmentId) => api.post('/favorites', { apartment_id: apartmentId }),
  remove: (apartmentId) => api.delete(`/favorites/${apartmentId}`),
};

// API функции для правил отмены
export const cancellationRulesAPI = {
  getAll: () => api.get('/cancellation-rules'),
  create: (data) => api.post('/cancellation-rules', data),
  getById: (id) => api.get(`/cancellation-rules/${id}`),
  update: (id, data) => api.put(`/cancellation-rules/${id}`, data),
  delete: (id) => api.delete(`/cancellation-rules/${id}`),
};

// API функции для настроек платформы
export const platformSettingsAPI = {
  getServiceFee: () => api.get('/settings/service-fee'),
  getAllSettings: () => api.get('/settings'),
  getSettingByKey: (key) => api.get(`/settings/${key}`),
  createSetting: (data) => api.post('/settings', data),
  updateSetting: (key, data) => api.put(`/settings/${key}`, data),
  deleteSetting: (key) => api.delete(`/settings/${key}`),
};

// API для управления типами квартир
export const apartmentTypesAPI = {
  // Публичные методы
  getAll: () => api.get('/apartment-types'),
  getById: (id) => api.get(`/apartment-types/${id}`),
  
  // Админские методы
  adminGetAll: () => api.get('/admin/apartment-types'),
  adminGetById: (id) => api.get(`/admin/apartment-types/${id}`),
  create: (data) => api.post('/admin/apartment-types', data),
  update: (id, data) => api.put(`/admin/apartment-types/${id}`, data),
  delete: (id) => api.delete(`/admin/apartment-types/${id}`),
};

export default api; 