import { cookies } from "next/headers";

import {
  fetchFanProfileFollowingPage,
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
import { getViewerProfile } from "@/features/viewer-profile";
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
  const viewerProfileRequest = sessionToken ? { sessionToken } : {};
  let followingItems = state.followingItems;
  let libraryItems = state.libraryItems;
  let overview = state.overview;
  let pinnedItems = state.pinnedItems;
  let viewerProfile: Awaited<ReturnType<typeof getViewerProfile>> | null = null;

  try {
    switch (activeTab) {
      case "following": {
        const [nextOverview, nextViewerProfile, followingPage] = await Promise.all([
          fetchFanProfileOverview({ sessionToken }),
          getViewerProfile(viewerProfileRequest),
          fetchFanProfileFollowingPage({ sessionToken }),
        ]);

        overview = nextOverview;
        viewerProfile = nextViewerProfile;
        followingItems = followingPage.items;
        break;
      }
      case "library": {
        const [nextOverview, nextViewerProfile, libraryPage] = await Promise.all([
          fetchFanProfileOverview({ sessionToken }),
          getViewerProfile(viewerProfileRequest),
          fetchFanProfileLibraryPage({ sessionToken }),
        ]);

        overview = nextOverview;
        viewerProfile = nextViewerProfile;
        libraryItems = libraryPage.items;
        break;
      }
      case "pinned": {
        const [nextOverview, nextViewerProfile, pinnedPage] = await Promise.all([
          fetchFanProfileOverview({ sessionToken }),
          getViewerProfile(viewerProfileRequest),
          fetchFanProfilePinnedShortsPage({ sessionToken }),
        ]);

        overview = nextOverview;
        viewerProfile = nextViewerProfile;
        pinnedItems = pinnedPage.items;
        break;
      }
    }
  } catch (error) {
    if (isAuthRequiredApiError(error)) {
      return <FanAuthRequiredDialogTrigger />;
    }

    throw error;
  }

  if (viewerProfile === null) {
    throw new Error("viewer profile was not loaded for the fan hub");
  }

  return (
    <FanHubShell
      headerProfile={{
        avatarUrl: viewerProfile.avatar?.url ?? null,
        displayName: viewerProfile.displayName,
        handle: viewerProfile.handle,
      }}
      state={{ ...state, followingItems, libraryItems, overview, pinnedItems }}
    />
  );
}
