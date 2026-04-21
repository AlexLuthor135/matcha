import { Button } from "../components/Button";
import "../UserProfile/UserProfile.css";

export default function DatingSlider() {
    
        return(
        <div id="user-profile-container">
            <div id="user-profile">
                <img
                    className="profile-avatar"
                    src="/1.png"
                    // alt={`${userProfileData.userName || "User"} avatar`}
                />
                <div className="profile-grid">
                    <div className="field username">
                        <p>Username</p>
                    </div>
                    <div className="field first-name">
                        <p>First name</p>
                    </div>
                    <div className="field last-name">
                        <p>Last name</p>
                    </div>
                    <div className="field email">
                        <p>Email</p>
                    </div>
                    <div className="field interests">
                        <p>Interests</p>
                    </div>
                    <div className="field bio">
                        <p>BIO</p>
                    </div>
                </div>
                <div className="profile-actions">
                    <Button >💔</Button>
                    <Button >❤️</Button>
                </div>

            </div>
        </div>
    );
}