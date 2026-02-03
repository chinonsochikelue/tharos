import { title } from '@/lib/layout.shared'
import { baseUrl, createMetadata } from '@/lib/metadata'
import '@/styles/globals.css'
import type { Viewport } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import { Body } from './layout.client'
import { Providers } from './providers'
import { NextProvider } from 'fumadocs-core/framework/next'
import { TreeContextProvider } from 'fumadocs-ui/contexts/tree'
import { source } from '@/lib/source'
import { url } from '@/lib/url'

const geist = Geist({
  variable: '--font-sans',
  subsets: ['latin'],
})

const mono = Geist_Mono({
  variable: '--font-mono',
  subsets: ['latin'],
})

// Enhanced SEO Metadata
export const metadata = createMetadata({
  // Primary SEO
  title: {
    default: "Tharos - AI-Powered Security & Quality Analysis",
    template: "%s | Tharos Security"
  },
  description: "Tharos is a next-gen security analysis tool combining fast AST scanning with deep AI semantic reasoning. Catch SQLi, secrets, and logic flaws before production.",
  metadataBase: baseUrl,
  alternates: {
    canonical: "https://tharos.dev/",
    languages: {
      "en-US": "https://tharos.dev/en-US",
      "x-default": "https://tharos.dev/",
    },
    types: {
      'application/rss+xml': [
        {
          title,
          url: url('/rss.xml'),
        },
      ],
    },
  },
  // Keywords for better discoverability
  keywords: [
    "security scanner",
    "static analysis",
    "sast",
    "ai security",
    "code quality",
    "devsecops",
    "golang",
    "ast analysis",
    "vulnerability detection",
    "secure coding",
    "sql injection",
    "secret detection",
    "automated code review",
    "tharos",
    "cybersecurity",
    "application security"
  ],

  // Author and creator information
  authors: [
    {
      name: "Tharos Security Team",
      url: "https://github.com/tharos-security"
    }
  ],
  creator: "Tharos Security Team",
  publisher: "Tharos Open Source Project",

  // Open Graph metadata for social sharing
  openGraph: {
    type: "website",
    locale: "en_US",
    url: "https://tharos.dev/",
    title: "Tharos - AI-Powered Security Analysis",
    description: "Catch vulnerabilities before they reach production. Deep AST analysis + AI semantic insights.",
    siteName: "Tharos Security",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Tharos Security Engine",
        type: "image/png",
      }
    ],
  },

  // Twitter Card metadata
  twitter: {
    card: "summary_large_image",
    site: "@tharos_security",
    creator: "@tharos_security",
    title: "Tharos - AI-Powered Security Analysis",
    description: "Catch vulnerabilities before they reach production. Deep AST analysis + AI semantic insights.",
    images: ["/twitter-card.png"],
  },

  // Additional metadata
  category: "Technology",
  classification: "Security Software",

  // App-specific metadata
  applicationName: "Tharos Docs",
  appleWebApp: {
    capable: true,
    title: "Tharos",
    statusBarStyle: "default",
  },

  // Verification and ownership
  verification: {
    google: "ayqp5JY2qzW_am918ybg1JSEJ84P5xCGxdbeD4JK0ik",
  },

  // Robots directives
  robots: {
    index: true,
    follow: true,
  },

  // Additional structured data
  other: {
    "application-name": "Tharos",
    "msapplication-TileColor": "#000000",
    "theme-color": "#0a0a0a",
  },
});

// Viewport configuration for Next.js 15
export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  userScalable: true,
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#ffffff" },
    { media: "(prefers-color-scheme: dark)", color: "#0a0a0a" },
  ],
  colorScheme: "dark",
};


export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html
      className={`${geist.variable} ${mono.variable}`}
      lang='en'
      suppressHydrationWarning
    >
      <head>
        {/* Additional SEO meta tags */}
        <meta name="format-detection" content="telephone=no" />
        <meta name="mobile-web-app-capable" content="yes" />

        {/* Favicons and app icons */}
        <link rel="icon" href="/favicon.ico" sizes="any" />

        {/* DNS prefetch for external resources */}
        <link rel="dns-prefetch" href="//fonts.googleapis.com" />

        {/* Structured Data - JSON-LD */}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "SoftwareApplication",
              "name": "Tharos",
              "operatingSystem": "Cross-platform",
              "applicationCategory": "DeveloperTool",
              "description": "AI-powered static analysis security tool.",
              "offers": {
                "@type": "Offer",
                "price": "0",
                "priceCurrency": "USD"
              }
            }),
          }}
        />
      </head>
      <Body tree={source.pageTree}>
        <NextProvider>
          <TreeContextProvider tree={source.pageTree}>
            <Providers>{children}</Providers>
          </TreeContextProvider>
        </NextProvider>
      </Body>
    </html>
  )
}
