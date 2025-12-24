import { useEffect, useMemo, useRef, useState } from "react";

type ToastVariant = "info" | "success" | "warning" | "error";

interface ToastProps {
  message: string;
  title?: string;
  variant?: ToastVariant;
  onClose: () => void;
  autoHideDuration?: number;
}

const variantStyles: Record<ToastVariant, { bg: string; text: string; border: string; iconBg: string }> = {
  info: {
    bg: "bg-white dark:bg-gray-800",
    text: "text-slate-800 dark:text-white",
    border: "border-blue-500",
    iconBg: "bg-blue-500 text-white",
  },
  success: {
    bg: "bg-white dark:bg-gray-800",
    text: "text-slate-800 dark:text-white",
    border: "border-green-600",
    iconBg: "bg-green-600 text-white",
  },
  warning: {
    bg: "bg-white dark:bg-gray-800",
    text: "text-slate-800 dark:text-white",
    border: "border-amber-500",
    iconBg: "bg-amber-500 text-white",
  },
  error: {
    bg: "bg-white dark:bg-gray-800",
    text: "text-slate-800 dark:text-white",
    border: "border-red-600",
    iconBg: "bg-red-600 text-white",
  },
};

const defaultTitle: Record<ToastVariant, string> = {
  info: "Info",
  success: "Success!",
  warning: "Heads up",
  error: "Error",
};

export function Toast({
  message,
  title,
  variant = "info",
  onClose,
  autoHideDuration = 3500,
}: ToastProps) {
  const tone = useMemo(() => variantStyles[variant] ?? variantStyles.info, [variant]);
  const [isVisible, setIsVisible] = useState(true);
  const timerRef = useRef<number | null>(null);
  const startRef = useRef<number>(0);
  const remainingRef = useRef<number>(autoHideDuration);

  useEffect(() => {
    startRef.current = Date.now();
    remainingRef.current = autoHideDuration;

    timerRef.current = window.setTimeout(() => {
      setIsVisible(false);
      // allow animation to play before calling onClose
      setTimeout(onClose, 200);
    }, remainingRef.current);

    return () => {
      if (timerRef.current) window.clearTimeout(timerRef.current);
    };
  }, [autoHideDuration, onClose]);

  function handleMouseEnter() {
    // pause
    if (timerRef.current) {
      window.clearTimeout(timerRef.current);
      const elapsed = Date.now() - startRef.current;
      remainingRef.current = Math.max(0, remainingRef.current - elapsed);
    }
  }

  function handleMouseLeave() {
    // resume
    startRef.current = Date.now();
    if (remainingRef.current <= 0) {
      setIsVisible(false);
      setTimeout(onClose, 200);
      return;
    }
    timerRef.current = window.setTimeout(() => {
      setIsVisible(false);
      setTimeout(onClose, 200);
    }, remainingRef.current);
  }

  function iconForVariant(v: ToastVariant) {
    switch (v) {
      case "success":
        return (
          <svg className="h-6 w-6" viewBox="0 0 20 20" fill="none" aria-hidden>
            <path d="M16.7 5.3a1 1 0 00-1.4-1.4L8 11.2 4.7 8a1 1 0 00-1.4 1.4l4 4a1 1 0 001.4 0l8-8z" fill="currentColor" />
          </svg>
        );
      case "warning":
        return (
          <svg className="h-6 w-6" viewBox="0 0 20 20" fill="none" aria-hidden>
            <path d="M9 2a1 1 0 01.99.858L10 3v6a1 1 0 01-1.994.117L8 9V3a1 1 0 011-1zm0 12a1.5 1.5 0 100 3 1.5 1.5 0 000-3z" fill="currentColor" />
          </svg>
        );
      case "error":
        return (
          <svg className="h-6 w-6" viewBox="0 0 20 20" fill="none" aria-hidden>
            <path d="M10 9l4-4m0 0l-4 4m4-4l-4 4M6 6l4 4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        );
      default:
        return (
          <svg className="h-6 w-6" viewBox="0 0 20 20" fill="none" aria-hidden>
            <path d="M9 2a1 1 0 011.044.96L10 3v6a1 1 0 01-2 .117L8 9V3a1 1 0 011-1zm0 12a1.5 1.5 0 100 3 1.5 1.5 0 000-3z" fill="currentColor" />
          </svg>
        );
    }
  }

  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      className={`pointer-events-auto fixed right-4 top-4 z-50 flex w-[360px] max-w-[calc(100vw-32px)]`}
    >
      <div
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`pointer-events-auto flex w-full items-start gap-4 rounded-lg border-b-4 ${tone.border} px-4 py-3 shadow-md transform transition-all duration-200 ${isVisible ? "animate-toast-in" : "opacity-0 translate-y-2 scale-98"} ${tone.bg}`}
      >
        <div className={`flex h-10 w-10 items-center justify-center rounded-full ${tone.iconBg} shrink-0`}>{iconForVariant(variant)}</div>
        <div className="flex min-w-0 flex-1 flex-col">
          <span className={`block truncate text-sm font-semibold ${tone.text}`}>{title ?? defaultTitle[variant]}</span>
          <span className="mt-1 block truncate text-sm text-slate-600 dark:text-slate-300">{message}</span>
        </div>
        <button
          type="button"
          aria-label="Dismiss notification"
          className="ml-3 rounded p-1 text-slate-600 hover:text-slate-800 dark:text-slate-300 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-indigo-500"
          onClick={() => {
            setIsVisible(false);
            setTimeout(onClose, 150);
          }}
        >
          <svg
            className="h-4 w-4"
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>
      </div>
    </div>
  );
}
