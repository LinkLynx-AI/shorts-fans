"use client";

import {
  CreatorRegistrationSectionHeading,
  creatorRegistrationInlineSurfaceClassName,
  creatorRegistrationSectionClassName,
} from "./creator-registration-ui-primitives";

const previewTiles = [
  {
    detail: "確認が終わると、投稿の準備と管理ができるようになります。",
    label: "投稿準備",
    value: "利用開始後に表示",
  },
  {
    detail: "申請中は固定の見本だけが表示されます。",
    label: "確認状況",
    value: "申請中は見本のみ",
  },
  {
    detail: "売上や反応の数字は利用開始後に確認できます。",
    label: "反応と売上",
    value: "利用開始後に確認可能",
  },
] as const;

/**
 * 利用開始前に見せる creator workspace の見本を表示する。
 */
export function CreatorRegistrationStaticWorkspacePreview() {
  return (
    <section className={`mt-6 ${creatorRegistrationSectionClassName}`}>
      <CreatorRegistrationSectionHeading>
        利用開始後の画面
      </CreatorRegistrationSectionHeading>
      <h2 className="mt-3 font-display text-[24px] font-semibold leading-[1.12] tracking-[-0.04em] text-foreground">
        利用開始後に使える画面
      </h2>
      <p className="mt-2 text-sm leading-6 text-muted">
        ここでは、確認が終わったあとに使える画面の雰囲気だけを先に確認できます。
      </p>

      <div className="mt-4 grid gap-3">
        {previewTiles.map((tile) => (
          <div
            className={creatorRegistrationInlineSurfaceClassName}
            key={tile.label}
          >
            <p className="text-[12px] font-black tracking-[0.08em] text-[#a3adbc]">
              {tile.label}
            </p>
            <p className="mt-2 text-[16px] font-bold tracking-[-0.02em] text-foreground">
              {tile.value}
            </p>
            <p className="mt-1 text-sm leading-6 text-muted">{tile.detail}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
