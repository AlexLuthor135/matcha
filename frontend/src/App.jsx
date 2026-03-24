import LoginPage from './jsx/Authorization/LoginPage';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import RegisterPage from './jsx/Authorization/RegisterPage';
import BioСompletion from './jsx/UserProfile/BioСompletion';
import PrivateRoute from './jsx/routes/PrivateRoute';
import UserProfile from './jsx/UserProfile/UserProfile';

export default function App() {
  return (
   <BrowserRouter>
      <main>
        <Routes>
          <Route path="/" element={<LoginPage />}/>
          <Route path="/registration" element={<RegisterPage />}/>
          <Route element={<PrivateRoute/>}>
            <Route path="/biocompletion" element={<BioСompletion/>}/>
            <Route path="/userprofile" element={<UserProfile/>}/>
          </Route>
        </Routes>
      </main>
    </BrowserRouter>
  );
}
