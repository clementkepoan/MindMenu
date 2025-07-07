
import { NextResponse } from "next/server";
import bcrypt from "bcryptjs";

export async function POST(req: Request) {
  const { name, email, password } = await req.json();

  if (!name || !email || !password) {
    return NextResponse.json({ message: "All fields are required" }, { status: 400 });
  }

  // Mock: Check if email already exists (you can expand this logic)
  if (email === "test@example.com") {
    return NextResponse.json({ message: "Email already in use" }, { status: 409 });
  }

  // Mock: Hash password (in real case, save to DB)
  const hashedPassword = await bcrypt.hash(password, 10);

  console.log("New user registered:", { name, email, hashedPassword });

  return NextResponse.json({ message: "User registered successfully" }, { status: 201 });
}
