import Link from "next/link";
import type { ReactNode } from "react";

import { Button } from "@/shared/ui";

type RouteAction = {
  href: string;
  label: string;
  variant?: "default" | "secondary";
};

type RouteStageProps = {
  actions?: RouteAction[];
  children?: ReactNode;
  description: string;
  eyebrow: string;
  highlights: string[];
  title: string;
};

/**
 * route ごとの意図を説明するステージ UI を表示する。
 */
export function RouteStage({
  actions = [],
  children,
  description,
  eyebrow,
  highlights,
  title,
}: RouteStageProps) {
  return (
    <section className="grid gap-5 lg:grid-cols-[1.45fr_0.95fr]">
      <article className="rounded-[2rem] border border-white/80 bg-white/74 p-7 shadow-[0_24px_80px_rgba(87,38,8,0.12)] backdrop-blur-lg sm:p-8">
        <p className="text-xs font-semibold uppercase tracking-[0.28em] text-accent">{eyebrow}</p>
        <h1 className="mt-5 text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">{title}</h1>
        <p className="mt-4 max-w-3xl text-sm leading-7 text-muted sm:text-base">{description}</p>
        <div className="mt-8 flex flex-wrap gap-3">
          {actions.map((action) => (
            <Button key={`${action.href}:${action.label}`} asChild variant={action.variant}>
              <Link href={action.href}>{action.label}</Link>
            </Button>
          ))}
        </div>
        {children ? <div className="mt-8">{children}</div> : null}
      </article>
      <aside className="rounded-[2rem] border border-white/70 bg-[#1b120f] p-6 text-white shadow-[0_24px_80px_rgba(35,16,8,0.24)]">
        <p className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-200/80">design intent</p>
        <div className="mt-5 grid gap-3">
          {highlights.map((item) => (
            <div
              key={item}
              className="rounded-[1.25rem] border border-white/8 bg-white/6 px-4 py-4 text-sm leading-6 text-stone-200"
            >
              {item}
            </div>
          ))}
        </div>
      </aside>
    </section>
  );
}
