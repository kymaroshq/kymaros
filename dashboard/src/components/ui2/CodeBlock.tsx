import { cn } from '@/lib/utils';
import { Copy, Check } from 'lucide-react';
import { useState } from 'react';

interface CodeBlockProps {
  children: string;
  maxLines?: number;
  showCopy?: boolean;
  className?: string;
}

export function CodeBlock({ children, maxLines, showCopy = true, className }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const copy = () => {
    void navigator.clipboard.writeText(children);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className={cn('relative group', className)}>
      <pre
        className={cn(
          'bg-surface-0 border border-border-subtle rounded-md px-3 py-2 overflow-x-auto',
          'text-2xs font-mono text-text-secondary leading-relaxed'
        )}
        style={maxLines ? { maxHeight: `${maxLines * 1.25}rem`, overflowY: 'auto' } : undefined}
      >
        <code>{children}</code>
      </pre>
      {showCopy && (
        <button
          onClick={copy}
          className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-150 h-6 w-6 flex items-center justify-center rounded bg-surface-3 border border-border-subtle text-text-tertiary hover:text-text-primary"
        >
          {copied ? <Check className="h-3 w-3 text-status-success" /> : <Copy className="h-3 w-3" />}
        </button>
      )}
    </div>
  );
}
