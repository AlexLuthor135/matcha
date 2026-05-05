import { useEffect, useMemo, useState } from "react";
import axiosInstance from "../AxiosInstance";
import { Button } from "../components/Button";
import "../UserProfile/UserProfile.css";
import "./DatingSlider.css";

function normalizeProfile(profile) {
    if (!profile || typeof profile !== "object") {
        return null;
    }

    return {
        id: profile.id,
        userName: String(profile.userName ?? ""),
        firstName: String(profile.firstName ?? ""),
        lastName: String(profile.lastName ?? ""),
        avatar: String(profile.avatar ?? ""),
        photos: Array.isArray(profile.photos) ? profile.photos : [],
        bio: String(profile.bio ?? ""),
    };
}

function getPhotoUrl(url) {
    return url ? "/backend" + url : "";
}

export default function DatingSlider() {
    const [profiles, setProfiles] = useState([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState("");

    useEffect(() => {
        const getProfiles = async () => {
            try {
                const response = await axiosInstance.get("/backend/api/accounts/profiles/feed");
                const profileList = Array.isArray(response.data?.profiles)
                    ? response.data.profiles.map(normalizeProfile).filter(Boolean)
                    : [];

                setProfiles(profileList);
                setCurrentIndex(0);
            } catch (requestError) {
                console.log("PROFILE FEED ERROR", requestError);
                setError("Could not load profiles");
            } finally {
                setIsLoading(false);
            }
        };

        getProfiles();
    }, []);

    const currentProfile = profiles[currentIndex] ?? null;
    const fullName = currentProfile
        ? [currentProfile.firstName, currentProfile.lastName].filter(Boolean).join(" ")
        : "";
    const photos = useMemo(() => {
        if (!currentProfile) {
            return [];
        }

        const avatarPhoto = currentProfile.avatar
            ? [{ id: "avatar", url: currentProfile.avatar }]
            : [];

        return [
            ...avatarPhoto,
            ...currentProfile.photos,
        ].filter(photo => photo?.url);
    }, [currentProfile]);

    const showNextProfile = () => {
        setCurrentIndex(prevIndex => prevIndex + 1);
    };

    if (isLoading) {
        return (
            <div id="user-profile-container">
                <div id="user-profile" className="dating-profile">
                    <p className="dating-status">Loading profiles...</p>
                </div>
            </div>
        );
    }

    if (error || !currentProfile) {
        return (
            <div id="user-profile-container">
                <div id="user-profile" className="dating-profile">
                    <p className="dating-status">{error || "No profiles yet"}</p>
                </div>
            </div>
        );
    }

    return(
        <div id="user-profile-container">
            <div id="user-profile" className="dating-profile">
                <div className="profile-header dating-header">
                    <img
                        className="profile-avatar"
                        src={photos[0] ? getPhotoUrl(photos[0].url) : "/1.png"}
                        alt={`${currentProfile.userName || "User"} avatar`}
                    />
                    <div className="profile-heading">
                        <p className="profile-kicker">Dating</p>
                        <h2>{fullName || currentProfile.userName || "Profile"}</h2>
                        <p className="profile-subtitle dating-username">
                            @{currentProfile.userName || "username"}
                        </p>
                    </div>
                </div>

                <section className="dating-photos" aria-label="Profile photos">
                    {photos.length > 0 ? (
                        photos.map((photo) => (
                            <img
                                key={photo.id ?? photo.url}
                                className="dating-photo"
                                src={getPhotoUrl(photo.url)}
                                alt={`${currentProfile.userName || "User"} profile`}
                            />
                        ))
                    ) : (
                        <div className="dating-photo dating-photo-empty">No photos</div>
                    )}
                </section>

                <div className="profile-grid">
                    <div className="field username">
                        <span className="field-label">Username</span>
                        <p className="dating-small">@{currentProfile.userName || "username"}</p>
                    </div>
                    <div className="field first-name">
                        <span className="field-label">First name</span>
                        <p>{currentProfile.firstName || "Not specified"}</p>
                    </div>
                    <div className="field last-name">
                        <span className="field-label">Last name</span>
                        <p>{currentProfile.lastName || "Not specified"}</p>
                    </div>
                    <div className="field bio">
                        <span className="field-label">Bio</span>
                        <p>{currentProfile.bio || "No bio yet"}</p>
                    </div>
                </div>
                <div className="profile-actions">
                    <Button onClick={showNextProfile}>Skip</Button>
                    <Button onClick={showNextProfile}>Like</Button>
                </div>

            </div>
        </div>
    );
}
