import React, { useEffect, useState } from 'react';
import '../css/SplashText.css';

const suffixes = [
  "", "2.0", "Reloaded", "Returns", "and the Chamber of Secrets", "Ultimate Edition",
  "Judgement Day", "Electric Boogaloo", "Origins", "NX", "Now in 4K", "with Cheese",
  "for Workgroups", "Turbo", "Quantum Edition", "AI-Powered", "Web3-Ready", "Serverless",
  "The Final Commit", "PeerSphere.dev", "Version Control", "Merge Conflict", "Syntax Error",
  "∞ Loop", "The Fellowship of the Norms", "Into the Shell", "Rise of the Peers",
  "The Peer Awakens", "Return of the Norminator", "The Linter Prophecy", "Clash of the Coders",
  "The Last Evaluator", "Forbidden Functions", "DOS Edition", "8-Bit Style", "Beta Max",
  "Now on Floppy", "Made in BASIC", "Dial-up Mode", "CRT Approved", "256 Colors",
  "Sponsored by Nobody™", "Now Gluten-Free", "Better Than ChatGPT", "Definitely Not a Cult",
  "Brought to You by Norms", "Still in Beta", "Now With Peerception™", "Evaluations? What Evaluations?"
];

export default function SplashText() {
  const [text, setText] = useState("Matcha");
  const [fade, setFade] = useState(false);

  useEffect(() => {
    const updateText = () => {
      setFade(true);
      setTimeout(() => {
        const suffix = suffixes[Math.floor(Math.random() * suffixes.length)];
        setText(`Matcha ${suffix}`.trim());
        setFade(false);
      }, 300);
    };

    updateText();
    const interval = setInterval(updateText, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <h1 className={`splash-text ${fade ? 'fade-out' : 'fade-in'}`}>
      {text}
    </h1>
  );
}
