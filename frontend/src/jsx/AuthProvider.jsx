import { useContext, createContext, useState, useEffect } from "react";
import axiosInstance from "./AxiosInstance";

const AuthContext = createContext();

const AuthProvider = ({ children }) => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [userID, setUserID] = useState(null);
  const [isCompleted, setIsCompleted] = useState(false)

  const verifyLogin = async () => {
    try {
      const response = await axiosInstance.get('/backend/api/accounts/verify_login/');
      setIsLoggedIn(true);
      setUserID(response.data.id);
      setIsCompleted(response.data.is_completed);
    } catch (error) {
      setIsLoggedIn(false);
      setUserID(null);
    } finally {
      setIsLoading(false);
    }
  };  

  useEffect(() => {
    verifyLogin();
  }, []);

  return (
    <AuthContext.Provider value={{ isLoggedIn, setIsLoggedIn,isCompleted,setIsCompleted, isLoading, userID, setUserID }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthProvider;
export const useAuth = () => useContext(AuthContext);
