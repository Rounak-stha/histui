import React from 'react'
import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'

import './globals.css'

const _geist = Geist({ subsets: ['latin'] })
const _geistMono = Geist_Mono({ subsets: ['latin'] })

export const metadata: Metadata = {
	title: 'histui – Find what changes together',
	description:
		'Discover file couplings and refactor with confidence. Interactive git history visualization for better code organization.',
	generator: 'v0.app',
	viewport: {
		width: 'device-width',
		initialScale: 1,
		userScalable: false
	}
}

export default function RootLayout({
	children
}: Readonly<{
	children: React.ReactNode
}>) {
	return (
		<html lang='en'>
			<body className='font-sans antialiased'>{children}</body>
		</html>
	)
}
