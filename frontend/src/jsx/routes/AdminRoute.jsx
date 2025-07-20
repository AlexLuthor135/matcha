import React from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../AuthProvider";
import LoadingScreen from "../LoadingScreen";
import LogoutButton from "../LogoutButton";

const AdminRoute = () => {
  const { isLoggedIn, isLoading, staff } = useAuth();
  if (isLoading) {
    return <LoadingScreen message="Loading authentication..." />;
  }

  if (!isLoggedIn || !staff) {
    return <Navigate to="/" replace />;
  }

  return (
    <>
      <LogoutButton />
      <Outlet />
    </>
  );
};

export default AdminRoute;
