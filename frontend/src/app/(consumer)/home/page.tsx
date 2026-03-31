import { RouteStage } from "@/widgets/route-stage";

export default function HomePage() {
  return (
    <RouteStage
      eyebrow="home / discovery hub"
      title="好みの creator を掘るための入口を先に固定する。"
      description="`/home` は discovery hub として残しつつ、サイト初期表示は `/` の shorts に合わせます。発見導線、ジャンル入口、creator 回遊の箱は home 側に集約し、初手視聴は shorts に寄せます。"
      highlights={[
        "genre / creator / new arrivals の3系統で発見導線を分離",
        "初手流入は `/`、探索は `/home` へ切り出して役割分担",
        "creator 詳細へ移る前の discovery hub として home の責務を固定",
      ]}
      actions={[
        { href: "/", label: "shorts を開く" },
        { href: "/creator/atelier-rin", label: "creator sample を見る", variant: "secondary" },
      ]}
    >
      <section className="grid gap-4 md:grid-cols-3">
        {[
          ["new", "新着 short から入る", "毎日の更新確認を最短にする棚。"],
          ["genre", "ジャンル別に潜る", "home 側で discover 導線を持たせる前提。"],
          ["creator", "気になる creator を追う", "creator page と subscription 導線へ接続。"],
        ].map(([label, title, copy]) => (
          <article
            key={label}
            className="rounded-[1.5rem] border border-white/80 bg-white/80 p-5 shadow-[0_12px_30px_rgba(87,38,8,0.08)]"
          >
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-accent">{label}</p>
            <h2 className="mt-3 text-lg font-semibold text-foreground">{title}</h2>
            <p className="mt-2 text-sm leading-6 text-muted">{copy}</p>
          </article>
        ))}
      </section>
    </RouteStage>
  );
}
