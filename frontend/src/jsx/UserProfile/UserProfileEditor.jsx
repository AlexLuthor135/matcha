import { InputBlock } from "../components/InputBlock";
import CustomSelect from "../components/CustomSelect";
import { useState } from "react";
import "../Authorization/Authorization.css";
import "../components/InputBlock.css"
import TagSelector from "../components/TagSelector";
export default function UserProfileEditor(){
    const [userBio, setUserBio] = useState({
        gender : '',
        bio : '',
        interests : []
    });
    const [loading, setLoading] = useState(false);

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
    }
    return (
        <form>
        <div id="authorize-container">
            <h2>Edit</h2>
            <div id="authorize">
                <div style={{ width: 300 }}>
                    <CustomSelect
                        placeholder="Gender"
                        options={["Male", "Female", "Other"]}
                        onChange={(value) => console.log(value)}
                    />
                </div>
                <div style={{ width: 300 }}>
                    <CustomSelect
                        placeholder="Sexual preferences"
                        options={["Male", "Female", "Other"]}
                        onChange={(value) => console.log(value)}
                    />
                </div>
                <InputBlock
                        type="text"
                        placeholder="BIO"
                        className="input-bio"
                />
                <TagSelector
                    tags={["#shit", "#more shit", "#govno"]}
                    onChange={(selectedTags) =>
                        handleChange("interests", selectedTags)}
                />
                
            </div>
        </div>
        </form>
    );
}