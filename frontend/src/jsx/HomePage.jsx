import React from 'react';
import AuthForm from './AuthForm';
import SplashText from './SplashText';

export default function LoginPage() {
  return (
    <div className="login-page">
      <SplashText />
      <AuthForm />
    </div>
  );
}
