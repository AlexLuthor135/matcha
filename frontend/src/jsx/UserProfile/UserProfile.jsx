import { Button } from "../components/Button";
import axiosInstance from "../AxiosInstance";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import UpdateUserProfile from "./UpdateUserProfile";

function profileResponseData(responseData){
    if(!responseData ||typeof responseData !== "object")
        throw new Error("Invalid profile data")
    return {
        username: String(responseData.username ?? ""),
        firstName: String(responseData.firstName ?? ""),
        lastName: String(responseData.lastName ?? ""),
        email: String(responseData.email ?? ""),
        // password: String(responseData.password ?? ""),
        gender: String(responseData.gender ?? ""),
        preferences: String(responseData.preferences ?? ""),
        bio: String(responseData.bio ?? ""),
        interests: Array.isArray(responseData.interests) ? responseData.interests : []
    }
}
export default function UserProfile(){

    const [userProfileData, setUserProfileData] = useState({
        username: "",
        firstName: "",
        lastName: "",
        email: "",
        gender: "",
        preferences: "",
        bio: "",
        interests: []
        });

    const navigate = useNavigate();

    useEffect(() => {
        const getUserProfile = async () => {
            try{
                const response = await axiosInstance.get('/backend/api/accounts/bio/get');
                console.log("USER response DATA : ",response.data)
                const userProfileData = profileResponseData(response.data)
                setUserProfileData(userProfileData);
                console.log("USER PROFILE  DATA : ",userProfileData)

            }
            catch(error){
                console.log("GET ERROR", error);
            }
        };
        getUserProfile();
    },[]);

    const handleUpdateUserProfile = (e) => {
        e.preventDefault();
        console.log("NAVIGATING TO UPDATEPROFILE => ");
        // navigate('/updateprofile');
    }

    const handleLogout= async (e) => {
        e.preventDefault();
        try{
            const response = await axiosInstance.post('/backend/api/accounts/logout/', {
            })
            console.log("LOGOUT COMLETED", response.status)
        }
        catch(error){
            console.log("LOGOUT ERROR", error);
        }
        finally{

        }
    }
    return(
        <div id="authorize-container">
            <h2>User Profile</h2>
            <Button>Change</Button>
            <p>User Name</p>
            <Button>Change</Button>
            <p>First Name</p>
            <Button>Change</Button>
            <p>Last Name</p>
            <Button>Change</Button>
            <p>email</p>
            <Button>Change</Button>
            <p>{userProfileData.gender}</p>
            <Button>Change</Button>
            <p>{userProfileData.preferences}</p>
            <Button>Change</Button>
            <p>{userProfileData.interests}</p>
            <Button>Change</Button>
            <p>{userProfileData.bio}</p>
            <Button onClick={handleUpdateUserProfile}>Change</Button>
            <Button onClick={handleLogout}>LOGOUT</Button>
        </div>
    );
}