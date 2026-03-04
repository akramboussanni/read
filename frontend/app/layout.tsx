import type { Metadata } from "next";
import { Nunito } from "next/font/google";
import "./globals.css";
import { AuthInitializer } from "@/components/auth-initializer";
import { ConditionalHeader } from "@/components/conditional-header";
import { EmailVerificationGate } from "@/components/email-verification-gate";

const nunito = Nunito({
  variable: "--font-nunito",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800", "900"],
});

export const metadata: Metadata = {
  title: "Iqra · Apprends en t'amusant !",
  description:
    "Apprends et mémorise grâce à des quiz interactifs, des parcours ludiques et un suivi de progression amusant.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="fr">
      <body className={`${nunito.variable} antialiased`}>
        <AuthInitializer>
          <EmailVerificationGate>
            <ConditionalHeader />
            <main className="pt-16">{children}</main>
          </EmailVerificationGate>
        </AuthInitializer>
      </body>
    </html>
  );
}
