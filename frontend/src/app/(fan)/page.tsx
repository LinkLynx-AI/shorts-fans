import { getFeedShortForTab, type FeedTab } from "@/entities/short";
import { FeedShell } from "@/widgets/feed-shell";

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
  const short = getFeedShortForTab(activeTab);

  return <FeedShell activeTab={activeTab} short={short} />;
}
