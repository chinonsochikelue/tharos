import { RootProvider } from 'fumadocs-ui/provider/next';
import './global.css';
import { Outfit } from 'next/font/google';

const font = Outfit({
  subsets: ['latin'],
});

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={font.className} suppressHydrationWarning>
      <body className="flex flex-col min-h-screen bg-blue-deep text-white relative overflow-x-hidden">
        {/* Ambient Glows */}
        <div className="fixed top-[-10%] left-[-10%] w-[40%] h-[40%] bg-cyan-neon/10 blur-[120px] rounded-full pointer-events-none" />
        <div className="fixed bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-neon/10 blur-[120px] rounded-full pointer-events-none" />

        <RootProvider>{children}</RootProvider>
      </body>
    </html>
  );
}
