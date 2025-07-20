import { BrowserRouter as Router, Route, Routes, Navigate } from 'react-router-dom';
import PublicRoute from './jsx/routes/PublicRoute';
import PrivateRoute from './jsx/routes/PrivateRoute';
import AdminRoute from './jsx/routes/AdminRoute';
import HomePage from './jsx/HomePage';
import LoginCallback from './jsx/LoginCallback';
import ProfilePage from './jsx/ProfilePage';
import Stars from './jsx/Stars';
import AdminDashboard from './jsx/AdminDashboard';
import Static from './jsx/Static';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import UserList from './jsx/UserList';
import AdminCommands from './jsx/AdminCommands';

export default function App() {
  return (
    <div>
      <Stars />
      <Router>
        <Routes>
        <Route path="/" element={<PublicRoute> <HomePage /></PublicRoute>} />
          <Route path="/login/callback" element={<LoginCallback />} />
          <Route element={<PrivateRoute />}>
            <Route path="/profile" element={<ProfilePage />} />
          </Route>
          <Route element={<AdminRoute />}>
            <Route path="/admins/dashboard" element={<AdminDashboard />} />
            <Route path="/admins/cluster" element={<UserList />} />
            <Route path="/admins/commands" element={<AdminCommands />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
      <ToastContainer />
      <Static />
    </div>
  );
}
