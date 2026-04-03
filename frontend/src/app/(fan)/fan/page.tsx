import { getFanHubState, normalizeFanHubTab } from "@/entities/fan-profile";
import { FanHubShell } from "@/widgets/fan-hub-shell";

export default async function FanPage({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string | string[] }>;
}) {
  const { tab } = await searchParams;

  return <FanHubShell state={getFanHubState(normalizeFanHubTab(tab))} />;
}
