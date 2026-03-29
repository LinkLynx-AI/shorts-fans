import { RouteStage } from "@/widgets/route-stage";

export default function ShortsPage() {
  return (
    <RouteStage
      eyebrow="shorts / primary lane"
      title="縦型 feed をこのアプリの主導線として固定する。"
      description="本番では連続視聴、縦型アクション、creator 遷移をここに積み上げます。今は shell と route を先に確保し、以後の UI 実装がここを中心に伸びる状態にしています。"
      highlights={[
        "mobile bottom nav の既定体験は shorts を中心に設計",
        "公開 short は非購読でも見られる前提で情報密度を抑制",
        "creator 遷移・subscription 導線の受け口をここから伸ばす",
      ]}
      actions={[
        { href: "/creator/atelier-rin", label: "creator へ移動" },
        { href: "/subscriptions", label: "購読 feed を確認", variant: "secondary" },
      ]}
    />
  );
}
