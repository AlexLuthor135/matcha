import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthProvider';
import axiosInstance from './AxiosInstance';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

export default function LoginCallback() {
  const navigate = useNavigate();
  const { setIsLoggedIn, setStaff } = useAuth();

  useEffect(() => {
    const verifyLogin = async () => {
      try {
        const response = await axiosInstance.get('/backend/api/accounts/verify_login/');
        const isStaff = response.data?.is_staff ?? false;
        setIsLoggedIn(true);
        setStaff(isStaff);
        if (isStaff) {
          navigate('/admins/dashboard');
        } else {
          navigate('/profile');
        }
      } catch (error) {
        const errorMessage = error.response?.data?.error || "Login verification failed. Please try again.";
        toast.error(errorMessage);
        navigate('/');
      }
    };
    verifyLogin();
  }, [navigate, setIsLoggedIn]);

  return <div>Verifying tokens, please wait...</div>;
}
