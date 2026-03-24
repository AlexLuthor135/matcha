import { useNavigate } from "react-router-dom";
import { Button } from "../components/Button";
import { InputBlock } from "../components/InputBlock";
import {useEffect, useState } from "react";
import { useAuth } from "../AuthProvider";
import axiosInstance from "../AxiosInstance";
import "./Authorization.css";

export default function LoginPage() {

  const [loginData, setLoginData] = useState({
      loginEmail: '',
      password: '',
  });
  const [loading, setLoading] = useState(false);
  const {isLoggedIn, setIsLoggedIn,isCompleted, setIsCompleted} = useAuth();

  const navigate = useNavigate();

  useEffect(() => {

      if (isLoggedIn && !isCompleted) {
      console.log("REGISTR IS NOT COMPLETED", isCompleted);
        navigate('/biocompletion');
      }else if(isLoggedIn && isCompleted){
        navigate("/userprofile");
        console.log("REGISTR IS COMPLETED", isCompleted);
      }
  }, [isLoggedIn, navigate]);

  const handleChange = (e) => {
      const { name, value } = e.target;
      setLoginData({ ...loginData, [name]: value });
  };

  function handleRegisterPage() {
    navigate('/registration');
  }
  const handleLoginSubmit = async (e) =>{
    e.preventDefault();
    if(loading)
      return;
    setLoading(true);
      
    try{
      const response = await axiosInstance.post('/backend/api/login', {
        email : loginData.loginEmail,
        password : loginData.password
      });
      console.log('LOGIN SUCCESS : ', isLoggedIn)
      console.log('LOGIN RESPONSE DATA : ', response)
      setIsLoggedIn(true);
      // if(response.status === 200)
      //   setIsCompleted(true)
    }
    catch(err){
      console.log(err)
      setLoading(false)
    }
    finally{
      setLoading(false)
    }
  }
  return (
    <form onSubmit={handleLoginSubmit}>
    <div id="authorize-container">
      <h2>Sign In</h2>
        <div id="authorize">
          <InputBlock type="email" 
                      placeholder="Email"
                      value={loginData.loginEmail}
                      onChange={handleChange}
                      name="loginEmail"
          />
          <InputBlock type="password" 
                      placeholder="Password"
                      value={loginData.password}
                      onChange={handleChange}
                      name="password"
          />
          <div className="buttons">
            <Button 
                      type="submit"
                      disabled={loading}
            >
                        {loading ? "Loading..." : "Accept"}
            </Button>
            <Button 
                      onClick={handleRegisterPage}
            >
                      Register
            </Button>
          </div>
          {/* <Button onClick={navigate("/testUpdate")}>Forgot password</Button> */}
          {/* FORGOT PASS BUTTON */}
        </div>
    </div>
    </form>
  );
}
