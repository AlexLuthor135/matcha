import React from 'react';
import '../css/index.css';

export default function LoginButton({ disabled }) {
  function handleLogin() {
    if (disabled) return;
    window.location.href = '/backend/api/accounts/42/login/';
  }

  return (
    <button 
      onClick={handleLogin}
      className="button"
      disabled={disabled}
    >
      Login with 42
    </button>
  );
}
