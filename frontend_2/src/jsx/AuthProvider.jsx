import { useContext, createContext, useState, useEffect } from "react";
import axiosInstance from "./AxiosInstance";

const AuthContext = createContext();

const AuthProvider = ({ children }) => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [staff, setStaff] = useState(null);

  const verifyLogin = async () => {
    try {
      const response = await axiosInstance.get('/backend/api/accounts/verify_login/');
      setIsLoggedIn(true);
      setStaff(response.data.is_staff);
    } catch (error) {
      setIsLoggedIn(false);
    } finally {
      setIsLoading(false);
    }
  };  

  useEffect(() => {
    verifyLogin();
  }, []);

  return (
    <AuthContext.Provider value={{ isLoggedIn, setIsLoggedIn, isLoading, staff, setStaff }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthProvider;
export const useAuth = () => useContext(AuthContext);
