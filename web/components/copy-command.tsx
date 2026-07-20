'use client'

import { Check, Copy } from 'lucide-react'
import { useState } from 'react'

export default function CopyCommand({ command, label }: { command: string; label?: string }) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    await navigator.clipboard.writeText(command)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="copy-command">
      {label && <span className="command-label">{label}</span>}
      <code><span className="prompt">$</span> {command}</code>
      <button onClick={copy} aria-label="Copy command">{copied ? <Check size={16} /> : <Copy size={16} />}</button>
    </div>
  )
}
