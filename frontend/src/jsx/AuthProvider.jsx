import { useCallback, useContext, createContext, useState, useEffect } from "react";
import axiosInstance from "./AxiosInstance";

const AuthContext = createContext();

const AuthProvider = ({ children }) => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [userID, setUserID] = useState(null);
  const [isCompleted, setIsCompleted] = useState(false)

  const verifyLogin = useCallback(async () => {
    try {
      const response = await axiosInstance.get('/backend/api/accounts/verify_login');
      setIsLoggedIn(true);
      setUserID(response.data.id);
      setIsCompleted(response.data.is_completed);
      return response.data;
    } catch {
      setIsLoggedIn(false);
      setUserID(null);
      setIsCompleted(false);
      return null;
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    verifyLogin();
  }, [verifyLogin]);

  return (
    <AuthContext.Provider value={{
      isLoggedIn,
      setIsLoggedIn,
      isCompleted,
      setIsCompleted,
      isLoading,
      userID,
      setUserID,
      verifyLogin,
    }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthProvider;
export const useAuth = () => useContext(AuthContext);
