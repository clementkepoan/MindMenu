"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { cn } from "@/core/lib/utils";
import Link from "next/link";
import { BackgroundBeamsWithCollision } from "../components/ui/background-beams-with-collision";

export default function RegisterForm() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    email: "",
    password: "",
    name: "",
  });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData((prev) => ({
      ...prev,
      [e.target.name]: e.target.value,
    }));
    setError("");
    setSuccess("");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    // TODO: Replace with your registration API call
    if (!formData.email || !formData.password || !formData.name) {
      setError("All fields are required.");
      return;
    }
    // Mock success
    setSuccess("Registration successful! You can now log in.");
    setTimeout(() => router.push("/auth/Login"), 1500);
  };

  return (
    <BackgroundBeamsWithCollision className="flex min-h-screen items-center justify-center bg-[#1A1A1A]">
      <form
        onSubmit={handleSubmit}
        className={cn(
          "w-full max-w-md rounded-2xl border border-[#222] bg-[#23272e]/80 p-8",
          "flex flex-col gap-6",
          "shadow-[0_8px_32px_0_rgba(169,251,215,0.25),0_1.5px_10px_0_rgba(169,251,215,0.10)]"
        )}
      >
        <h2 className="text-3xl font-bold text-[#A9FBD7] font-sora text-center mb-2">
          Register 
        </h2>
        <div className="flex flex-col gap-2">
          <label htmlFor="name" className="text-[#A9FBD7] font-medium">
            Name
          </label>
          <input
            id="name"
            name="name"
            type="text"
            autoComplete="name"
            required
            value={formData.name}
            onChange={handleChange}
            className="rounded-lg bg-[#181C1F] border border-[#333] px-4 py-2 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#A9FBD7] transition"
            placeholder="Your name"
          />
        </div>
        <div className="flex flex-col gap-2">
          <label htmlFor="email" className="text-[#A9FBD7] font-medium">
            Email
          </label>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            required
            value={formData.email}
            onChange={handleChange}
            className="rounded-lg bg-[#181C1F] border border-[#333] px-4 py-2 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#A9FBD7] transition"
            placeholder="you@email.com"
          />
        </div>
        <div className="flex flex-col gap-2">
          <label htmlFor="password" className="text-[#A9FBD7] font-medium">
            Password
          </label>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete="new-password"
            required
            value={formData.password}
            onChange={handleChange}
            className="rounded-lg bg-[#181C1F] border border-[#333] px-4 py-2 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#A9FBD7] transition"
            placeholder="Create a password"
          />
        </div>
        {error && (
          <div className="rounded bg-[#A9FBD7]/10 text-[#A9FBD7] px-3 py-2 text-center text-sm font-medium border border-[#A9FBD7]/30">
            {error}
          </div>
        )}
        {success && (
          <div className="rounded bg-[#A9FBD7]/10 text-[#A9FBD7] px-3 py-2 text-center text-sm font-medium border border-[#A9FBD7]/30">
            {success}
          </div>
        )}
        <button
          type="submit"
          className="mt-2 rounded-lg bg-[#A9FBD7] px-4 py-2 font-bold text-[#1A1A1A] shadow-md transition hover:bg-[#7beec2] focus:outline-none focus:ring-2 focus:ring-[#A9FBD7]"
        >
          Register
        </button>
        <div className="text-center text-sm text-[#A9FBD7]/70 mt-2">
          Already have an account?{" "}
          <Link href="/auth/Login"
            className="underline hover:text-[#A9FBD7] transition"
          >
            Log In
          </Link>
        </div>
      </form>
    </BackgroundBeamsWithCollision>
  );
}
