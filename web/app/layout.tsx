import React from 'react'
import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'

import './globals.css'

export const metadata: Metadata = {
  title: 'histui – Detect file coupling and refactor with confidence',
  description:
    'histui analyzes your git history to reveal file coupling patterns. Understand your codebase architecture, plan refactors, and ship better code.',
  generator: 'v0.app',
  keywords: [
    'git analysis',
    'file coupling',
    'code refactoring',
    'architecture analysis',
    'developer tools',
  ],
  viewport: {
    width: 'device-width',
    initialScale: 1,
    userScalable: false,
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://histui.dev',
    title: 'histui – Detect file coupling and refactor with confidence',
    description:
      'Analyze git history to understand file coupling patterns and plan architectural refactors.',
    siteName: 'histui',
    images: [
      {
        url: '/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'histui - Git history analysis tool',
        type: 'image/jpeg',
      },
      {
        url: '/og-image.jpg',
        width: 800,
        height: 420,
        alt: 'histui - Git history analysis tool',
        type: 'image/jpeg',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'histui – Detect file coupling and refactor with confidence',
    description:
      'Analyze your git history to understand code architecture and plan refactors.',
    images: ['/og-image.jpg'],
    creator: '@histui',
    site: '@histui',
  },
  metadataBase: new URL('https://histui.dev'),
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <body className="font-sans antialiased">{children}</body>
    </html>
  )
}
