import React from 'react';
import axiosInstance from './AxiosInstance';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import '../css/LogoutButton.css';
import "../css/index.css"

const handleLogout = async () => {
    try {
      await axiosInstance.post('backend/api/accounts/42/logout/');
      toast.success("Logged out.");
      window.location.href = '/';
    } catch (error) {
      toast.error("Logout failed.");
    }
  };
  

export default function LogoutButton() {
return (
    <div className="logout-button-container">
        <button className="button" onClick={handleLogout}>
            Logout
        </button>
    </div>
);
}
