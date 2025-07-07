"use client";
import { useEffect, useState } from "react";

export default function BrainIntro({ onFinishAction }: { onFinishAction: () => void }) {
  const [pulse, setPulse] = useState(0);
  const [flash, setFlash] = useState(false);
  useEffect(() => {
    let t1 = setTimeout(() => setPulse(1), 400);
    let t2 = setTimeout(() => setPulse(2), 900);
    let t3 = setTimeout(() => setPulse(3), 1400);
    let t4 = setTimeout(() => setPulse(4), 2000);
    let t5 = setTimeout(() => setPulse(5), 2500);
    let t6 = setTimeout(() => setFlash(true), 3000);
    let t7 = setTimeout(() => onFinishAction(), 3400);
    return () => [t1, t2, t3, t4, t5, t6, t7].forEach(clearTimeout);
  }, [onFinishAction]);
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#1A1A1A]">
      <div
        className={`relative flex items-center justify-center transition-all duration-300 ${flash ? "" : ""}`}
        style={{ minHeight: 220 }}
      >
        {/* Brain SVG */}
        <svg
          width="160"
          height="160"
          viewBox="0 0 160 160"
          fill="none"
          className={`transition-all duration-300 ${pulse === 5 ? "opacity-0" : "opacity-100"}`}
          style={{
            filter:
              pulse === 5
                ? "none"
                : `drop-shadow(0 0 ${pulse === 4 ? 32 : 16 + 8 * (pulse % 2)}px #A9FBD7AA)`
          }}
        >
          <ellipse cx="80" cy="80" rx="60" ry="55" fill="#222" stroke="#A9FBD7" strokeWidth="4" />
          <path d="M60 80 Q55 60 80 60 Q105 60 100 80 Q105 100 80 100 Q55 100 60 80" stroke="#A9FBD7" strokeWidth="3" fill="none" />
          <circle cx="70" cy="75" r="7" fill="#A9FBD7" fillOpacity="0.7" />
          <circle cx="90" cy="85" r="6" fill="#A9FBD7" fillOpacity="0.5" />
        </svg>
        {/* Ripple flash */}
        {flash && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="animate-ripple w-[160px] h-[160px] rounded-full bg-[#A9FBD7] opacity-80" />
          </div>
        )}
        <style>{`
          @keyframes ripple {
            0% { transform: scale(1); opacity: 0.8; }
            100% { transform: scale(12); opacity: 0; }
          }
          .animate-ripple {
            animation: ripple 0.5s cubic-bezier(.4,2,.6,1) forwards;
          }
        `}</style>
      </div>
    </div>
  );
}
