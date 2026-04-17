"use client";

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
    <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
      <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
        利用開始後の画面
      </p>
      <h2 className="mt-3 text-[22px] font-extrabold leading-tight text-foreground">
        利用開始後に使える画面
      </h2>
      <p className="mt-1 text-[13px] font-medium leading-relaxed text-muted">
        ここでは、確認が終わったあとに使える画面の雰囲気だけを先に確認できます。
      </p>

      <div className="mt-4 space-y-3">
        {previewTiles.map((tile) => (
          <div
            className="rounded-[22px] border border-gray-100 bg-[#f8f9fc] px-4 py-4"
            key={tile.label}
          >
            <p className="text-[12px] font-black tracking-[0.08em] text-[#a3adbc]">
              {tile.label}
            </p>
            <p className="mt-2 text-[16px] font-bold tracking-[-0.02em] text-foreground">
              {tile.value}
            </p>
            <p className="mt-1 text-[13px] font-medium leading-relaxed text-muted">
              {tile.detail}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
