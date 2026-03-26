import React from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../AuthProvider";
import LoadingScreen from "../components/LoadingScreen";
const PrivateRoute = () => {
  const { isLoggedIn, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingScreen message="Authentication..." />;
  }

  if (!isLoggedIn) {
    return <Navigate to="/" replace />;
  }
  console.log("Logged IN ALREADY")
  return (
    <>
      <Outlet />
    </>
  );
};


export default PrivateRoute;