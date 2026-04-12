import { cn } from "@/lib/utils";

type MaterialSymbolProps = {
    name: string;
    className?: string;
    filled?: boolean;
    size?: "sm" | "md" | "lg";
};

const sizeClass = {
    sm: "text-base",
    md: "text-xl",
    lg: "text-2xl",
} as const;

/** Google Material Symbols Outlined (loaded via index.html). */
export function MaterialSymbol({
    name,
    className,
    filled = false,
    size = "md",
}: MaterialSymbolProps) {
    return (
        <span
            className={cn("material-symbols-outlined align-middle", sizeClass[size], className)}
            style={
                filled
                    ? { fontVariationSettings: "'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24" }
                    : undefined
            }
            aria-hidden
        >
            {name}
        </span>
    );
}
