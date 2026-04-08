import type { CreatorId } from "@/entities/creator";

export type ApprovedCreatorWorkspaceManagedTab = "main" | "shorts";

export type ApprovedCreatorWorkspaceOverviewMetrics = {
  grossUnlockRevenueJpy: number;
  uniquePurchaserCount: number;
  unlockCount: number;
};

export type ApprovedCreatorWorkspacePoster = {
  shortId: string;
  title: string;
  tile: {
    bottom: string;
    mid: string;
    top: string;
  };
};

export type ApprovedCreatorWorkspaceRevisionRequestedSummary = {
  mainCount: number;
  shortCount: number;
  totalCount: number;
};

export type ApprovedCreatorWorkspaceTopPerformer = {
  kind: ApprovedCreatorWorkspaceManagedTab;
  label: string;
  metric: string;
  shortId: string;
};

export type ApprovedCreatorWorkspaceManagedItemTone =
  | "approved"
  | "hidden"
  | "paused"
  | "pending"
  | "removed"
  | "revision";

export type ApprovedCreatorWorkspaceManagedItem = {
  detail: string;
  metric: string;
  shortId: string;
  status: string;
  title: string;
  tone: ApprovedCreatorWorkspaceManagedItemTone;
};

export type ApprovedCreatorWorkspaceDetailMetric = {
  label: string;
  value: string;
};

export type ApprovedCreatorWorkspaceDetailSetting = {
  label: string;
  value: string;
};

export type ApprovedCreatorWorkspaceDetailState = {
  durationLabel: string;
  kindLabel: string;
  linkedMainShortId: string | null;
  linkedShortIds: readonly string[];
  metrics: readonly ApprovedCreatorWorkspaceDetailMetric[];
  settings: readonly ApprovedCreatorWorkspaceDetailSetting[];
  shortId: string;
  statusLabel: string;
  statusTone: ApprovedCreatorWorkspaceManagedItemTone;
  summary: string;
};

export type ApprovedCreatorWorkspaceState = {
  detailsByTab: Readonly<Record<ApprovedCreatorWorkspaceManagedTab, Readonly<Record<string, ApprovedCreatorWorkspaceDetailState>>>>;
  managedCollections: {
    defaultTab: ApprovedCreatorWorkspaceManagedTab;
    itemsByTab: Readonly<Record<ApprovedCreatorWorkspaceManagedTab, readonly ApprovedCreatorWorkspaceManagedItem[]>>;
    tabs: readonly {
      key: ApprovedCreatorWorkspaceManagedTab;
      label: string;
    }[];
  };
  overviewMetrics: ApprovedCreatorWorkspaceOverviewMetrics;
  posters: Readonly<Record<string, ApprovedCreatorWorkspacePoster>>;
  revisionRequestedSummary: ApprovedCreatorWorkspaceRevisionRequestedSummary | null;
  topPerformers: readonly ApprovedCreatorWorkspaceTopPerformer[];
};

const dashboardTabs = [
  {
    key: "shorts",
    label: "Shorts",
  },
  {
    key: "main",
    label: "Main",
  },
] as const satisfies ApprovedCreatorWorkspaceState["managedCollections"]["tabs"];

const minaWorkspace = {
  detailsByTab: {
    main: {
      mirror: {
        durationLabel: "11分",
        kindLabel: "本編",
        linkedMainShortId: null,
        linkedShortIds: [],
        metrics: [
          { label: "paywall views", value: "1.6K" },
          { label: "unlocks", value: "143" },
          { label: "conversion", value: "8.9%" },
          { label: "revenue", value: "¥36K" },
        ],
        settings: [
          { label: "レビュー", value: "審査中" },
          { label: "価格", value: "¥2,400" },
          { label: "最終更新", value: "今日 09:18" },
        ],
        shortId: "mirror",
        statusLabel: "審査中",
        statusTone: "pending",
        summary: "paywall を開いたあとに unlock へつながる本編として再レビュー待ちです。",
      },
      rooftop: {
        durationLabel: "8分",
        kindLabel: "本編",
        linkedMainShortId: null,
        linkedShortIds: ["rooftop", "rooftopside"],
        metrics: [
          { label: "paywall views", value: "2.4K" },
          { label: "unlocks", value: "238" },
          { label: "conversion", value: "9.9%" },
          { label: "revenue", value: "¥84K" },
        ],
        settings: [
          { label: "レビュー", value: "承認済み" },
          { label: "価格", value: "¥1,800" },
          { label: "最終更新", value: "昨日 21:05" },
        ],
        shortId: "rooftop",
        statusLabel: "公開中",
        statusTone: "approved",
        summary: "linked short からの流入を unlock に変えている本編です。",
      },
    },
    shorts: {
      mirror: {
        durationLabel: "11分",
        kindLabel: "ショート",
        linkedMainShortId: "mirror",
        linkedShortIds: [],
        metrics: [
          { label: "plays", value: "94K" },
          { label: "handoff reach", value: "18K" },
          { label: "paywall opens", value: "2.8K" },
          { label: "unlocks", value: "143" },
          { label: "revenue", value: "¥36K" },
        ],
        settings: [
          { label: "レビュー", value: "再確認待ち" },
          { label: "公開状態", value: "非公開" },
          { label: "最終更新", value: "今日 09:18" },
        ],
        shortId: "mirror",
        statusLabel: "審査中",
        statusTone: "pending",
        summary: "main への handoff を調整中で、差し戻し対応が残っているショートです。",
      },
      rooftop: {
        durationLabel: "8分",
        kindLabel: "ショート",
        linkedMainShortId: "rooftop",
        linkedShortIds: [],
        metrics: [
          { label: "plays", value: "128K" },
          { label: "handoff reach", value: "31K" },
          { label: "paywall opens", value: "4.3K" },
          { label: "unlocks", value: "186" },
          { label: "revenue", value: "¥48K" },
        ],
        settings: [
          { label: "レビュー", value: "承認済み" },
          { label: "公開状態", value: "公開" },
          { label: "最終更新", value: "今日 12:24" },
        ],
        shortId: "rooftop",
        statusLabel: "公開中",
        statusTone: "approved",
        summary: "handoff と paywall open が強く、main unlock に最もつながっているショートです。",
      },
      rooftopside: {
        durationLabel: "8分",
        kindLabel: "ショート",
        linkedMainShortId: "rooftop",
        linkedShortIds: [],
        metrics: [
          { label: "plays", value: "62K" },
          { label: "handoff reach", value: "12K" },
          { label: "paywall opens", value: "1.4K" },
          { label: "unlocks", value: "52" },
          { label: "revenue", value: "¥22K" },
        ],
        settings: [
          { label: "レビュー", value: "承認済み" },
          { label: "公開状態", value: "公開" },
          { label: "最終更新", value: "昨日 19:12" },
        ],
        shortId: "rooftopside",
        statusLabel: "公開中",
        statusTone: "approved",
        summary: "同じ main に送る別導線として比較しているショートです。",
      },
    },
  },
  managedCollections: {
    defaultTab: "shorts",
    itemsByTab: {
      main: [
        {
          detail: "2 linked shorts",
          metric: "¥48K",
          shortId: "rooftop",
          status: "Approved",
          title: "quiet rooftop main",
          tone: "approved",
        },
        {
          detail: "unlock review running",
          metric: "Queue",
          shortId: "mirror",
          status: "Pending",
          title: "hotel mirror main",
          tone: "pending",
        },
      ],
      shorts: [
        {
          detail: "paywall views 1.2K",
          metric: "¥48K",
          shortId: "rooftop",
          status: "Approved",
          title: "quiet rooftop",
          tone: "approved",
        },
        {
          detail: "save rate 8.2%",
          metric: "¥22K",
          shortId: "rooftopside",
          status: "Approved",
          title: "rooftop side",
          tone: "approved",
        },
        {
          detail: "review ETA today",
          metric: "¥36K",
          shortId: "mirror",
          status: "Pending",
          title: "hotel mirror",
          tone: "pending",
        },
      ],
    },
    tabs: dashboardTabs,
  },
  overviewMetrics: {
    grossUnlockRevenueJpy: 120000,
    uniquePurchaserCount: 164,
    unlockCount: 238,
  },
  posters: {
    mirror: {
      shortId: "mirror",
      tile: {
        bottom: "#081521",
        mid: "#629bde",
        top: "#edf7ff",
      },
      title: "hotel mirror preview",
    },
    rooftop: {
      shortId: "rooftop",
      tile: {
        bottom: "#0f2234",
        mid: "#4cc0eb",
        top: "#d8f3ff",
      },
      title: "quiet rooftop preview",
    },
    rooftopside: {
      shortId: "rooftopside",
      tile: {
        bottom: "#11253a",
        mid: "#77b8e8",
        top: "#eef9ff",
      },
      title: "rooftop side preview",
    },
  },
  revisionRequestedSummary: {
    mainCount: 0,
    shortCount: 1,
    totalCount: 1,
  },
  topPerformers: [
    { kind: "main", label: "Top main", metric: "¥84K", shortId: "rooftop" },
    { kind: "shorts", label: "Top short", metric: "186 unlocks", shortId: "rooftop" },
  ],
} as const satisfies ApprovedCreatorWorkspaceState;

const workspaceByCreatorID: Readonly<Record<CreatorId, ApprovedCreatorWorkspaceState>> = {
  creator_mina_rei: minaWorkspace,
};

/**
 * approved creator dashboard の mock workspace state を返す。
 */
export function getMockApprovedCreatorWorkspaceState(creatorId: CreatorId): ApprovedCreatorWorkspaceState {
  const workspace = workspaceByCreatorID[creatorId];

  if (!workspace) {
    throw new Error(`Unknown approved creator workspace: ${creatorId}`);
  }

  return workspace;
}
