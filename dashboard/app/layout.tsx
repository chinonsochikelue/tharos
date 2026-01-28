import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"
import { Header } from "@/components/header";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Tharos Dashboard - Security & Compliance",
  description: "AI-powered security analysis and compliance management",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.className} bg-sidebar text-sidebar-foreground`}>
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset>
            <div className="flex flex-col h-full">
              <Header />
              <main className="flex-1 overflow-y-auto p-6">
                {children}
              </main>
            </div>
          </SidebarInset>
        </SidebarProvider>
      </body>
    </html>
  );
}
