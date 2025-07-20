import React from 'react';
import '../css/Static.css';

const notionLink = import.meta.env.VITE_NOTION_LINK;

const Static = () => {
  return (
    <div className="Static">
        <p>
            Developed by 42 WOBBER. Please refer to this Notion Page:{" "}
            <a href={notionLink} target="_blank" rel="noopener noreferrer">
                PeerSphere
            </a>
        </p>
    </div>
  );
};

export default Static;
