import { getCreatorById, type CreatorSummary } from "@/entities/creator";
import { getMainById } from "@/entities/main";
import {
  getShortById,
  type ShortPreviewMeta,
} from "@/entities/short";

export type FanHubTab = "following" | "library" | "pinned";

export type FanProfileOverview = {
  counts: {
    following: number;
    library: number;
    pinnedShorts: number;
  };
  title: string;
};

export type FanFollowingItem = {
  creator: CreatorSummary;
  viewer: {
    isFollowing: true;
  };
};

export type FanPinnedShortItem = {
  creator: CreatorSummary;
  short: FanPinnedShortSummary;
};

export type FanPinnedShortSummary = {
  caption: string;
  canonicalMainId: string;
  creatorId: string;
  id: string;
  media: ShortPreviewMeta["media"];
  previewDurationSeconds: number;
};

export type FanLibraryItem = {
  access: {
    mainId: string;
    reason: "owner_preview" | "session_unlocked";
    status: "owner" | "unlocked";
  };
  creator: CreatorSummary;
  entryShort: ShortPreviewMeta;
  main: {
    durationSeconds: number;
    id: string;
  };
};

export type FanSettingsSection = {
  available: boolean;
  key: "account" | "payment" | "safety";
  label: string;
};

export type FanHubState = {
  activeTab: FanHubTab;
  followingItems: readonly FanFollowingItem[];
  libraryItems: readonly FanLibraryItem[];
  overview: FanProfileOverview;
  pinnedItems: readonly FanPinnedShortItem[];
};

const followingCreatorIds = ["aoi", "mina", "sora"] as const;
const pinnedShortIds = ["afterrain", "balcony", "rooftop"] as const;
const libraryShortIds = ["softlight", "mirror", "rooftop"] as const;

const settingsSections = [
  { available: true, key: "account", label: "Account" },
  { available: true, key: "payment", label: "Payment" },
  { available: true, key: "safety", label: "Safety" },
] as const satisfies readonly FanSettingsSection[];

function requireCreator(creatorId: string): CreatorSummary {
  const creator = getCreatorById(creatorId);

  if (!creator) {
    throw new Error(`Unknown creator for fan profile state: ${creatorId}`);
  }

  return creator;
}

function requireShort(shortId: string): ShortPreviewMeta {
  const short = getShortById(shortId);

  if (!short) {
    throw new Error(`Unknown short for fan profile state: ${shortId}`);
  }

  return short;
}

function requireMain(mainId: string) {
  const main = getMainById(mainId);

  if (!main) {
    throw new Error(`Unknown main for fan profile state: ${mainId}`);
  }

  return main;
}

const followingItems: readonly FanFollowingItem[] = followingCreatorIds.map((creatorId) => ({
  creator: requireCreator(creatorId),
  viewer: {
    isFollowing: true as const,
  },
}));

const pinnedItems: readonly FanPinnedShortItem[] = pinnedShortIds.map((shortId) => {
  const short = requireShort(shortId);

  return {
    creator: requireCreator(short.creatorId),
    short: {
      caption: short.caption,
      canonicalMainId: short.canonicalMainId,
      creatorId: short.creatorId,
      id: short.id,
      media: short.media,
      previewDurationSeconds: short.previewDurationSeconds,
    },
  };
});

const libraryItems: readonly FanLibraryItem[] = libraryShortIds.map((shortId) => {
  const entryShort = requireShort(shortId);
  const main = requireMain(entryShort.canonicalMainId);

  return {
    access: {
      mainId: main.id,
      reason: "session_unlocked" as const,
      status: "unlocked" as const,
    },
    creator: requireCreator(entryShort.creatorId),
    entryShort,
    main: {
      durationSeconds: main.durationSeconds,
      id: main.id,
    },
  };
});

const fanOverview = {
  counts: {
    following: followingItems.length,
    library: libraryItems.length,
    pinnedShorts: pinnedItems.length,
  },
  title: "My archive",
} as const satisfies FanProfileOverview;

/**
 * fan hub の tab 文字列を正規化する。
 */
export function normalizeFanHubTab(tab: string | string[] | undefined): FanHubTab {
  if (tab === "following") {
    return "following";
  }

  if (tab === "library") {
    return "library";
  }

  return "pinned";
}

/**
 * fan profile overview を取得する。
 */
export function getFanProfileOverview(): FanProfileOverview {
  return fanOverview;
}

/**
 * fan hub 用の mock state を取得する。
 */
export function getFanHubState(activeTab: FanHubTab): FanHubState {
  return {
    activeTab,
    followingItems,
    libraryItems,
    overview: fanOverview,
    pinnedItems,
  };
}

/**
 * following 一覧を取得する。
 */
export function listFollowingItems(): readonly FanFollowingItem[] {
  return followingItems;
}

/**
 * settings section 一覧を取得する。
 */
export function listFanSettingsSections(): readonly FanSettingsSection[] {
  return settingsSections;
}
