import { useState } from "react";
import { InputBlock } from "./InputBlock";
import { Button } from "./Button";

export default function EditInputButton(props) {
    const [isEditing, setIsEditing] = useState(false);
    const { userProfileData, onChange, onSave, value, ...inputProps } = props;

    const handleSave = () => {
        if (typeof onSave === "function") {
            onSave(value);
        }
        setIsEditing(false);
    };

    return (
        <div className="profile-field-control">
            {isEditing ? (
                <>
                    <InputBlock {...inputProps} value={value} onChange={onChange} />
                    <Button onClick={handleSave}>Save</Button>
                </>
            ) : (
                <>
                    <p>{value}</p>
                    <Button
                        onClick={() => setIsEditing(true)}
                        iconSrc="/edit_button.png"
                        ariaLabel="Edit"
                        iconAlt="Edit"
                    />
                </>
            )}
        </div>
    );
}