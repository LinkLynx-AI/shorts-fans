import { cookies } from "next/headers";

import {
  fetchFanProfileLibraryPage,
  fetchFanProfilePinnedShortsPage,
  fetchFanProfileOverview,
  getFanHubState,
  normalizeFanHubTab,
} from "@/entities/fan-profile";
import { viewerSessionCookieName } from "@/entities/viewer";
import {
  FanAuthRequiredDialogTrigger,
  isAuthRequiredApiError,
} from "@/features/fan-auth";
import { FanHubShell } from "@/widgets/fan-hub-shell";

export default async function FanPage({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string | string[] }>;
}) {
  const cookieStore = await cookies();
  const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
  const { tab } = await searchParams;
  const activeTab = normalizeFanHubTab(tab);
  const state = getFanHubState(activeTab);
  let libraryItems = state.libraryItems;
  let overview;
  let pinnedItems = state.pinnedItems;

  try {
    if (activeTab === "pinned") {
      const [nextOverview, pinnedPage] = await Promise.all([
        fetchFanProfileOverview({ sessionToken }),
        fetchFanProfilePinnedShortsPage({ sessionToken }),
      ]);

      overview = nextOverview;
      pinnedItems = pinnedPage.items;
    } else {
      const [nextOverview, libraryPage] = await Promise.all([
        fetchFanProfileOverview({ sessionToken }),
        fetchFanProfileLibraryPage({ sessionToken }),
      ]);

      overview = nextOverview;
      libraryItems = libraryPage.items;
    }
  } catch (error) {
    if (isAuthRequiredApiError(error)) {
      return <FanAuthRequiredDialogTrigger />;
    }

    throw error;
  }

  return <FanHubShell state={{ ...state, libraryItems, overview, pinnedItems }} />;
}
