import { Button } from "../components/Button";
import axiosInstance from "../AxiosInstance";
import { useEffect, useState } from "react";
import EditInputButton from "../components/EditInputButton";
import EditSelectButton from "../components/EditSelectButton";
import EditTagSelectButton from "../components/EditTagSelectorButton";
import "./UserProfile.css"

function profileResponseData(responseData){
    if(!responseData ||typeof responseData !== "object")
        throw new Error("Invalid profile data")
    return {
        userName: String(responseData.userName ?? ""),
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
        userName: "",
        firstName: "",
        lastName: "",
        email: "",
        gender: "",
        preferences: "",
        bio: "",
        interests: []
        });

    // const navigate = useNavigate();

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

    const handleSaveData = async (name, value) =>{
        if(["bio", "interests", "preferences", "gender"].includes(name)){
            try{
                const response = await axiosInstance.patch('/backend/api/accounts/bio/update',{
                    [name] : value
                })
                console.log("AXIOS BIO SAVED" , response.data);
            }
            catch(error){
                console.log("ERROR", error);
            }
        }else if(["username", "firstName", "lastName", "email"]){
            try{
                const response = await axiosInstance.patch('/backend/api/accounts/user/update',{
                    [name] : value
                })
                console.log("AXIOS USER SAVED" , response.data);
            }
            catch(error){
                console.log("ERROR", error);
            }
        }
    }

    const handleChange = (name, value) => {
    setUserProfileData(prev => ({
    ...prev,
    [name]: value
    }));
    };

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

    const handleAvatarClick = () => {
        console.log("avatar clicked");
    }

    return(
        <div id="user-profile-container">
            <div id="user-profile">
            <h2>User Profile</h2>
                <div className="avatar-wrapper">
                    <img
                        className="profile-avatar"
                        src="/1.png"
                        alt={`${userProfileData.userName || "User"} avatar`}
                    />
                        <button
                            type="button"
                            className="avatar-overlay-btn"
                            onClick={handleAvatarClick}>
                                Change Avatar
                    </button>
                </div>
                <div className="profile-grid">
                    <div className="field username">
                    <EditInputButton
                        type="text"
                        value={userProfileData.userName}
                        placeholder="Username"
                        // className="input-bio"
                        onChange={(e) => handleChange("userName",e.target.value)}
                        onSave={(savedValue) => {
                            console.log("Saved value:", savedValue);
                            handleSaveData("username", savedValue);
                        }}
                    />
                    </div>
                    <div className="field first-name">
                        <EditInputButton
                            type="text"
                            value={userProfileData.firstName}
                            placeholder="firstName"
                            // className="input-bio"
                            onChange={(e) => handleChange("firstName",e.target.value)}
                            onSave={(savedValue) => {
                                console.log("Saved value:", savedValue);
                                handleSaveData("firstName", savedValue);
                            }}
                        />
                    </div>
                    <div className="field last-name">
                        <EditInputButton
                            type="text"
                            value={userProfileData.lastName}
                            placeholder="lastName"
                            // className="input-bio"
                            onChange={(e) => handleChange("lastName",e.target.value)}
                            onSave={(savedValue) => {
                                console.log("Saved value:", savedValue);
                                handleSaveData("lastName", savedValue);
                            }}
                        />
                    </div>
                    <div className="field email">
                        <EditInputButton
                            type="text"
                            value={userProfileData.email}
                            placeholder="email"
                            // className="input-bio"
                            onChange={(e) => handleChange("email",e.target.value)}
                            onSave={(savedValue) => {
                                console.log("Saved value:", savedValue);
                                handleSaveData("email", savedValue);
                            }}
                        />
                    </div>
                    <div className="field gender">
                        <EditSelectButton
                            value={userProfileData.gender}
                            placeholder="Gender"
                            options={["Male", "Female", "Other"]}
                            onChange={(value) => handleChange("gender", value)}
                            onSave={(savedValue) => {
                                console.log("Saved value:", savedValue);
                                handleSaveData("gender", savedValue);
                            }}
                        />
                    </div>
                    <div className="field preferences">
                        <EditSelectButton
                            value={userProfileData.preferences}
                            placeholder="Preferences"
                            options={["Male", "Female", "Other"]}
                            onChange={(value) => handleChange("preferences", value)}
                            onSave={(savedValue) => {
                                console.log("Saved value:", savedValue);
                                handleSaveData("preferences", savedValue);
                            }}
                        />
                    </div>
                    <div className="field interests">
                        <EditTagSelectButton
                            value={userProfileData.interests}
                            tags={["#tag1", "#tag2", "#tag3"]}
                            onChange={(selectedTags) =>
                                handleChange("interests", selectedTags)}
                                onSave={(savedValue) => {
                                    console.log("Saved value:", savedValue);
                                handleSaveData("interests", savedValue);
                        }}
                        />
                    </div>
                    <div className="field bio">
                        <EditInputButton
                            type="text"
                            value={userProfileData.bio}
                            placeholder="BIO"
                            // className="input-bio"
                            onChange={(e) => handleChange("bio",e.target.value)}
                            onSave={(savedValue) => {
                                console.log("Saved value:", savedValue);
                                handleSaveData("bio", savedValue);
                            }}
                        />
                    </div>
                </div>
                <div className="profile-actions">
                    <Button onClick={handleLogout}>LOGOUT</Button>
                </div>
            </div>
        </div>
    );
}