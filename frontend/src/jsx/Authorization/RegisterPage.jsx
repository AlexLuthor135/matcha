import { useNavigate } from "react-router-dom";
import { Button } from "../components/Button";
import { InputBlock } from "../components/InputBlock";
import { useState } from "react";
import { passwordValidation } from "../components/Validation";
import "./Authorization.css";
import "../components/InputBlock.css"
import UserProfileEditor from "../UserProfile/UserProfileEditor";

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

    function handleLoginPage(){
        navigate('/');
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setUserData({ ...userData, [name]: value });
    };

    const handleRegistrationSubmit = async (e) =>{
        e.preventDefault();

        if(loading)
            return;

        setLoading(true);

        if(passwordValidation(userData.password, confirmPassword)){
            setUserData({...userData, password: ''})
            setConfirmPassword('');
            return;
        }


        // if(!username || !password || !confirmPassword || !email){
        //     alert("Not all fields are filled!");
        //     return;
        // }

        // setLoading(true);

        // try{
            // const response = await api.post('/registration', {
            //     username: username,
            //     lastName: lastName,
            //     firstName: firstName,
            //     password: password,
            //     email: email
            // });
        //     alert('SUCCESS')
        //     navigate('/');
        // } catch(err){
        //     alert("FAILER");
        // } finally {
        //     setLoading(false);
        // }
        // navigate('/');
        navigate("/testUpdate")
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