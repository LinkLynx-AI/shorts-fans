export type CreatorId = string;

export type CreatorStat = {
  label: string;
  value: string;
};

export type CreatorSummary = {
  avatar: {
    accent: string;
    from: string;
    to: string;
  };
  bio: string;
  handle: `@${string}`;
  id: CreatorId;
  name: string;
  stats: readonly CreatorStat[];
};

const creators = [
  {
    avatar: {
      accent: "#65bae0",
      from: "#d6f5ff",
      to: "#1c4e6f",
    },
    bio: "soft light と close framing の short を中心に更新中。",
    handle: "@aoina",
    id: "aoi",
    name: "Aoi N",
    stats: [
      { label: "shorts", value: "11" },
      { label: "fans", value: "19K" },
      { label: "views", value: "132K" },
    ],
  },
  {
    avatar: {
      accent: "#7bcbe6",
      from: "#edf7ff",
      to: "#315f8d",
    },
    bio: "quiet rooftop と hotel light の preview を軸に投稿。",
    handle: "@minarei",
    id: "mina",
    name: "Mina Rei",
    stats: [
      { label: "shorts", value: "14" },
      { label: "fans", value: "24K" },
      { label: "views", value: "184K" },
    ],
  },
  {
    avatar: {
      accent: "#79c8ef",
      from: "#fff4dc",
      to: "#264f70",
    },
    bio: "after rain と balcony mood の short をまとめています。",
    handle: "@soravale",
    id: "sora",
    name: "Sora Vale",
    stats: [
      { label: "shorts", value: "9" },
      { label: "fans", value: "16K" },
      { label: "views", value: "118K" },
    ],
  },
] as const satisfies readonly CreatorSummary[];

/**
 * mock creator 一覧を取得する。
 */
export function listCreators(): readonly CreatorSummary[] {
  return creators;
}

/**
 * creator ID から summary を取得する。
 */
export function getCreatorById(id: CreatorId): CreatorSummary | undefined {
  return creators.find((creator) => creator.id === id);
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
