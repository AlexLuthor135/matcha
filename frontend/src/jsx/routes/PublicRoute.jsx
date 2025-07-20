import React from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../AuthProvider";
import LoadingScreen from "../LoadingScreen";

const PublicRoute = ({ children }) => {
  const { isLoggedIn, isLoading, staff } = useAuth();
  if (isLoading) return <LoadingScreen message="Authentication..." />;

  if (isLoggedIn && staff) {
    return <Navigate to="/admins/dashboard" replace />;
  }

  if (isLoggedIn) {
    return <Navigate to="/profile" replace />;
  }

  return children;
};

export default PublicRoute;
