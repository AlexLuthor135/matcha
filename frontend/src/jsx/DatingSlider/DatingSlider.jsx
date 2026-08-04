import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import axiosInstance from "../AxiosInstance";
import { useAuth } from "../AuthProvider";
import { Button } from "../components/Button";
import "../UserProfile/UserProfile.css";
import "./DatingSlider.css";

const profileDecisionStoragePrefix = "matcha.profile-decisions";

function normalizeProfile(profile) {
    if (!profile || typeof profile !== "object") {
        return null;
    }

    return {
        id: profile.id,
        userName: String(profile.user_name ?? ""),
        firstName: String(profile.first_name ?? ""),
        lastName: String(profile.last_name ?? ""),
        avatar: String(profile.avatar ?? ""),
        photos: Array.isArray(profile.photos) ? profile.photos : [],
        bio: String(profile.bio ?? ""),
    };
}

function getPhotoUrl(url) {
    return url ? "/backend" + url : "";
}

function getProfileDecisionStorageKey(userID) {
    return `${profileDecisionStoragePrefix}.${userID || "guest"}`;
}

function readProfileDecisions(userID) {
    try {
        const rawDecisions = localStorage.getItem(getProfileDecisionStorageKey(userID));
        const decisions = rawDecisions ? JSON.parse(rawDecisions) : {};
        return decisions && typeof decisions === "object" ? decisions : {};
    } catch (storageError) {
        console.log("PROFILE DECISIONS READ ERROR", storageError);
        return {};
    }
}

function writeProfileDecision(userID, profileID, decision) {
    const decisions = readProfileDecisions(userID);
    decisions[String(profileID)] = {
        decision,
        updatedAt: new Date().toISOString(),
    };

    localStorage.setItem(getProfileDecisionStorageKey(userID), JSON.stringify(decisions));
}

function buildDisplayName(profileData, fallbackUserID) {
    const fullName = [
        String(profileData?.firstName ?? "").trim(),
        String(profileData?.lastName ?? "").trim(),
    ].filter(Boolean).join(" ");

    if (fullName) {
        return fullName;
    }

    if (profileData?.userName) {
        return String(profileData.userName);
    }

    return fallbackUserID ? `User #${fallbackUserID}` : "Someone";
}

export default function DatingSlider() {
    const navigate = useNavigate();
    const { userID } = useAuth();
    const [profiles, setProfiles] = useState([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [currentUserName, setCurrentUserName] = useState("");
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState("");

    useEffect(() => {
        const getCurrentUserName = async () => {
            try {
                const response = await axiosInstance.get("/backend/api/accounts/profile");
                setCurrentUserName(buildDisplayName(normalizeProfile(response.data), userID));
            } catch (requestError) {
                console.log("CURRENT PROFILE ERROR", requestError);
                setCurrentUserName(buildDisplayName(null, userID));
            }
        };

        getCurrentUserName();
    }, [userID]);

    useEffect(() => {
        const getProfiles = async () => {
            try {
                const response = await axiosInstance.get("/backend/api/accounts/profiles/feed");
                const profileList = Array.isArray(response.data?.profiles)
                    ? response.data.profiles.map(normalizeProfile).filter(Boolean)
                    : [];
                const decisions = readProfileDecisions(userID);
                const visibleProfiles = profileList.filter((profile) => !decisions[String(profile.id)]);

                setProfiles(visibleProfiles);
                setCurrentIndex(0);
            } catch (requestError) {
                console.log("PROFILE FEED ERROR", requestError);
                setError("Could not load profiles");
            } finally {
                setIsLoading(false);
            }
        };

        getProfiles();
    }, [userID]);

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

    const sendLikeNotification = async (profileID) => {
        const likedByName = currentUserName || buildDisplayName(null, userID);

        try {
            await axiosInstance.post("/backend/api/accounts/notifications/send", {
                user_id: profileID,
                title: "New like",
                message: `${likedByName} liked your profile`,
                data: {
                    kind: "profile_like",
                    liked_by: {
                        id: userID,
                        name: likedByName,
                    },
                },
            });
        } catch (requestError) {
            console.log("LIKE NOTIFICATION ERROR", requestError);
        }
    };

    const markProfile = async (decision) => {
        if (!currentProfile) {
            return;
        }

        if (decision === "like") {
            await sendLikeNotification(currentProfile.id);
        }

        try {
            writeProfileDecision(userID, currentProfile.id, decision);
        } catch (storageError) {
            console.log("PROFILE DECISION WRITE ERROR", storageError);
        }

        const nextProfiles = profiles.filter(
            (profile) => profile.id !== currentProfile.id
        );
        setProfiles(nextProfiles);
        setCurrentIndex((prevIndex) => Math.max(0, Math.min(prevIndex, nextProfiles.length - 1)));
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
                    <Button onClick={() => markProfile("dislike")}>Dislike</Button>
                    <Button onClick={() => navigate(`/chat?user=${currentProfile.id}`)}>Message</Button>
                    <Button onClick={() => markProfile("like")}>Like</Button>
                </div>

            </div>
        </div>
    );
}
