import React, { useState } from 'react';
import { toast } from 'react-toastify';
import axiosInstance from './AxiosInstance';
import '../css/index.css';

export default function LoginForm() {
  const [formData, setFormData] = useState({ username: '', password: '' });
  const [loading, setLoading] = useState(false);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      await axiosInstance.post('/backend/api/login/', formData);
      toast.success('Login successful!', { position: 'top-center' });
      window.location.href = '/';
    } catch (error) {
      const errorMessage =
        (error.response && error.response.data && error.response.data.error) ||
        'Login failed. Please try again.';
      toast.error(errorMessage, { position: 'top-center' });
      console.error('Login failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-form-container">
      <h2>Login</h2>
      <form onSubmit={handleSubmit} className="form">
        <div className="form-group">
          <label htmlFor="username">Username:</label>
          <input
            type="text"
            id="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            disabled={loading}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="password">Password:</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            disabled={loading}
            required
          />
        </div>
        <button type="submit" className="button" disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>
    </div>
  );
}