import type { FeedTab } from "@/entities/short";
import { getOptionalClientEnv } from "@/shared/config";
import { FeedShell, fetchRecommendedFeedShellState, getMockFeedShellState } from "@/widgets/feed-shell";

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
  const { NEXT_PUBLIC_API_BASE_URL: apiBaseUrl } = getOptionalClientEnv();
  const state =
    activeTab === "recommended" && apiBaseUrl
      ? await fetchRecommendedFeedShellState({
          baseUrl: apiBaseUrl,
        })
      : getMockFeedShellState(activeTab);

  return <FeedShell state={state} />;
}
