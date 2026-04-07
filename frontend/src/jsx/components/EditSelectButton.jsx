import { useState } from "react";
import { Button } from "./Button";
import CustomSelect from "./CustomSelect";

export default function EditSelectButton(props) {
    const [isEditing, setIsEditing] = useState(false);
    const {onChange, value, onSave, ...inputProps } = props;

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
                    <CustomSelect {...inputProps} onChange={onChange} />
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