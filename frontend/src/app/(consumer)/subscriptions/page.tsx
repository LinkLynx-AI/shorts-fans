import { RouteStage } from "@/widgets/route-stage";

export default function SubscriptionsPage() {
  return (
    <RouteStage
      eyebrow="subscriptions / retention"
      title="購読後の継続視聴面を独立 route として先に確保する。"
      description="公開 short と限定動画は役割が違うので、継続視聴の面は最初から独立 route にします。ここでは購読 creator の更新と追いかけ直しの箱を先に置いています。"
      highlights={[
        "購読中 creator の更新確認に特化した面として固定",
        "限定動画の視聴継続導線を shorts feed と分離",
        "将来の課金継続率改善を考えて route を独立維持",
      ]}
      actions={[
        { href: "/profile", label: "profile を確認" },
        { href: "/", label: "shorts に戻る", variant: "secondary" },
      ]}
    />
  );
}
