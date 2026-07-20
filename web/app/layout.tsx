import type { Metadata, Viewport } from 'next'
import { Analytics } from '@vercel/analytics/next'
import './globals.css'

export const metadata: Metadata = {
  title: {
    default: 'histui · Git history for better code changes',
    template: '%s · histui',
  },
  description: 'Find files that repeatedly changed together. Local, indexed Git-history context for refactors, architecture work, and coding agents.',
  keywords: ['Git history', 'file coupling', 'refactoring', 'coding agents', 'architecture analysis'],
  metadataBase: new URL('https://histui.vercel.app'),
  openGraph: {
    type: 'website',
    title: 'histui · See what your next change might miss',
    description: 'Focused historical coupling evidence for humans and coding agents.',
    images: ['/og-image.jpg'],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'histui · Git history for better code changes',
    description: 'Focused historical coupling evidence for humans and coding agents.',
    images: ['/og-image.jpg'],
  },
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#f4f2eb',
}

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}<Analytics /></body>
    </html>
  )
}
