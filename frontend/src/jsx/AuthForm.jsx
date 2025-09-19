import React, { useState } from 'react';
import LoginForm from './LoginForm';
import RegisterForm from './RegisterForm';
import '../css/Auth.css';

export default function AuthForm() {
  const [isLogin, setIsLogin] = useState(true);

  const toggleForm = () => {
    setIsLogin(!isLogin);
  };

  return (
    <div className="auth-container">
      <div className="form-toggle-container">
        <button 
          onClick={toggleForm} 
          className={`toggle-button ${isLogin ? 'active' : ''}`}
        >
          Login
        </button>
        <button 
          onClick={toggleForm} 
          className={`toggle-button ${!isLogin ? 'active' : ''}`}
        >
          Register
        </button>
      </div>

      <div className="form-content">
        {isLogin ? <LoginForm /> : <RegisterForm />}
      </div>
    </div>
  );
}