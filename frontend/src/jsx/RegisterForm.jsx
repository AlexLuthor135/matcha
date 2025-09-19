import React, { useState } from 'react';
import { toast } from 'react-toastify';
import axiosInstance from './AxiosInstance';
import '../css/index.css';

export default function RegisterForm() {
  const [formData, setFormData] = useState({
    email: '',
    login: '',
    first_name: '',
    last_name: '',
    password: '',
  });
  const [loading, setLoading] = useState(false);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      // Use axiosInstance to send a POST request to your registration endpoint.
      // Make sure the endpoint path is correct, e.g., '/backend/api/register/'.
      await axiosInstance.post('/backend/api/register/', formData);
      toast.success('Registration successful! Please check your email for a verification link.', { position: 'top-center' });
    } catch (error) {
      const errorMessage =
        (error.response && error.response.data && error.response.data.error) ||
        'Registration failed. Please try again.';
      toast.error(errorMessage, { position: 'top-center' });
      console.error('Registration failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="registration-form-container">
      <h2>Register</h2>
      <form onSubmit={handleSubmit} className="form">
        <div className="form-group">
          <label htmlFor="email">Email:</label>
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            disabled={loading}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="login">Username:</label>
          <input
            type="text"
            id="login"
            name="login"
            value={formData.login}
            onChange={handleChange}
            disabled={loading}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="first_name">First Name:</label>
          <input
            type="text"
            id="first_name"
            name="first_name"
            value={formData.first_name}
            onChange={handleChange}
            disabled={loading}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="last_name">Last Name:</label>
          <input
            type="text"
            id="last_name"
            name="last_name"
            value={formData.last_name}
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
          {loading ? 'Registering...' : 'Register'}
        </button>
      </form>
    </div>
  );
}