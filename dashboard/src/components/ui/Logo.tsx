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
        viewBox="0 0 64 64"
        fill="none"
        aria-hidden="true"
      >
        <defs>
          <linearGradient id="kymaros-sg" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#3B82F6" />
            <stop offset="100%" stopColor="#1D4ED8" />
          </linearGradient>
        </defs>
        <path d="M32 4 C32 4 8 12 8 12 C6 13 4 16 4 19 L4 32 C4 46 16 56 32 62 C48 56 60 46 60 32 L60 19 C60 16 58 13 56 12 C56 12 32 4 32 4Z" fill="url(#kymaros-sg)" />
        <g transform="translate(32, 28)" stroke="white" fill="none" strokeWidth="2.5" strokeLinecap="round">
          <circle cx="0" cy="-10" r="4" />
          <line x1="0" y1="-6" x2="0" y2="14" />
          <line x1="-9" y1="2" x2="9" y2="2" />
          <path d="M-1 14 C-1 14, -11 12, -12 5" />
          <path d="M1 14 C1 14, 11 12, 12 5" />
        </g>
        <path d="M8 38 C16 34, 24 42, 32 36 C40 30, 48 38, 56 34" fill="none" stroke="rgba(255,255,255,0.5)" strokeWidth="1.5" strokeLinecap="round" />
        <g transform="translate(48, 50)">
          <circle cx="0" cy="0" r="7" fill="#10B981" />
          <path d="M-3 0 L-1 2 L4 -2" fill="none" stroke="white" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </g>
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
