import type { CSSProperties } from "react";

export type FeedTab = "following" | "recommended";
export type FanCollectionTab = "library" | "pinned";
export type ShortId = string;

export type ShortMediaAsset = {
  durationSeconds: number | null;
  id: string;
  kind: "video";
  posterUrl: string | null;
  url: string;
};

export type ShortSummary = {
  caption: string;
  canonicalMainId: string;
  creatorId: string;
  id: ShortId;
  media: ShortMediaAsset;
  previewDurationSeconds: number;
  title: string;
};

export type ShortPreviewMeta = ShortSummary;

type ShortTheme = {
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

const fallbackShortThemes = [
  {
    background: {
      accent: "#6aaac7",
      end: "#08131d",
      mid: "#254863",
      start: "#d7f6ff",
    },
    tile: {
      bottom: "#0b1c2a",
      mid: "#5db6da",
      top: "#eefaff",
    },
  },
  {
    background: {
      accent: "#73bdd6",
      end: "#07121c",
      mid: "#2d5877",
      start: "#e3f8ff",
    },
    tile: {
      bottom: "#0a1724",
      mid: "#7fb8d3",
      top: "#f6fcff",
    },
  },
  {
    background: {
      accent: "#5fa8c2",
      end: "#07131d",
      mid: "#20465e",
      start: "#d8f3ff",
    },
    tile: {
      bottom: "#0b1a27",
      mid: "#66a9c9",
      top: "#e8fbff",
    },
  },
] as const satisfies readonly ShortTheme[];

const shorts = [
  {
    caption: "雨上がりの balcony preview。続きは main で。",
    canonicalMainId: "main_sora_after_rain",
    creatorId: "sora",
    id: "afterrain",
    media: {
      durationSeconds: 17,
      id: "asset_short_sora_afterrain",
      kind: "video",
      posterUrl: "https://cdn.example.com/shorts/sora-after-rain-poster.jpg",
      url: "https://cdn.example.com/shorts/sora-after-rain.mp4",
    },
    previewDurationSeconds: 17,
    title: "after rain preview",
  },
  {
    caption: "blue tone の balcony preview。",
    canonicalMainId: "main_aoi_blue_balcony",
    creatorId: "aoi",
    id: "balcony",
    media: {
      durationSeconds: 15,
      id: "asset_short_aoi_balcony",
      kind: "video",
      posterUrl: "https://cdn.example.com/shorts/aoi-balcony-poster.jpg",
      url: "https://cdn.example.com/shorts/aoi-balcony.mp4",
    },
    previewDurationSeconds: 15,
    title: "balcony cut preview",
  },
  {
    caption: "hotel mirror の preview。",
    canonicalMainId: "main_mina_hotel_mirror",
    creatorId: "mina",
    id: "mirror",
    media: {
      durationSeconds: 18,
      id: "asset_short_mina_mirror",
      kind: "video",
      posterUrl: "https://cdn.example.com/shorts/mina-mirror-poster.jpg",
      url: "https://cdn.example.com/shorts/mina-mirror.mp4",
    },
    previewDurationSeconds: 18,
    title: "hotel mirror preview",
  },
  {
    caption: "poolside の short cut。",
    canonicalMainId: "main_sora_poolside_cut",
    creatorId: "sora",
    id: "poolcut",
    media: {
      durationSeconds: 14,
      id: "asset_short_sora_poolcut",
      kind: "video",
      posterUrl: "https://cdn.example.com/shorts/sora-poolcut-poster.jpg",
      url: "https://cdn.example.com/shorts/sora-poolcut.mp4",
    },
    previewDurationSeconds: 14,
    title: "poolside cut preview",
  },
  {
    caption: "quiet rooftop preview.",
    canonicalMainId: "main_mina_quiet_rooftop",
    creatorId: "mina",
    id: "rooftop",
    media: {
      durationSeconds: 16,
      id: "asset_short_mina_rooftop",
      kind: "video",
      posterUrl: "https://cdn.example.com/shorts/mina-rooftop-poster.jpg",
      url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
    },
    previewDurationSeconds: 16,
    title: "quiet rooftop preview",
  },
  {
    caption: "soft light の preview。",
    canonicalMainId: "main_aoi_blue_balcony",
    creatorId: "aoi",
    id: "softlight",
    media: {
      durationSeconds: 18,
      id: "asset_short_aoi_softlight",
      kind: "video",
      posterUrl: "https://cdn.example.com/shorts/aoi-softlight-poster.jpg",
      url: "https://cdn.example.com/shorts/aoi-softlight.mp4",
    },
    previewDurationSeconds: 18,
    title: "soft light preview",
  },
] as const satisfies readonly ShortSummary[];

const shortThemes: Record<string, ShortTheme> = {
  afterrain: {
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
  balcony: {
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
  mirror: {
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
  poolcut: {
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
  rooftop: {
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
  softlight: {
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
};

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
export function getShortThemeStyle(short: Pick<ShortPreviewMeta, "id"> | ShortId): CSSProperties {
  const shortId = typeof short === "string" ? short : short.id;
  const theme = shortThemes[shortId] ?? getFallbackShortTheme(shortId);

  return {
    "--short-bg-accent": theme.background.accent,
    "--short-bg-end": theme.background.end,
    "--short-bg-mid": theme.background.mid,
    "--short-bg-start": theme.background.start,
    "--short-tile-bottom": theme.tile.bottom,
    "--short-tile-mid": theme.tile.mid,
    "--short-tile-top": theme.tile.top,
  } as CSSProperties;
}

function hashShortId(shortId: string): number {
  return Array.from(shortId).reduce((accumulator, character) => accumulator + character.charCodeAt(0), 0);
}

function getFallbackShortTheme(shortId: string): ShortTheme {
  return fallbackShortThemes[hashShortId(shortId) % fallbackShortThemes.length] ?? fallbackShortThemes[0];
}
