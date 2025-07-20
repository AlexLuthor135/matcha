import React from "react";
import { Link } from "react-router-dom";

const AdminDashboard = () => {
  return (
    <div className="flex flex-col justify-center items-center min-h-screen p-6 text-center">
        <h1 className="text-3xl font-bold mb-8 text-white">Admin Dashboard</h1>
            <div className="flex justify-center">
                <Link to="/admins/cluster" className="bg-white border border-gray-300 shadow-lg rounded-xl p-6 w-64 h-40 flex flex-col items-center justify-center hover:shadow-xl hover:bg-blue-50 transition duration-200">
                    <div className="text-5xl mb-3 text-blue-500">🗺️</div>
                    <h2 className="text-xl font-semibold text-blue-800">Cluster Map</h2>
                </Link>
                <Link to="/admins/commands" className="bg-white border border-gray-300 shadow-lg rounded-xl p-6 w-64 h-40 flex flex-col items-center justify-center hover:shadow-xl hover:bg-blue-50 transition duration-200">
                    <div className="text-5xl mb-3 text-blue-500">🛠️</div>
                    <h2 className="text-xl font-semibold text-blue-800">Commands</h2>
                </Link>
            </div>
    </div>
  );
};

export default AdminDashboard;
