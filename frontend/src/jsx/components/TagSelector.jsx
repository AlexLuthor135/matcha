import { useState } from "react";
import "./TagSelector.css"

export default function TagSelector({tags, onChange }){
    const [selectedTag, setSelectedTag] = useState([]);

    const toggleTag = (tag) => {
        let updated;

        if(selectedTag.includes(tag)) {
            updated = selectedTag.filter(t => t !== tag);
        } else {
            updated  = [...selectedTag, tag];
        }

        setSelectedTag(updated);
        onChange?.(updated);
    };

    return (
        <div>
            {tags.map(tag => (
                <button
                    key={tag}
                    type="button"
                    className={'tag ${selected.includes(tag) ? "active" : ""}'}
                    onClick={() => toggleTag(tag)}
                >
                    {tag}
                </button>
            ))}
        </div>
    );
}