import { useNavigate } from "react-router-dom";
import { Button } from "../components/Button";
import { InputBlock } from "../components/InputBlock";
import { useState } from "react";
import "./Authorization.css";

export default function LoginPage() {

  const [loginData, setLoginData] = useState({
      loginEmail: '',
      password: '',
  });
  const [loading, setLoading] = useState(false);

  const navigate = useNavigate();

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
      
//AXIOS LOGIN FORM 
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
          <Button onClick={navigate("/testUpdate")}>Forgot password</Button>
          {/* FORGOT PASS BUTTON */}
        </div>
    </div>
    </form>
  );
}
