import { useState, useEffect } from 'react';
import { Clock, X } from 'lucide-react';
import { useLicenseContext } from '../../hooks/useLicense';

const DISMISS_KEY = 'kymaros_banner_dismissed';
const DISMISS_DURATION_MS = 7 * 24 * 60 * 60 * 1000; // 1 week

export default function TrialBanner() {
  const { isTrialing, trialDaysLeft, isCommunity } = useLicenseContext();
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    const ts = localStorage.getItem(DISMISS_KEY);
    if (ts && Date.now() - parseInt(ts, 10) < DISMISS_DURATION_MS) {
      setDismissed(true);
    }
  }, []);

  const handleDismiss = () => {
    setDismissed(true);
    localStorage.setItem(DISMISS_KEY, Date.now().toString());
  };

  // Trial banner (not dismissable)
  if (isTrialing && trialDaysLeft != null) {
    const urgent = trialDaysLeft <= 3;
    return (
      <div
        className={`flex items-center justify-between px-4 py-2 text-sm ${
          urgent
            ? 'bg-red-500/10 border-b border-red-500/20'
            : 'bg-amber-500/10 border-b border-amber-500/20'
        }`}
      >
        <div className="flex items-center gap-2">
          <Clock className={`h-4 w-4 ${urgent ? 'text-red-400' : 'text-amber-400'}`} />
          <span className={urgent ? 'text-red-300' : 'text-amber-300'}>
            Trial &mdash; {trialDaysLeft} day{trialDaysLeft !== 1 ? 's' : ''} left
          </span>
        </div>
        <a
          href="https://kymaros.io/#pricing"
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs font-medium text-white hover:underline"
        >
          Upgrade now &rarr;
        </a>
      </div>
    );
  }

  // Community nudge (dismissable, reappears after 1 week)
  if (isCommunity && !dismissed) {
    return (
      <div className="flex items-center justify-between px-4 py-2 text-sm bg-blue-500/5 border-b border-blue-500/10">
        <span className="text-slate-400">
          You&apos;re on Kymaros Community.{' '}
          <a
            href="https://kymaros.io/#pricing"
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-400 hover:text-blue-300"
          >
            Upgrade to Team &rarr;
          </a>
          <span className="text-slate-500 ml-1">
            for 90-day history, compliance reports, and more.
          </span>
        </span>
        <button
          type="button"
          onClick={handleDismiss}
          className="text-slate-500 hover:text-slate-300 p-1"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
    );
  }

  return null;
}
