import { cookies } from "next/headers";

import type { FeedTab } from "@/entities/short";
import { viewerSessionCookieName } from "@/entities/viewer";
import {
  FeedShell,
  getFollowingFeedShellState,
  loadFeedShellState,
} from "@/widgets/feed-shell";

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

  const cookieStore = await cookies();
  const sessionToken = cookieStore?.get?.(viewerSessionCookieName)?.value;

  if (!sessionToken && activeTab === "following") {
    return <FeedShell state={getFollowingFeedShellState("auth_required")} />;
  }

  const state = await loadFeedShellState(activeTab, {
    sessionToken,
  });

  return <FeedShell state={state} />;
}
