import { Navigate } from "react-router-dom";
import { useAuth } from "../AuthProvider";
import LoadingScreen from "../components/LoadingScreen";

const PublicRoute = ({ children }) => {
  const { isLoggedIn, isLoading } = useAuth();
  if (isLoading) return <LoadingScreen message="Authentication..." />;

  if (isLoggedIn) {
    return <Navigate to="/userprofile" replace />;
  }

  return children;
};

export default PublicRoute;
