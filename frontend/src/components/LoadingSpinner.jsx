import React from 'react';
import { Spin } from 'antd';

const LoadingSpinner = ({ size = 'large', text = 'Загрузка...' }) => {
  return (
    <div className="fixed inset-0 flex items-center justify-center bg-white z-50">
      <div className="text-center">
        <Spin size={size} />
        {text && <div className="mt-4 text-gray-600">{text}</div>}
      </div>
    </div>
  );
};

export default LoadingSpinner; 