/* eslint-disable react/prop-types */
import { useEffect, useState } from "react";
import "./ProfilePhotos.css";

export default function ProfilePhotos({ photos, deletingPhotoId, onUpload, onDelete }) {
    const [isDeletingPhoto, setIsDeletingPhoto] = useState(false);
    const photoList = Array.isArray(photos) ? photos : [];

    useEffect(() => {
        if (photoList.length === 0) {
            setIsDeletingPhoto(false);
        }
    }, [photoList.length]);

    const toggleDeleteMode = () => {
        if (photoList.length > 0) {
            setIsDeletingPhoto(prev => !prev);
        }
    };

    return (
        <>
            <div className="photo-actions">
                <label className="photos-upload-button">
                    Add photos
                    <input
                        type="file"
                        accept="image/*"
                        multiple
                        onChange={onUpload}
                    />
                </label>
                <button
                    type="button"
                    className="photos-delete-button"
                    disabled={photoList.length === 0}
                    onClick={toggleDeleteMode}
                >
                    {isDeletingPhoto ? "Cancel" : "Delete photo"}
                </button>
            </div>

            <section className="profile-photos" aria-label="Profile photos">
                {photoList.map((photo) => (
                    <div
                        key={photo.id ?? photo.url}
                        className={`profile-photo-frame ${isDeletingPhoto ? "is-delete-mode" : ""}`}
                    >
                        <img
                            className="profile-photo"
                            src={photo.url ? "/backend" + photo.url : ""}
                            alt="User uploaded"
                        />
                        {isDeletingPhoto && (
                            <button
                                type="button"
                                className="profile-photo-delete"
                                disabled={deletingPhotoId === photo.id}
                                onClick={() => onDelete(photo)}
                            >
                                {deletingPhotoId === photo.id ? "..." : "Delete"}
                            </button>
                        )}
                    </div>
                ))}
                {Array.from({ length: Math.max(0, 5 - photoList.length) }).map((_, index) => (
                    <div key={`empty-photo-${index}`} className="profile-photo profile-photo-empty" />
                ))}
            </section>
        </>
    );
}
