import type { FC } from 'react';

interface LogoProps {
  size?: 'sm' | 'md' | 'lg';
  showText?: boolean;
}

const sizeConfig = {
  sm: { icon: 20, text: 'text-sm', gap: 'gap-1.5' },
  md: { icon: 28, text: 'text-lg', gap: 'gap-2' },
  lg: { icon: 36, text: 'text-2xl', gap: 'gap-2.5' },
} as const;

const Logo: FC<LogoProps> = ({ size = 'md', showText = true }) => {
  const config = sizeConfig[size];

  return (
    <div className={`flex items-center ${config.gap}`}>
      <svg
        width={config.icon}
        height={config.icon}
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        {/* Anchor icon — geometric: ring + vertical line + crossbar + hook */}
        <circle cx="12" cy="5" r="3" />
        <line x1="12" y1="8" x2="12" y2="22" />
        <line x1="8" y1="12" x2="16" y2="12" />
        <path d="M5 19a7 7 0 0 0 14 0" />
      </svg>
      {showText && (
        <span className={`font-semibold tracking-tight ${config.text}`}>
          Kymaros
        </span>
      )}
    </div>
  );
};

export default Logo;
