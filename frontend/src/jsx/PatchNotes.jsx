import React, { useState, useEffect } from 'react';
import '../css/PatchNotes.css';

export default function PatchNotesPanel() {
  const [notes, setNotes] = useState([]);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    fetch('/patchNotes.json')
      .then((res) => res.json())
      .then(setNotes)
      .catch((err) => console.error('Failed to load patch notes:', err));
  }, []);

  return (
    <div className="patch-notes-panel">
      <button className="toggle-button" onClick={() => setVisible(!visible)}>
        📋 Patch Notes
      </button>
      {visible && (
        <div className="notes-content">
          <h3>Latest Updates</h3>
          <ul>
            {notes.map((note, index) => (
              <li key={index}>
                <strong>{note.version}</strong> – <em>{note.date}</em>
                <ul>
                  {note.notes.map((n, i) => (
                    <li key={i}>{n}</li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
