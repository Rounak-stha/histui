'use client'

import { CSSProperties } from 'react'

interface GlitchTextProps {
  children: string
  className?: string
}

export default function GlitchText({
  children,
  className = '',
}: GlitchTextProps) {
  return (
    <div
      className={`relative inline-block ${className}`}
      style={
        {
          '--glitch-offset': '4px',
        } as CSSProperties
      }
    >
      <span className="relative z-0">{children}</span>
      <span
        className="absolute top-0 left-0 text-accent opacity-0 mix-blend-screen animate-pulse"
        style={{
          transform: 'translate(2px, 2px)',
          animation: 'glitch-red 0.3s infinite',
        }}
      >
        {children}
      </span>
      <span
        className="absolute top-0 left-0 text-blue-500 opacity-0 mix-blend-screen animate-pulse"
        style={{
          transform: 'translate(-2px, -2px)',
          animation: 'glitch-blue 0.3s infinite 0.15s',
        }}
      >
        {children}
      </span>

      <style jsx>{`
        @keyframes glitch-red {
          0% {
            opacity: 0;
          }
          20% {
            opacity: 0.8;
          }
          40% {
            opacity: 0;
          }
        }

        @keyframes glitch-blue {
          0% {
            opacity: 0;
          }
          20% {
            opacity: 0.8;
          }
          40% {
            opacity: 0;
          }
        }
      `}</style>
    </div>
  )
}
