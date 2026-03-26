import { InputBlock } from "../components/InputBlock";
import CustomSelect from "../components/CustomSelect";
import { useState } from "react";
import "../Authorization/Authorization.css";
import "../components/InputBlock.css"
import TagSelector from "../components/TagSelector";
import { Button } from "../components/Button";
import axiosInstance from "../AxiosInstance";
import { useNavigate } from "react-router-dom";

export default function UpdateUserProfile(){
const [userBio, setUserBio] = useState({
        gender : '',
        preferences: '',
        bio : '',
        interests : [],
    });
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();
    // const {isCompleted, setIsCompleted} = useAuth();

    const handleChange = (name, value) => {
        setUserBio(prev => ({
        ...prev,
        [name]: value
        }));
    };

    const  handleBioSubmit = async (e) => {
        e.preventDefault();
        if(loading)
            return;

        setLoading(true);
        console.log("Updated USER BIO : ", userBio)
        try{
            const response = await axiosInstance.patch('/backend/api/accounts/bio/update', {
                gender: userBio.gender,
                preferences: userBio.preferences,
                bio: userBio.bio,
                interests: userBio.interests,
                isCompleted: true
            })
        console.log("response status",response.status)
        if(response.status === 200){
            console.log("Naivgating User Profile")
            navigate('/userprofile');
        }

        }
        catch(err){
            console.log("UPDATING BIO ERROR : ", err)
            setLoading(false);
        }
        finally{
            setLoading(false);
        }
    }

    return(
                <form onSubmit={handleBioSubmit}>
                <div id="authorize-container">
                    <h2>Edit</h2>
                    <div id="authorize">
                        <div style={{ width: 300 }}>
                            <CustomSelect
                                placeholder="Gender"
                                options={["Male", "Female", "Other"]}
                                onChange={(value) => handleChange("gender", value)}
                            />
                        </div>
                        <div style={{ width: 300 }}>
                            <CustomSelect
                                placeholder="Sexual preferences"
                                options={["Male", "Female", "Other"]}
                                onChange={(value) => handleChange("preferences", value)}
                            />
                        </div>
                        <InputBlock
                                type="text"
                                value={userBio.bio}
                                placeholder="BIO"
                                className="input-bio"
                                onChange={(e) => handleChange("bio",e.target.value)}
                        />
                        <TagSelector
                            tags={["#tag1", "#tag2", "#tag3"]}
                            onChange={(selectedTags) =>
                                handleChange("interests", selectedTags)}
                            name="interests"
                        />
                        <Button type="submit">Submit</Button>
                        
                    </div>
                </div>
                </form>
    )
}