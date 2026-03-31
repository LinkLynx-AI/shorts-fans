import { RouteStage } from "@/widgets/route-stage";

export default function ProfilePage() {
  return (
    <RouteStage
      eyebrow="profile / account"
      title="profile まわりを consumer shell の一部として先に区切る。"
      description="アカウント設定、購読一覧、視聴履歴、将来の支払い情報を profile 下に寄せる前提で route を固定します。今は account 面の枠だけを先に用意しています。"
      highlights={[
        "account / subscriptions / watch history の受け口",
        "consumer shell 配下で bottom nav と一体運用",
        "決済・通知・設定をあとから足しても route がぶれない",
      ]}
      actions={[
        { href: "/subscriptions", label: "購読状況を見る" },
        { href: "/home", label: "home に戻る", variant: "secondary" },
      ]}
    />
  );
}
