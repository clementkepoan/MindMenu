"use client"

import React, { useState } from 'react'
import BrainIntro from './BrainIntro';
import HeadersSection from './HeadersSection';
import HeroSection from './HeroSection';

const LandingPage = () => {
  const [showIntro,SetShowIntro]= useState(true);

  return(
    <>
    {showIntro ? (
      <BrainIntro onFinishAction={() => SetShowIntro(false)} />
    ) : (
      <>
        <HeadersSection />
        <HeroSection />
      </>
    )}
    </>
  );

};
export default LandingPage;