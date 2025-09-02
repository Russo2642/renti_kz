import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'

// Обработчик ошибок для игнорирования ошибок от расширений браузера
window.addEventListener('error', (event) => {
  // Игнорируем ошибки от расширений браузера
  if (event.message && event.message.includes('Could not establish connection')) {
    event.preventDefault();
    return false;
  }
  if (event.filename && event.filename.includes('chrome-extension://')) {
    event.preventDefault();
    return false;
  }
  if (event.filename && event.filename.includes('polyfill.js')) {
    event.preventDefault();
    return false;
  }
});

// Обработчик для unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
  // Игнорируем ошибки от расширений браузера
  if (event.reason && typeof event.reason === 'object' && event.reason.message) {
    if (event.reason.message.includes('Could not establish connection')) {
      event.preventDefault();
      return false;
    }
    if (event.reason.message.includes('Receiving end does not exist')) {
      event.preventDefault();
      return false;
    }
  }
});

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
