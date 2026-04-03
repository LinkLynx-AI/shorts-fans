import type { CSSProperties } from "react";

export type FeedTab = "following" | "recommended";
export type FanCollectionTab = "library" | "pinned";
export type ShortId = string;

export type ShortPreviewMeta = {
  caption: string;
  creatorId: string;
  duration: string;
  id: ShortId;
  price: string;
  progress: string;
  searchLabel: string;
  theme: {
    background: {
      accent: string;
      end: string;
      mid: string;
      start: string;
    };
    tile: {
      bottom: string;
      mid: string;
      top: string;
    };
  };
  title: string;
};

const shorts = [
  {
    caption: "雨上がりの balcony preview。続きは main で。",
    creatorId: "sora",
    duration: "9分",
    id: "afterrain",
    price: "¥2,100",
    progress: "5:12 left",
    searchLabel: "after rain",
    theme: {
      background: {
        accent: "#6cbfe3",
        end: "#07131d",
        mid: "#264f70",
        start: "#cff6ff",
      },
      tile: {
        bottom: "#0f2234",
        mid: "#79c8ef",
        top: "#fff4dc",
      },
    },
    title: "after rain preview",
  },
  {
    caption: "blue tone の balcony preview。",
    creatorId: "aoi",
    duration: "10分",
    id: "balcony",
    price: "¥2,200",
    progress: "4:48 left",
    searchLabel: "blue lace set",
    theme: {
      background: {
        accent: "#7bcbe6",
        end: "#07131d",
        mid: "#2d5474",
        start: "#d9f3ff",
      },
      tile: {
        bottom: "#081521",
        mid: "#63d0d3",
        top: "#edfaff",
      },
    },
    title: "balcony cut preview",
  },
  {
    caption: "hotel mirror の preview。",
    creatorId: "mina",
    duration: "11分",
    id: "mirror",
    price: "¥2,400",
    progress: "6:26 left",
    searchLabel: "hotel mirror",
    theme: {
      background: {
        accent: "#81c7f1",
        end: "#08131d",
        mid: "#315f8d",
        start: "#d4f3ff",
      },
      tile: {
        bottom: "#081521",
        mid: "#629bde",
        top: "#edf7ff",
      },
    },
    title: "hotel mirror preview",
  },
  {
    caption: "poolside の short cut。",
    creatorId: "sora",
    duration: "8分",
    id: "poolcut",
    price: "¥1,900",
    progress: "3:55 left",
    searchLabel: "poolside cut",
    theme: {
      background: {
        accent: "#70b0d1",
        end: "#07131d",
        mid: "#233e57",
        start: "#dff9ff",
      },
      tile: {
        bottom: "#081521",
        mid: "#738aa6",
        top: "#f2feff",
      },
    },
    title: "poolside cut preview",
  },
  {
    caption: "quiet rooftop preview.",
    creatorId: "mina",
    duration: "8分",
    id: "rooftop",
    price: "¥1,800",
    progress: "8:14 left",
    searchLabel: "quiet rooftop",
    theme: {
      background: {
        accent: "#68c0eb",
        end: "#07131d",
        mid: "#2a648f",
        start: "#94e0ff",
      },
      tile: {
        bottom: "#0f2234",
        mid: "#4cc0eb",
        top: "#d8f3ff",
      },
    },
    title: "quiet rooftop preview",
  },
  {
    caption: "soft light の preview。",
    creatorId: "aoi",
    duration: "12分",
    id: "softlight",
    price: "¥2,600",
    progress: "3:42 left",
    searchLabel: "soft light",
    theme: {
      background: {
        accent: "#59a9d4",
        end: "#06111a",
        mid: "#1b4264",
        start: "#a7e8ff",
      },
      tile: {
        bottom: "#091827",
        mid: "#4f97c6",
        top: "#93e4ff",
      },
    },
    title: "soft light preview",
  },
] as const satisfies readonly ShortPreviewMeta[];

const feedShortByTab = {
  following: "softlight",
  recommended: "rooftop",
} as const satisfies Record<FeedTab, ShortId>;

const pinnedShortIds = ["afterrain", "balcony", "rooftop"] as const;
const libraryShortIds = ["softlight", "balcony", "mirror"] as const;

/**
 * mock short 一覧を取得する。
 */
export function listShorts(): readonly ShortPreviewMeta[] {
  return shorts;
}

/**
 * short ID から preview meta を取得する。
 */
export function getShortById(id: ShortId): ShortPreviewMeta | undefined {
  return shorts.find((short) => short.id === id);
}

/**
 * static route 用の short ID 一覧を取得する。
 */
export function getShortIds(): readonly ShortId[] {
  return shorts.map((short) => short.id);
}

/**
 * feed tab に対応する short を取得する。
 */
export function getFeedShortForTab(tab: FeedTab): ShortPreviewMeta {
  const short = getShortById(feedShortByTab[tab]);

  if (!short) {
    throw new Error(`Unknown feed short for tab: ${tab}`);
  }

  return short;
}

/**
 * creator ごとの short 一覧を取得する。
 */
export function getShortsByCreatorId(creatorId: string): readonly ShortPreviewMeta[] {
  return shorts.filter((short) => short.creatorId === creatorId);
}

/**
 * pinned short 一覧を取得する。
 */
export function getPinnedShorts(): readonly ShortPreviewMeta[] {
  return pinnedShortIds.flatMap((id) => {
    const short = getShortById(id);
    return short ? [short] : [];
  });
}

/**
 * library short 一覧を取得する。
 */
export function getLibraryShorts(): readonly ShortPreviewMeta[] {
  return libraryShortIds.flatMap((id) => {
    const short = getShortById(id);
    return short ? [short] : [];
  });
}

/**
 * short 背景用の CSS variable style を返す。
 */
export function getShortThemeStyle(short: ShortPreviewMeta): CSSProperties {
  return {
    "--short-bg-accent": short.theme.background.accent,
    "--short-bg-end": short.theme.background.end,
    "--short-bg-mid": short.theme.background.mid,
    "--short-bg-start": short.theme.background.start,
    "--short-tile-bottom": short.theme.tile.bottom,
    "--short-tile-mid": short.theme.tile.mid,
    "--short-tile-top": short.theme.tile.top,
  } as CSSProperties;
}
