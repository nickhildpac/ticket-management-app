import { useMemo, useState } from "react";

type AlertVariant = "info" | "success" | "warning" | "error";

interface AlertProps {
  message: string;
  variant?: AlertVariant;
  className?: string;
  onClose?: () => void;
}

const variantStyles: Record<AlertVariant, string> = {
  info: "bg-blue-50 text-blue-900 border-blue-200 dark:bg-blue-950 dark:text-blue-50 dark:border-blue-900",
  success: "bg-green-50 text-green-900 border-green-200 dark:bg-green-950 dark:text-green-50 dark:border-green-900",
  warning: "bg-amber-50 text-amber-900 border-amber-200 dark:bg-amber-950 dark:text-amber-50 dark:border-amber-900",
  error: "bg-red-50 text-red-900 border-red-200 dark:bg-red-950 dark:text-red-50 dark:border-red-900",
};

export function Alert({ message, variant = "info", className = "", onClose }: AlertProps) {
  const [isOpen, setIsOpen] = useState(true);

  const tone = useMemo(() => variantStyles[variant] ?? variantStyles.info, [variant]);

  if (!isOpen) return null;

  return (
    <div
      role="alert"
      className={`relative flex items-start gap-3 rounded-md border px-4 py-3 text-sm shadow-sm transition-colors duration-200 ${tone} ${className}`}
    >
      <span className="leading-relaxed">{message}</span>
      <button
        type="button"
        aria-label="Dismiss alert"
        className="ml-auto rounded p-1 text-current/70 transition hover:text-current focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-current focus-visible:ring-offset-2 focus-visible:ring-offset-transparent"
        onClick={() => {
          setIsOpen(false);
          onClose?.();
        }}
      >
        <svg
          className="h-4 w-4"
          viewBox="0 0 16 16"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          aria-hidden="true"
        >
          <path
            d="M4 4l8 8M12 4l-8 8"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
          />
        </svg>
      </button>
    </div>
  );
}
