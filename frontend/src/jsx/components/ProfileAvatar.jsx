/* eslint-disable react/prop-types */
import "./ProfileAvatar.css";

export default function ProfileAvatar({ avatarUrl, userName, onChange }) {
    return (
        <div className="profile-avatar-wrapper">
            <img
                className="profile-avatar"
                src={avatarUrl ? "/backend" + avatarUrl : ""}
                alt={`${userName || "User"} avatar`}
            />
            <label className="avatar-upload-control">
                Change
                <input
                    type="file"
                    accept="image/*"
                    onChange={onChange}
                />
            </label>
        </div>
    );
}
