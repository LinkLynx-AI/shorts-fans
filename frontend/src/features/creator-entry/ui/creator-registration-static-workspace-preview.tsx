"use client";

const previewTiles = [
  {
    detail: "approval 後に package 作成と upload が開きます。",
    label: "Upload / Intake",
    value: "Locked until approved",
  },
  {
    detail: "review queue と revision summary は onboarding 中は固定 preview です。",
    label: "Review Status",
    value: "Static mock only",
  },
  {
    detail: "analytics や unlock 指標は creator capability 解放後に表示されます。",
    label: "Analytics",
    value: "Preview after approval",
  },
] as const;

/**
 * approval 前に見せる creator workspace の static mock preview を表示する。
 */
export function CreatorRegistrationStaticWorkspacePreview() {
  return (
    <section className="mt-6 rounded-[24px] border border-[#d7e7ef] bg-[linear-gradient(180deg,#f9fcfe_0%,#f1f8fb_100%)] px-4 py-4 text-foreground">
      <p className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
        Workspace preview
      </p>
      <h2 className="mt-2 text-[18px] font-semibold tracking-[-0.02em] text-foreground">
        Approval 後に解放される creator workspace
      </h2>
      <p className="mt-2 text-sm leading-6 text-muted">
        これは操作用の dashboard ではなく、approval 後に何が解放されるかを伝える static mock です。
      </p>

      <div className="mt-4 grid gap-3">
        {previewTiles.map((tile) => (
          <div
            className="rounded-[20px] border border-white/90 bg-white/92 px-4 py-4 shadow-[0_10px_24px_rgba(15,23,42,0.05)]"
            key={tile.label}
          >
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm font-semibold tracking-[-0.02em] text-foreground">{tile.label}</p>
              <span className="rounded-full bg-[#eef7fb] px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-accent-strong">
                preview
              </span>
            </div>
            <p className="mt-2 text-[15px] font-semibold tracking-[-0.02em] text-foreground">{tile.value}</p>
            <p className="mt-1 text-sm leading-6 text-muted">{tile.detail}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
