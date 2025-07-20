import React from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../AuthProvider";
import LoadingScreen from "../LoadingScreen";
import LogoutButton from "../LogoutButton";

const PrivateRoute = () => {
  const { isLoggedIn, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingScreen message="Authentication..." />;
  }

  if (!isLoggedIn) {
    return <Navigate to="/" replace />;
  }

  return (
    <>
      <LogoutButton />
      <Outlet />
    </>
  );
};


export default PrivateRoute;