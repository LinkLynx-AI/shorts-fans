import { redirect } from "next/navigation";

import type { FeedTab } from "@/entities/short";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { FeedShell, getMockFeedShellState } from "@/widgets/feed-shell";

function normalizeFeedTab(tab: string | string[] | undefined): FeedTab {
  return tab === "following" ? "following" : "recommended";
}

export default async function RootPage({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string | string[] }>;
}) {
  const { tab } = await searchParams;
  const activeTab = normalizeFeedTab(tab);

  if (activeTab === "following") {
    const viewerState = await getFanAuthGateState();

    if (!viewerState.hasSession) {
      redirect(buildFanLoginHref());
    }
  }

  const state = getMockFeedShellState(activeTab);

  return <FeedShell state={state} />;
}
