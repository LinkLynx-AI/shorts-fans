import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import type { FeedTab } from "@/entities/short";
import { viewerSessionCookieName } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { FeedShell, loadFeedShellState } from "@/widgets/feed-shell";

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
      return null;
    }
  }

  const cookieStore = await cookies();
  const sessionToken = cookieStore?.get?.(viewerSessionCookieName)?.value;
  const state = await loadFeedShellState(activeTab, {
    sessionToken,
  });

  return <FeedShell state={state} />;
}
