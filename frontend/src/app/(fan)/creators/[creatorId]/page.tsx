import { DetailShell } from "@/widgets/detail-shell";
import { RouteStructurePanel } from "@/widgets/route-structure-panel";

export default async function CreatorProfilePage({
  params,
}: {
  params: Promise<{ creatorId: string }>;
}) {
  const { creatorId } = await params;

  return (
    <DetailShell backHref="/" variant="surface">
      <h1 className="font-display text-[30px] font-semibold tracking-[-0.05em]">Creator profile structure</h1>
      <RouteStructurePanel
        description={`creator route "/creators/${creatorId}" の枠だけを先に定義しています。profile 本文は \`SHO-6\` で実装します。`}
        items={[
          {
            description: "header、stats、follow action をここに配置する",
            key: "header",
            label: "Profile header slot",
          },
          {
            description: "short grid と short detail への遷移をここに差し込む",
            key: "grid",
            label: "Short grid slot",
          },
          {
            description: "search / feed から流入した際の戻り導線をこの shell で受ける",
            key: "navigation",
            label: "Back navigation",
          },
        ]}
        title="Creator route blueprint"
      />
    </DetailShell>
  );
}
