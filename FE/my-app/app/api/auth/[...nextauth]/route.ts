import NextAuth, { AuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";
import type { User, Session } from "next-auth";
import type { JWT } from "next-auth/jwt";

declare module "next-auth" {
  interface Session {
    user: {
      id?: string;
      name?: string | null;
      email?: string | null;
      image?: string | null;
    };
  }
}
export const authOptions: AuthOptions = {
  providers: [
    CredentialsProvider({
      name: "Credentials",
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        // TODO: ini mock test , kalian validate credentials disini
        const users = [
          { id: "1", name: "Test User", email: "test@example.com", password: "password123" },
          { id: "2", name: "Jane Doe", email: "jane@example.com", password: "securepass" },
        ];
      
        // Find user matching the email and password
        const user = users.find(
          (u) => u.email === credentials?.email && u.password === credentials?.password
        );
      
        if (user) {
          // Return user object without password for the session
          const { password, ...userWithoutPassword } = user;
          return userWithoutPassword;
        }
      
        // Return null if credentials invalid
        return null;
      },
    }),
  ],
  session: {
    strategy: "jwt" ,
  },
  callbacks: {
    async jwt({ token, user }:{token:JWT; user?:User}) {
      if (user) {
        token.id = user.id;
      }
      return token;
    },
    async session({ session, token }: { session: Session; token: JWT }) {
      if (token && session.user) {
        session.user.id = token.id as string;
      }
      return session;
    },
  },
  secret: process.env.NEXTAUTH_SECRET,
};

const handler = NextAuth(authOptions);
export { handler as GET, handler as POST };
