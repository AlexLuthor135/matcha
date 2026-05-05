import { Button } from "../components/Button";
import axiosInstance from "../AxiosInstance";
import { useEffect, useState } from "react";
import EditInputButton from "../components/EditInputButton";
import EditSelectButton from "../components/EditSelectButton";
import EditTagSelectButton from "../components/EditTagSelectorButton";
import ProfileAvatar from "../components/ProfileAvatar";
import ProfilePhotos from "../components/ProfilePhotos";
import "./UserProfile.css"

function profileResponseData(responseData){
    if(!responseData ||typeof responseData !== "object")
        throw new Error("Invalid profile data")
    return {
        userName: String(responseData.userName ?? ""),
        firstName: String(responseData.firstName ?? ""),
        lastName: String(responseData.lastName ?? ""),
        email: String(responseData.email ?? ""),
        avatar: String(responseData.avatar ?? ""),
        photos: Array.isArray(responseData.photos) ? responseData.photos : [],
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
        avatar: "",
        photos: [],
        gender: "",
        preferences: "",
        bio: "",
        interests: []
        });
    const [deletingPhotoId, setDeletingPhotoId] = useState(null);

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
        }else if(["username", "firstName", "lastName", "email"].includes(name)){
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
    }

    const uploadAvatar = async (e) => {
        const file = e.target.files?.[0];
        if (!file) {
            console.warn("No file selected");
            return;
        }

        const formData = new FormData();
        formData.append("avatar", file);

        try {
            const response = await axiosInstance.post(
                "/backend/api/accounts/avatar/upload",
                formData
            );

            console.log("Upload success:", response.data.avatar_url);
            setUserProfileData(prev => ({
                ...prev,
                avatar: response.data.avatar_url ?? ""
                }));
            e.target.value = '';

        } catch (error) {
            console.error("Error message:", error.message);
            e.target.value = '';
        }
    };

    const uploadPhotos = async (e) => {
        const files = Array.from(e.target.files ?? []);
        if (files.length === 0) {
            console.warn("No file selected");
            return;
        }

        const remainingSlots = 5 - userProfileData.photos.length;
        if (remainingSlots <= 0) {
            console.warn("You can upload up to 5 photos");
            e.target.value = '';
            return;
        }
        if (files.length > remainingSlots) {
            console.warn(`You can upload only ${remainingSlots} more photo(s)`);
            e.target.value = '';
            return;
        }

        const formData = new FormData();
        files.forEach((file) => {
            formData.append("photos", file);
        });

        try {
            const response = await axiosInstance.post(
                "/backend/api/accounts/photo/upload",
                formData
            );

            console.log("Upload success:", response.data.photos);
            setUserProfileData(prev => ({
                ...prev,
                photos: [
                    ...(prev.photos ?? []),
                    ...(Array.isArray(response.data.photos) ? response.data.photos : [])
                ].slice(0, 5)
                }));
            e.target.value = '';

        } catch (error) {
            console.error("Error message:", error.message);
            e.target.value = '';
        }
    };

    const deletePhoto = async (photo) => {
        if (!photo?.id || deletingPhotoId) {
            return;
        }

        setDeletingPhotoId(photo.id);
        try {
            await axiosInstance.delete(`/backend/api/accounts/photo/${photo.id}`);
            setUserProfileData(prev => ({
                ...prev,
                photos: (prev.photos ?? []).filter(currentPhoto => currentPhoto.id !== photo.id)
            }));
        } catch (error) {
            console.error("Delete photo error:", error.message);
        } finally {
            setDeletingPhotoId(null);
        }
    };

    const fullName = [userProfileData.firstName, userProfileData.lastName].filter(Boolean).join(" ");

    return(
        <div id="user-profile-container">
            <div id="user-profile">
                <div className="profile-header">
                    <ProfileAvatar
                        avatarUrl={userProfileData.avatar}
                        userName={userProfileData.userName}
                        onChange={uploadAvatar}
                    />
                    <div className="profile-heading">
                        <p className="profile-kicker">User Profile</p>
                        <h2>{fullName || userProfileData.userName || "Your profile"}</h2>
                        <p className="profile-subtitle">{userProfileData.email || "Add your email"}</p>
                    </div>
                </div>
                <ProfilePhotos
                    photos={userProfileData.photos}
                    deletingPhotoId={deletingPhotoId}
                    onUpload={uploadPhotos}
                    onDelete={deletePhoto}
                />
                <div className="profile-grid">
                    <div className="field username">
                    <span className="field-label">Username</span>
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
                        <span className="field-label">First name</span>
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
                        <span className="field-label">Last name</span>
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
                        <span className="field-label">Email</span>
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
                        <span className="field-label">Gender</span>
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
                        <span className="field-label">Preferences</span>
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
                        <span className="field-label">Interests</span>
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
                        <span className="field-label">Bio</span>
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
