export type CreatorId = string;

type CreatorAvatarAsset = {
  durationSeconds: null;
  id: string;
  kind: "image";
  posterUrl: null;
  url: string;
};

export type CreatorSummary = {
  avatar: CreatorAvatarAsset;
  bio: string;
  displayName: string;
  handle: `@${string}`;
  id: CreatorId;
};

export type CreatorProfileStats = {
  fanCount: number;
  shortCount: number;
  viewCount: number;
};

const recentCreatorIds = ["aoi", "mina"] as const satisfies readonly CreatorId[];

function createAvatarDataUrl(from: string, accent: string, to: string): string {
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 80 80" fill="none">
      <defs>
        <linearGradient id="avatar-gradient" x1="40" y1="0" x2="40" y2="80" gradientUnits="userSpaceOnUse">
          <stop stop-color="${from}" />
          <stop offset="0.56" stop-color="${accent}" />
          <stop offset="1" stop-color="${to}" />
        </linearGradient>
      </defs>
      <rect width="80" height="80" rx="40" fill="url(#avatar-gradient)" />
    </svg>
  `.trim();

  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
}

function createAvatarAsset(creatorId: CreatorId, from: string, accent: string, to: string): CreatorAvatarAsset {
  return {
    durationSeconds: null,
    id: `asset_creator_${creatorId}_avatar`,
    kind: "image",
    posterUrl: null,
    url: createAvatarDataUrl(from, accent, to),
  };
}

const creators = [
  {
    avatar: createAvatarAsset("aoi", "#d6f5ff", "#65bae0", "#1c4e6f"),
    bio: "soft light と close framing の short を中心に更新中。",
    displayName: "Aoi N",
    handle: "@aoina",
    id: "aoi",
  },
  {
    avatar: createAvatarAsset("mina", "#edf7ff", "#7bcbe6", "#315f8d"),
    bio: "quiet rooftop と hotel light の preview を軸に投稿。",
    displayName: "Mina Rei",
    handle: "@minarei",
    id: "mina",
  },
  {
    avatar: createAvatarAsset("sora", "#fff4dc", "#79c8ef", "#264f70"),
    bio: "after rain と balcony mood の short をまとめています。",
    displayName: "Sora Vale",
    handle: "@soravale",
    id: "sora",
  },
] as const satisfies readonly CreatorSummary[];

const creatorProfileStatsById: Record<string, CreatorProfileStats> = {
  aoi: { fanCount: 19000, shortCount: 11, viewCount: 132000 },
  mina: { fanCount: 24000, shortCount: 2, viewCount: 184000 },
  sora: { fanCount: 16000, shortCount: 0, viewCount: 118000 },
};

function normalizeCreatorSearchQuery(query: string): string {
  return query.trim().toLowerCase();
}

function getCreatorSearchText(creator: CreatorSummary): string {
  return `${creator.displayName} ${creator.handle}`.toLowerCase();
}

/**
 * mock creator 一覧を取得する。
 */
export function listCreators(): readonly CreatorSummary[] {
  return creators;
}

/**
 * creator search 初期表示用の recent creators を取得する。
 */
export function getRecentCreators(): readonly CreatorSummary[] {
  return recentCreatorIds.flatMap((id) => {
    const creator = getCreatorById(id);

    return creator ? [creator] : [];
  });
}

/**
 * creator ID から summary を取得する。
 */
export function getCreatorById(id: CreatorId): CreatorSummary | undefined {
  return creators.find((creator) => creator.id === id);
}

/**
 * display name / handle のみを対象に creator を検索する。
 */
export function searchCreators(query: string): readonly CreatorSummary[] {
  const normalizedQuery = normalizeCreatorSearchQuery(query);

  if (normalizedQuery.length === 0) {
    return getRecentCreators();
  }

  return creators.filter((creator) => getCreatorSearchText(creator).includes(normalizedQuery));
}

/**
 * creator profile 用の stats を取得する。
 */
export function getCreatorProfileStatsById(id: CreatorId): CreatorProfileStats | undefined {
  return creatorProfileStatsById[id];
}

/**
 * static route 用の creator ID 一覧を取得する。
 */
export function getCreatorIds(): readonly CreatorId[] {
  return creators.map((creator) => creator.id);
}

/**
 * avatar fallback 用の initials を返す。
 */
export function getCreatorInitials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}
