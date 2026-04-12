import { getCreatorById, type CreatorSummary } from "@/entities/creator";
import {
  getShortById,
  type ShortPreviewMeta,
} from "@/entities/short";

export type FanHubTab = "library" | "pinned";

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
  libraryItems: readonly FanLibraryItem[];
  overview: FanProfileOverview;
  pinnedItems: readonly FanPinnedShortItem[];
};

const followingCreatorIds = ["aoi", "mina", "sora"] as const;
const pinnedShortIds = ["afterrain", "balcony", "rooftop"] as const;
const libraryDefinitions = [
  {
    main: {
      durationSeconds: 720,
      id: "main_aoi_soft_light",
    },
    shortId: "softlight",
  },
  {
    main: {
      durationSeconds: 600,
      id: "main_aoi_balcony_cut",
    },
    shortId: "balcony",
  },
  {
    main: {
      durationSeconds: 660,
      id: "main_mina_hotel_mirror",
    },
    shortId: "mirror",
  },
] as const;

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

const libraryItems: readonly FanLibraryItem[] = libraryDefinitions.map((definition) => {
  const entryShort = requireShort(definition.shortId);

  return {
    access: {
      mainId: definition.main.id,
      reason: "session_unlocked" as const,
      status: "unlocked" as const,
    },
    creator: requireCreator(entryShort.creatorId),
    entryShort,
    main: definition.main,
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
  return tab === "library" ? "library" : "pinned";
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
