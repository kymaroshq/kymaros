import { Lock, Sparkles } from 'lucide-react';

interface UpgradePromptProps {
  feature: string;
  tier: 'team' | 'enterprise';
  variant?: 'inline' | 'overlay' | 'banner';
  className?: string;
}

export default function UpgradePrompt({ feature, tier, variant = 'inline', className = '' }: UpgradePromptProps) {
  const tierLabel = tier === 'team' ? 'Team' : 'Enterprise';

  if (variant === 'banner') {
    return (
      <div className={`flex items-center justify-between gap-4 rounded-xl border border-blue-500/20 bg-blue-500/5 px-5 py-3.5 ${className}`}>
        <div className="flex items-center gap-3">
          <Lock className="h-4 w-4 shrink-0 text-blue-400" />
          <p className="text-sm text-slate-300">
            <span className="font-medium text-slate-200">{feature}</span>
            {' '}is available on Kymaros {tierLabel}.
          </p>
        </div>
        <a
          href="mailto:sales@kymaros.io"
          className="shrink-0 rounded-lg bg-blue-600 px-3.5 py-1.5 text-xs font-semibold text-white transition-colors hover:bg-blue-500"
        >
          Upgrade
        </a>
      </div>
    );
  }

  if (variant === 'overlay') {
    return (
      <div className={`absolute inset-0 z-10 flex flex-col items-center justify-center rounded-xl bg-slate-900/80 backdrop-blur-sm ${className}`}>
        <div className="flex flex-col items-center gap-3 text-center px-6">
          <div className="rounded-xl bg-blue-500/10 p-3">
            <Sparkles className="h-6 w-6 text-blue-400" />
          </div>
          <h4 className="text-sm font-semibold text-slate-200">{feature}</h4>
          <p className="text-xs text-slate-400">Available on Kymaros {tierLabel}</p>
          <a
            href="mailto:sales@kymaros.io"
            className="mt-1 rounded-lg bg-blue-600 px-4 py-2 text-xs font-semibold text-white transition-colors hover:bg-blue-500"
          >
            Start Free Trial
          </a>
        </div>
      </div>
    );
  }

  // inline (default)
  return (
    <div className={`flex items-center gap-3 rounded-lg border border-slate-700/50 bg-slate-800/50 px-4 py-3 ${className}`}>
      <Lock className="h-4 w-4 shrink-0 text-slate-500" />
      <p className="text-xs text-slate-400">
        <span className="font-medium text-slate-300">{feature}</span>
        {' '}— available on Kymaros {tierLabel}.{' '}
        <a href="mailto:sales@kymaros.io" className="text-blue-400 hover:text-blue-300">
          Upgrade &rarr;
        </a>
      </p>
    </div>
  );
}
