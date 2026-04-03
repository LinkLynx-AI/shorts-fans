import type { FanCollectionTab } from "@/entities/short";
import { FanHubShell } from "@/widgets/fan-hub-shell";

function normalizeFanTab(tab: string | string[] | undefined): FanCollectionTab {
  return tab === "library" ? "library" : "pinned";
}

export default async function FanPage({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string | string[] }>;
}) {
  const { tab } = await searchParams;

  return <FanHubShell activeTab={normalizeFanTab(tab)} />;
}
