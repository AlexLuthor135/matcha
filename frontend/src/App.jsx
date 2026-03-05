import LoginPage from './jsx/Authorization/LoginPage';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import RegisterPage from './jsx/Authorization/RegisterPage';
import UserProfileEditor from './jsx/UserProfile/UserProfileEditor';

export default function App() {
  return (
   <BrowserRouter>  
      <main>
        <Routes>
          <Route path="/" element={<LoginPage />}/>
          <Route path="/registration" element={<RegisterPage />}/>
          <Route path="/testUpdate" element={<UserProfileEditor/>}/>
        </Routes>
      </main>
    </BrowserRouter>
  );
}
