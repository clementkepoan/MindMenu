import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

const accent = "#A9FBD7";

const navLinks = [
  { href: "#features", label: "Features" },
  { href: "#how-it-works", label: "How it Works" },
  { href: "/login", label: "Login" },
];

// (removed duplicate NavLink definition)

// Modern animated nav link with Framer Motion underline
const NavLink: React.FC<{ href: string; accent: string; children: React.ReactNode }> = ({ href, accent, children }) => {
  const [hovered, setHovered] = React.useState(false);
  return (
    <a
      href={href}
      className="relative px-2 py-1 transition-colors duration-200 hover:text-[#A9FBD7] focus:text-[#A9FBD7] focus:outline-none"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocus={() => setHovered(true)}
      onBlur={() => setHovered(false)}
      style={{ WebkitTapHighlightColor: 'transparent' }}
    >
      {children}
      <AnimatePresence>
        {hovered && (
          <motion.span
            layoutId="nav-underline"
            initial={{ width: 0, opacity: 0 }}
            animate={{ width: '100%', opacity: 1 }}
            exit={{ width: 0, opacity: 0 }}
            transition={{ duration: 0.28, ease: [0.4, 0, 0.2, 1] }}
            style={{
              position: 'absolute',
              left: 0,
              bottom: -2,
              height: 3,
              borderRadius: 2,
              background: accent,
              display: 'block',
            }}
          />
        )}
      </AnimatePresence>
    </a>
  );
};

const HeadersSection = () => {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header className="fixed top-0 w-full bg-[#181A1B]/70 backdrop-blur-lg text-white z-50 font-sans shadow-sm transition-all">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-12 py-2 flex justify-between items-center">
        {/* Logo */}
        <div className="text-2xl font-extrabold tracking-tight select-none" style={{fontFamily: 'Space Grotesk, Inter, sans-serif'}}>
          Mind<span className="text-[{accent}]" style={{color: accent}}>Menu</span>
        </div>

        {/* Desktop Nav */}
        <nav className="hidden md:flex gap-4 lg:gap-8 items-center text-base">
          {navLinks.map(link => (
            <NavLink key={link.href} href={link.href} accent={accent}>
              {link.label}
            </NavLink>
          ))}

          <a
            href="#demo"
            className="ml-2 px-4 py-1.5 rounded-lg font-semibold bg-[#A9FBD7]/10 border border-[#A9FBD7]/30 text-[#A9FBD7] hover:bg-[#A9FBD7]/20 hover:border-[#A9FBD7] transition-colors duration-200 shadow-sm"
          >
            Try Demo
          </a>
        </nav>

        {/* Hamburger for mobile */}
        <button
          className="md:hidden flex flex-col justify-center items-center w-10 h-10 rounded focus:outline-none focus:ring-2 focus:ring-[#A9FBD7]/60"
          aria-label="Open menu"
          onClick={() => setMobileOpen(v => !v)}
        >
          <span className="block w-6 h-0.5 bg-[#A9FBD7] mb-1 transition-all" style={{transform: mobileOpen ? 'rotate(45deg) translateY(7px)' : 'none'}} />
          <span className={`block w-6 h-0.5 bg-[#A9FBD7] mb-1 transition-all ${mobileOpen ? 'opacity-0' : ''}`} />
          <span className="block w-6 h-0.5 bg-[#A9FBD7] transition-all" style={{transform: mobileOpen ? 'rotate(-45deg) translateY(-7px)' : 'none'}} />
        </button>

        {/* Mobile Nav Drawer */}
        <div
          className={`fixed top-0 left-0 w-full h-full bg-[#181A1B]/90 backdrop-blur-lg z-50 flex flex-col items-center justify-center transition-all duration-300 ${mobileOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'}`}
        >
          <button
            className="absolute top-6 right-6 w-10 h-10 flex items-center justify-center rounded-full bg-[#23272A]/60 hover:bg-[#23272A]/80 text-[#A9FBD7] text-2xl focus:outline-none"
            aria-label="Close menu"
            onClick={() => setMobileOpen(false)}
          >
            ×
          </button>
          <nav className="flex flex-col gap-8 text-2xl font-semibold mt-8">
            {navLinks.map(link => (
              <a
                key={link.href}
                href={link.href}
                className="transition-colors duration-200 hover:text-[#A9FBD7] focus:text-[#A9FBD7] focus:outline-none"
                onClick={() => setMobileOpen(false)}
              >
                {link.label}
              </a>
            ))}
            <a
              href="#demo"
              className="px-6 py-2 rounded-lg font-semibold bg-[#A9FBD7]/10 border border-[#A9FBD7]/30 text-[#A9FBD7] hover:bg-[#A9FBD7]/20 hover:border-[#A9FBD7] transition-colors duration-200 shadow-sm"
              onClick={() => setMobileOpen(false)}
            >
              Try Demo
            </a>
          </nav>
        </div>
      </div>
      <style jsx>{`
        header {
          font-family: 'Space Grotesk', 'Inter', sans-serif;
        }
      `}</style>
    </header>
  );
};

export default HeadersSection;