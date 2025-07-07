
import React from "react";

import { BackgroundLines } from "./ui/background-lines";

const HeroSection = () => {
  return (
    <section className="relative min-h-screen w-full flex items-center justify-center px-0 py-0 text-white overflow-hidden">
      {/* Glowing Stars Animated Background */}
      <div className="absolute inset-0 -z-10 w-full h-full flex items-center justify-center pointer-events-none">
        <BackgroundLines className="absolute inset-0 w-full h-full min-h-screen min-w-full max-w-none max-h-none rounded-none border-0 bg-transparent scale-[2]" children={undefined} />
      </div>
      <div className="relative z-10 text-center">
        <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight">
          Welcome to MindMenu
        </h1>
        <p className="mt-4 text-sm sm:text-base lg:text-lg text-white/70 max-w-xl mx-auto">
          Your intelligent restaurant assistant powered by conversation.
        </p>
      </div>
    </section>
  );
};

export default HeroSection;
