import type { VariantProps } from "class-variance-authority"
import { cva } from "class-variance-authority"

export { default as Badge } from "./Badge.vue"

export const badgeVariants = cva(
  "inline-flex gap-1 items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground hover:bg-primary/80",
        secondary:
          "border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",
        destructive:
          "border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80",
        outline: "text-foreground",
        bullish:
          "border-emerald-500/30 bg-emerald-500/10 text-emerald-400 font-medium",
        bearish:
          "border-rose-500/30 bg-rose-500/10 text-rose-400 font-medium",
        neutral:
          "border-sky-500/30 bg-sky-500/10 text-sky-400 font-medium",
        safe:
          "border-emerald-500/40 bg-emerald-500/15 text-emerald-400 font-semibold",
        warning:
          "border-amber-500/40 bg-amber-500/15 text-amber-400 font-semibold",
        danger:
          "border-rose-500/40 bg-rose-500/15 text-rose-400 font-semibold",
        terminal:
          "border-border bg-card/80 text-muted-foreground font-mono text-[11px]",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
)

export type BadgeVariants = VariantProps<typeof badgeVariants>
