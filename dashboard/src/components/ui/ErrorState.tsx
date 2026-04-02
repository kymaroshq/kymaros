import { useState, type FC } from 'react';
import { AlertCircle, RefreshCw, ChevronDown, ChevronRight } from 'lucide-react';

interface ErrorStateProps {
  message: string;
  onRetry?: () => void;
  detail?: string;
}

const ErrorState: FC<ErrorStateProps> = ({ message, onRetry, detail }) => {
  const [showDetail, setShowDetail] = useState(false);

  return (
    <div className="flex flex-col items-center justify-center rounded-lg border border-red-200 bg-red-50 px-6 py-10 text-center dark:border-red-900/50 dark:bg-red-950/20">
      <AlertCircle className="mb-3 h-10 w-10 text-[#ef4444]" />
      <p className="text-sm font-medium text-slate-900 dark:text-white">
        {message}
      </p>

      {detail && (
        <div className="mt-3 w-full max-w-md">
          <button
            type="button"
            className="inline-flex items-center gap-1 text-xs text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200"
            onClick={() => setShowDetail(!showDetail)}
          >
            {showDetail ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronRight className="h-3 w-3" />
            )}
            Details
          </button>
          {showDetail && (
            <pre className="mt-2 max-h-40 overflow-auto rounded bg-slate-100 p-3 text-left font-mono text-xs text-slate-600 dark:bg-slate-800 dark:text-slate-400">
              {detail}
            </pre>
          )}
        </div>
      )}

      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-4 inline-flex items-center gap-2 rounded-md bg-[#ef4444] px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 dark:focus:ring-offset-slate-900"
        >
          <RefreshCw className="h-4 w-4" />
          Retry
        </button>
      )}
    </div>
  );
};

export default ErrorState;
