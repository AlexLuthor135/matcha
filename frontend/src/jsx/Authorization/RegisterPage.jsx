import { useNavigate } from "react-router-dom";
import { Button } from "../components/Button";
import { InputBlock } from "../components/InputBlock";
import { useEffect, useState } from "react";
import { passwordValidation } from "../components/Validation";
import axiosInstance from "../AxiosInstance";
import "./Authorization.css";
import "../components/InputBlock.css"

export default function RegisterPage(){

    const [userData, setUserData] = useState({
        username: '',
        password: '',
        email: '',
        lastName: '',
        firstName: '',
    });

    const navigate = useNavigate();
    const [confirmPassword, setConfirmPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [registered, setRegistered] = useState(false);
    function handleLoginPage(){
        navigate('/');
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setUserData({ ...userData, [name]: value });
    };

    useEffect(() => {
        if (registered) {
            navigate('/');
        }
    }, [registered, navigate]);
    
    const handleRegistrationSubmit = async (e) =>{
        e.preventDefault();

        if(loading)
            return;

        if(passwordValidation(userData.password, confirmPassword)){
            setUserData({...userData, password: ''})
            setConfirmPassword('');
            return;
        }
        if(!userData.username || !userData.password || !confirmPassword || !userData.email){
            alert("Not all fields are filled!");
            return;
        }

        setLoading(true);


        try{
            const response = await axiosInstance.post('/backend/api/register', {
                username: userData.username,
                last_name: userData.lastName,
                first_name: userData.firstName,
                password: userData.password,
                email: userData.email
            });

            console.log('REGISTR SUCCESS : ', response)
            // navigate('/');
            setRegistered(true);
        } catch(err){
            alert("REGISTR FAILER");
            console.log(err);
            setLoading(false);
        } finally {
            setLoading(false);
        }
        // navigate('/');
        // navigate("/testUpdate")
    }

    return(
        <form onSubmit={handleRegistrationSubmit}>
        <div id="authorize-container">
            <h2>Registration</h2>
            <div id="authorize" className="register">
                <InputBlock 
                            type="text"
                            value={userData.username}
                            onChange={handleChange}
                            placeholder="Username"
                            name="username"
                />
                <div className="input-row">
                <InputBlock 
                            type="text"
                            value={userData.lastName}
                            onChange={handleChange}
                            placeholder="Last name"
                            name="lastName"
                />
                <InputBlock 
                            type="text"
                            value={userData.firstName}
                            onChange={handleChange}
                            placeholder="First name"
                            name="firstName"
                />
                </div>
                <div className="input-row">
                <InputBlock 
                            type="password"
                            value={userData.password}
                            onChange={handleChange}
                            placeholder="Password"
                            name="password"
                />

                <InputBlock 
                            type="password"
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            placeholder="Confirm password"
                />
                </div>
                <InputBlock 
                            type="email"
                            value={userData.email}
                            onChange={handleChange}
                            placeholder="Email"
                            name="email"
                />
                <div className="buttons">
                    <Button type="submit" 
                            disabled={loading}
                    >
                        {loading ? "Loading..." : "Register"}
                    </Button>
                    <Button 
                    type="button"
                    onClick={handleLoginPage}
                    >
                        Cancel
                    </Button>
                </div>
            </div>
        </div>
        </form>
    );
}