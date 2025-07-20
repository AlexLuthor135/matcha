import React from "react";
import "../css/LoadingScreen.css";

const LoadingScreen = ({ message = "Loading..." }) => {
  return (
    <div className="loading-screen">
      <div className="spinner"></div>
      <div className="loading-message">{message}</div>
    </div>
  );
};

export default LoadingScreen;
