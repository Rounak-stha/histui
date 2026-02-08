'use client'

import React from 'react'

import { useEffect, useRef } from 'react'

export default function Footer() {
	const ref = useRef<HTMLDivElement>(null)

	useEffect(() => {
		const handleMouseMove = (e: MouseEvent) => {
			if (!ref.current) return

			const rect = ref.current.getBoundingClientRect()
			const x = e.clientX - rect.left
			const y = e.clientY - rect.top

			ref.current.style.setProperty('--mouse-x', `${x}px`)
			ref.current.style.setProperty('--mouse-y', `${y}px`)
		}

		window.addEventListener('mousemove', handleMouseMove)
		return () => window.removeEventListener('mousemove', handleMouseMove)
	}, [])

	return (
		<footer
			ref={ref}
			className='relative bg-background border-t border-border py-20 px-4 overflow-hidden'
			style={
				{
					'--mouse-x': '0px',
					'--mouse-y': '0px'
				} as React.CSSProperties
			}
		>
			{/* Radial gradient following mouse */}
			<div
				className='absolute inset-0 pointer-events-none opacity-20'
				style={{
					background: `radial-gradient(circle 200px at var(--mouse-x, 0px) var(--mouse-y, 0px), rgba(140, 200, 100, 0.3) 0%, transparent 100%)`
				}}
			/>

			<div className='max-w-6xl mx-auto relative z-10'>
				<div className='grid md:grid-cols-4 gap-12 mb-12'>
					{/* Brand */}
					<div>
						<h3 className='text-2xl font-bold text-foreground mb-4'>histui</h3>
						<p className='text-sm text-muted-foreground'>
							Find what changes together. Refactor with confidence.
						</p>
					</div>

					{/* Community */}
					<div>
						<h4 className='font-semibold text-foreground mb-4'>Community</h4>
						<ul className='space-y-2 text-sm text-muted-foreground'>
							<li>
								<a
									href='https://github.com/Rounak-stha/histui/issues'
									className='hover:text-accent transition-colors'
								>
									GitHub
								</a>
							</li>
							<li>
								<a
									href='https://github.com/Rounak-stha/histui/issues'
									className='hover:text-accent transition-colors'
								>
									Issues
								</a>
							</li>
						</ul>
					</div>
				</div>

				{/* Divider */}
				<div className='border-t border-border pt-8'>
					<div className='flex flex-col md:flex-row items-center justify-between text-xs text-muted-foreground'>
						<p>© 2025 histui. A developer tool for better code.</p>
						<p className='font-mono'>
							Built with <span className='text-accent'>♡</span> by developers
						</p>
					</div>
				</div>
			</div>

			{/* ASCII Art */}
			<div className='absolute bottom-0 left-0 right-0 pointer-events-none opacity-10 whitespace-pre font-mono text-xs text-muted-foreground overflow-hidden'>
				{`
╔════════════════════════════════════════════════════════════════════╗
║                                                                    ║
║          The future of code analysis is here.                      ║
║                                                                    ║
╚════════════════════════════════════════════════════════════════════╝
        `}
			</div>
		</footer>
	)
}
