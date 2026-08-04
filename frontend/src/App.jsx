import LoginPage from './jsx/Authorization/LoginPage';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import RegisterPage from './jsx/Authorization/RegisterPage';
import ProfileCompletion from './jsx/UserProfile/ProfileCompletion';
import PrivateRoute from './jsx/routes/PrivateRoute';
import UserProfile from './jsx/UserProfile/UserProfile';
import UpdateUserProfile from './jsx/UserProfile/UpdateUserProfile';
import NavigationBar from './jsx/NavigationBar/NavigationBar';
import DatingSlider from './jsx/DatingSlider/DatingSlider';
import Chat from './jsx/Chat/Chat';

export default function App() {
  return (
   <BrowserRouter>
      <main>
            <NavigationBar/>
        <Routes>
          <Route path="/" element={<LoginPage />}/>
          <Route path="/registration" element={<RegisterPage />}/>
          <Route element={<PrivateRoute/>}>
            <Route path="/profile-completion" element={<ProfileCompletion/>}/>
            <Route path="/userprofile" element={<UserProfile/>}/>
            <Route path='/updateprofile' element={<UpdateUserProfile/>}/>
            <Route path='/datingslider' element={<DatingSlider/>}/>
            <Route path='/chat' element={<Chat/>}/>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />}/>
        </Routes>
      </main>
    </BrowserRouter>
  );
}
