const creators = {
  aoi: {
    id: "aoi",
    name: "Aoi N",
    handle: "@aoina",
    bio: "soft light と close framing の short を中心に更新中。",
    stats: {
      fans: "19K",
      shorts: "11",
      views: "132K",
    },
  },
  mina: {
    id: "mina",
    name: "Mina Rei",
    handle: "@minarei",
    bio: "quiet rooftop と hotel light の preview を軸に投稿。",
    stats: {
      fans: "24K",
      shorts: "14",
      views: "184K",
    },
  },
  sora: {
    id: "sora",
    name: "Sora Vale",
    handle: "@soravale",
    bio: "after rain と balcony mood の short をまとめています。",
    stats: {
      fans: "16K",
      shorts: "9",
      views: "118K",
    },
  },
};

const shorts = {
  afterrain: {
    id: "afterrain",
    creatorId: "sora",
    duration: "9分",
    price: "¥2,100",
    progress: "5:12 left",
    searchLabel: "after rain",
    theme: "afterrain",
    title: "after rain preview",
    caption: "雨上がりの balcony preview。続きは main で。",
    tile: {
      bottom: "#0f2234",
      mid: "#79c8ef",
      top: "#fff4dc",
    },
  },
  balcony: {
    id: "balcony",
    creatorId: "aoi",
    duration: "10分",
    price: "¥2,200",
    progress: "4:48 left",
    searchLabel: "blue lace set",
    theme: "balcony",
    title: "balcony cut preview",
    caption: "blue tone の balcony preview。",
    tile: {
      bottom: "#081521",
      mid: "#63d0d3",
      top: "#edfaff",
    },
  },
  mirror: {
    id: "mirror",
    creatorId: "mina",
    duration: "11分",
    price: "¥2,400",
    progress: "6:26 left",
    searchLabel: "hotel mirror",
    theme: "mirror",
    title: "hotel mirror preview",
    caption: "hotel mirror の preview。",
    tile: {
      bottom: "#081521",
      mid: "#629bde",
      top: "#edf7ff",
    },
  },
  rooftopside: {
    id: "rooftopside",
    creatorId: "mina",
    duration: "8分",
    price: "¥1,800",
    progress: "6:40 left",
    searchLabel: "rooftop side light",
    theme: "rooftop",
    title: "rooftop side preview",
    caption: "quiet rooftop の別導線 preview。",
    tile: {
      bottom: "#11253a",
      mid: "#77b8e8",
      top: "#eef9ff",
    },
  },
  poolcut: {
    id: "poolcut",
    creatorId: "sora",
    duration: "8分",
    price: "¥1,900",
    progress: "3:55 left",
    searchLabel: "poolside cut",
    theme: "poolcut",
    title: "poolside cut preview",
    caption: "poolside の short cut。",
    tile: {
      bottom: "#081521",
      mid: "#738aa6",
      top: "#f2feff",
    },
  },
  rooftop: {
    id: "rooftop",
    creatorId: "mina",
    duration: "8分",
    price: "¥1,800",
    progress: "8:14 left",
    searchLabel: "quiet rooftop",
    theme: "rooftop",
    title: "quiet rooftop preview",
    caption: "quiet rooftop preview.",
    tile: {
      bottom: "#0f2234",
      mid: "#4cc0eb",
      top: "#d8f3ff",
    },
  },
  softlight: {
    id: "softlight",
    creatorId: "aoi",
    duration: "12分",
    price: "¥2,600",
    progress: "3:42 left",
    searchLabel: "soft light",
    theme: "softlight",
    title: "soft light preview",
    caption: "soft light の preview。",
    tile: {
      bottom: "#091827",
      mid: "#4f97c6",
      top: "#93e4ff",
    },
  },
};

const creatorShorts = {
  aoi: ["softlight", "balcony"],
  mina: ["rooftop", "rooftopside", "mirror"],
  sora: ["afterrain", "poolcut"],
};

const feedShortByTab = {
  following: "softlight",
  recommended: "rooftop",
};

const viewerCreatorId = "mina";

const creatorDashboardData = {
  mina: {
    description: "quiet rooftop 系の release をまとめて管理し、review と unlock の動きを profile から見返します。",
    profileStats: [
      { label: "shorts", value: "14" },
      { label: "unlocks", value: "238" },
      { label: "followers", value: "24K" },
    ],
    revisionNotice: {
      detail: "キャプション連携を確認してください",
      label: "差し戻しが1件あります",
    },
    mains: [
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
        revenue: "¥48K",
        shortId: "rooftop",
        status: "Approved",
        title: "quiet rooftop",
        tone: "approved",
      },
      {
        detail: "save rate 8.2%",
        revenue: "¥22K",
        shortId: "rooftopside",
        status: "Approved",
        title: "rooftop side",
        tone: "approved",
      },
      {
        detail: "review ETA today",
        revenue: "¥36K",
        shortId: "mirror",
        status: "Pending",
        title: "hotel mirror",
        tone: "pending",
      },
    ],
  },
};

const creatorManagerDetailData = {
  main: {
    mirror: {
      kindLabel: "本編",
      metrics: [
        { label: "視聴回数", value: "1.6K" },
        { label: "unlock数", value: "143" },
        { label: "売上", value: "¥36K" },
      ],
      settings: [
        { label: "レビュー", value: "審査中" },
        { label: "価格", value: "¥2,400" },
        { label: "紐づくショート", value: "1本" },
        { label: "最終更新", value: "今日 09:18" },
      ],
      statusLabel: "審査中",
      statusTone: "pending",
      summary: "差し戻し対応を反映した後の再レビュー待ちです。",
    },
    rooftop: {
      kindLabel: "本編",
      metrics: [
        { label: "視聴回数", value: "2.4K" },
        { label: "unlock数", value: "238" },
        { label: "売上", value: "¥84K" },
      ],
      settings: [
        { label: "レビュー", value: "承認済み" },
        { label: "価格", value: "¥1,800" },
        { label: "紐づくショート", value: "2本" },
        { label: "最終更新", value: "昨日 21:05" },
      ],
      statusLabel: "公開中",
      statusTone: "approved",
      summary: "short から unlock されたあとに再生される main です。",
    },
  },
  shorts: {
    mirror: {
      kindLabel: "ショート",
      linkedPublicShortCount: 1,
      mainKey: "mirror",
      metrics: [
        { label: "視聴回数", value: "94K" },
        { label: "unlock数", value: "143" },
        { label: "売上", value: "¥36K" },
      ],
      settings: [
        { label: "レビュー", value: "再確認待ち" },
        { label: "公開範囲", value: "プロフィールのみ" },
        { action: "open-linked-main", label: "リンク先本編", mainKey: "mirror", value: "本編へ" },
        { label: "最終更新", value: "今日 09:18" },
      ],
      statusLabel: "審査中",
      statusTone: "pending",
      summary: "差し戻し対応が必要なショートです。",
    },
    rooftop: {
      kindLabel: "ショート",
      linkedPublicShortCount: 2,
      mainKey: "rooftop",
      metrics: [
        { label: "視聴回数", value: "128K" },
        { label: "unlock数", value: "186" },
        { label: "売上", value: "¥48K" },
      ],
      settings: [
        { label: "レビュー", value: "承認済み" },
        { label: "公開範囲", value: "プロフィール / feed" },
        { action: "open-linked-main", label: "リンク先本編", mainKey: "rooftop", value: "本編へ" },
        { label: "最終更新", value: "今日 12:24" },
      ],
      statusLabel: "公開中",
      statusTone: "approved",
      summary: "プロフィールの先頭で使っている導入ショートです。",
    },
    rooftopside: {
      kindLabel: "ショート",
      linkedPublicShortCount: 2,
      mainKey: "rooftop",
      metrics: [
        { label: "視聴回数", value: "62K" },
        { label: "unlock数", value: "52" },
        { label: "売上", value: "¥22K" },
      ],
      settings: [
        { label: "レビュー", value: "承認済み" },
        { label: "公開範囲", value: "プロフィール / feed" },
        { action: "open-linked-main", label: "リンク先本編", mainKey: "rooftop", value: "本編へ" },
        { label: "最終更新", value: "昨日 19:12" },
      ],
      statusLabel: "公開中",
      statusTone: "approved",
      summary: "同じ rooftop main に流す別カットのショートです。",
    },
  },
};

const state = {
  acceptAge: false,
  acceptTerms: false,
  currentCreatorId: "mina",
  creatorManagerDetailShortId: "rooftop",
  creatorManagerDetailTab: "shorts",
  creatorManagerTab: "shorts",
  creatorMainFlowState: {
    mirror: "active",
    rooftop: "active",
  },
  creatorPendingAction: null,
  creatorShortVisibilityState: {
    mirror: "live",
    rooftop: "live",
    rooftopside: "live",
  },
  creatorUploadMode: "new-package",
  creatorUploadMainName: "",
  creatorUploadShortNames: [""],
  creatorUploadTargetMainId: null,
  currentShortId: "rooftop",
  fanTab: "pinned",
  feedTab: "recommended",
  followingQuery: "",
  followingCreatorIds: new Set(["aoi", "mina", "sora"]),
  hasPurchaseSetup: false,
  history: [],
  libraryShortIds: new Set(["softlight", "balcony", "mirror"]),
  libraryQuery: "",
  lastMainShortId: "softlight",
  overlay: null,
  overlayShortId: null,
  pinnedShortIds: new Set(["afterrain", "balcony", "rooftop"]),
  purchasedShortIds: new Set(["softlight", "balcony", "mirror"]),
  rootTab: "feed",
  screen: "feed",
  searchQuery: "",
};

const root = document.getElementById("app");

render();

document.addEventListener("click", (event) => {
  const actionButton = event.target.closest("[data-action]");

  if (!actionButton) {
    return;
  }

  const { action, creatorId, shortId, tab, kind, index } = actionButton.dataset;

  if (action === "back") {
    handleBack();
    return;
  }

  if (action === "close-overlay") {
    closeOverlay();
    return;
  }

  if (action === "confirm-paywall") {
    if (state.acceptAge && state.acceptTerms && state.overlayShortId) {
      unlockShort(state.overlayShortId);
    }
    return;
  }

  if (action === "confirm-creator-short-action") {
    confirmCreatorShortAction();
    return;
  }

  if (action === "confirm-creator-main-action") {
    confirmCreatorMainAction();
    return;
  }

  if (action === "open-creator-post-actions") {
    state.overlay = "creator-post-actions";
    render();
    return;
  }

  if (action === "open-creator" && creatorId) {
    navigate({
      currentCreatorId: creatorId,
      currentShortId: creatorShorts[creatorId][0],
      screen: "creator",
    });
    return;
  }

  if (action === "open-short" && shortId) {
    navigate({
      currentCreatorId: shorts[shortId].creatorId,
      currentShortId: shortId,
      screen: "short",
    });
    return;
  }

  if (action === "open-creator-dashboard") {
    navigate({
      creatorManagerDetailShortId: creatorShorts[viewerCreatorId][0],
      creatorManagerDetailTab: "shorts",
      creatorManagerTab: "shorts",
      currentCreatorId: viewerCreatorId,
      currentShortId: creatorShorts[viewerCreatorId][0],
      rootTab: "fan",
      screen: "creator-dashboard",
    });
    return;
  }

  if (action === "open-creator-upload") {
    navigate({
      creatorUploadMode: "new-package",
      creatorUploadMainName: "",
      creatorUploadShortNames: [""],
      creatorUploadTargetMainId: null,
      currentCreatorId: viewerCreatorId,
      currentShortId: creatorShorts[viewerCreatorId][0],
      rootTab: "fan",
      screen: "creator-upload",
    });
    return;
  }

  if (action === "open-creator-linked-short-upload") {
    const targetMainId = state.creatorManagerDetailShortId || state.currentShortId;

    navigate({
      creatorUploadMode: "link-short",
      creatorUploadMainName: "",
      creatorUploadShortNames: [""],
      creatorUploadTargetMainId: targetMainId,
      currentCreatorId: viewerCreatorId,
      currentShortId: targetMainId,
      rootTab: "fan",
      screen: "creator-upload",
    });
    return;
  }

  if (action === "set-creator-manager-tab" && tab) {
    state.creatorManagerTab = tab === "main" ? "main" : "shorts";
    render();
    return;
  }

  if (action === "open-creator-manager-detail" && kind && shortId) {
    navigate({
      creatorManagerDetailShortId: shortId,
      creatorManagerDetailTab: kind === "main" ? "main" : "shorts",
      currentCreatorId: viewerCreatorId,
      currentShortId: shortId,
      rootTab: "fan",
      screen: "creator-post-detail",
    });
    return;
  }

  if (action === "open-creator-linked-main" && shortId) {
    navigate({
      creatorManagerDetailShortId: shortId,
      creatorManagerDetailTab: "main",
      currentCreatorId: viewerCreatorId,
      currentShortId: shortId,
      rootTab: "fan",
      screen: "creator-post-detail",
    });
    return;
  }

  if (action === "request-creator-short-action" && kind) {
    requestCreatorShortAction(kind);
    return;
  }

  if (action === "request-creator-main-action" && kind) {
    requestCreatorMainAction(kind);
    return;
  }

  if (action === "pick-creator-upload-files" && kind) {
    const inputSelector =
      kind === "shorts" && typeof index !== "undefined"
        ? `[data-role="creator-upload-input"][data-kind="${kind}"][data-index="${index}"]`
        : `[data-role="creator-upload-input"][data-kind="${kind}"]`;
    const uploadInput = root.querySelector(inputSelector);

    if (uploadInput instanceof HTMLInputElement) {
      uploadInput.click();
    }

    return;
  }

  if (action === "add-creator-upload-short-slot") {
    state.creatorUploadShortNames = [...state.creatorUploadShortNames, ""];
    render();
    return;
  }

  if (action === "remove-creator-upload-short-slot" && typeof index !== "undefined") {
    const slotIndex = Number(index);

    if (Number.isInteger(slotIndex) && slotIndex >= 0) {
      state.creatorUploadShortNames = state.creatorUploadShortNames.filter((_, currentIndex) => currentIndex !== slotIndex);
      render();
    }

    return;
  }

  if (action === "submit-creator-upload-package") {
    if (!isCreatorUploadReady()) {
      return;
    }

    const uploadMode = state.creatorUploadMode;
    const targetMainId = state.creatorUploadTargetMainId;

    resetCreatorUploadDraft();

    if (uploadMode === "link-short" && targetMainId) {
      state.history.pop();
      navigate(
        {
          creatorManagerDetailShortId: targetMainId,
          creatorManagerDetailTab: "main",
          creatorManagerTab: "main",
          currentCreatorId: viewerCreatorId,
          currentShortId: targetMainId,
          rootTab: "fan",
          screen: "creator-post-detail",
        },
        true,
      );
      return;
    }

    navigate(
      {
        creatorManagerTab: "shorts",
        currentCreatorId: viewerCreatorId,
        currentShortId: creatorShorts[viewerCreatorId][0],
        rootTab: "fan",
        screen: "creator-dashboard",
      },
      true,
    );
    return;
  }

  if (action === "open-paywall" && shortId) {
    openPaywall(shortId);
    return;
  }

  if (action === "open-main" && shortId) {
    openMain(shortId);
    return;
  }

  if (action === "open-library") {
    openLibrary();
    return;
  }

  if (action === "open-following") {
    openFollowing();
    return;
  }

  if (action === "open-pinned") {
    openPinned();
    return;
  }

  if (action === "open-search") {
    state.searchQuery = "";
    switchPrimaryTab("search");
    return;
  }

  if (action === "open-fan") {
    switchPrimaryTab("fan");
    return;
  }

  if (action === "open-feed") {
    switchPrimaryTab("feed");
    return;
  }

  if (action === "set-fan-tab" && tab) {
    state.fanTab = tab;
    render();
    return;
  }

  if (action === "set-tab" && tab) {
    state.feedTab = tab;
    state.currentShortId = feedShortByTab[tab];
    state.currentCreatorId = shorts[state.currentShortId].creatorId;
    render();
    return;
  }

  if (action === "toggle-follow" && creatorId) {
    toggleSet(state.followingCreatorIds, creatorId);
    render();
    return;
  }

  if (action === "remove-follow" && creatorId) {
    state.followingCreatorIds.delete(creatorId);
    render();
    return;
  }

  if (action === "toggle-pin" && shortId) {
    toggleSet(state.pinnedShortIds, shortId);
    render();
    return;
  }

});

document.addEventListener("input", (event) => {
  const target = event.target;

  if (target.matches("[data-role='search-input']")) {
    state.searchQuery = target.value;
    updateSearchResults();
    return;
  }

  if (target.matches("[data-role='library-search-input']")) {
    state.libraryQuery = target.value;
    updateLibraryResults();
    return;
  }

  if (target.matches("[data-role='following-search-input']")) {
    state.followingQuery = target.value;
    updateFollowingResults();
    return;
  }

  if (target.matches("[data-role='accept-age']")) {
    state.acceptAge = target.checked;
    updatePaywallState();
    return;
  }

  if (target.matches("[data-role='accept-terms']")) {
    state.acceptTerms = target.checked;
    updatePaywallState();
  }
});

document.addEventListener("change", (event) => {
  const target = event.target;

  if (target instanceof HTMLInputElement && target.matches("[data-role='creator-upload-input']") && target.files?.length) {
    const fileNames = Array.from(target.files).map((file) => file.name);

    if (target.dataset.kind === "main") {
      state.creatorUploadMainName = fileNames[0] || "";
    }

    if (target.dataset.kind === "shorts") {
      const slotIndex = Number(target.dataset.index);

      if (Number.isInteger(slotIndex) && slotIndex >= 0) {
        state.creatorUploadShortNames = state.creatorUploadShortNames.map((fileName, currentIndex) =>
          currentIndex === slotIndex ? fileNames[0] || "" : fileName,
        );
      }
    }

    render();
  }
});

function closeOverlay() {
  state.acceptAge = false;
  state.creatorPendingAction = null;
  state.acceptTerms = false;
  state.overlay = null;
  state.overlayShortId = null;
  state.searchQuery = "";
  render();
}

function createBackAction() {
  return `<button aria-label="Back" class="back-button" data-action="back" type="button">&lt;</button>`;
}

function createCreatorAddAction() {
  return `<button aria-label="動画を追加" class="creator-topbar-add-button" data-action="open-creator-upload" type="button">+</button>`;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function resetCreatorUploadDraft() {
  state.creatorUploadMode = "new-package";
  state.creatorUploadMainName = "";
  state.creatorUploadShortNames = [""];
  state.creatorUploadTargetMainId = null;
}

function isCreatorUploadReady() {
  const hasShorts = state.creatorUploadShortNames.some(Boolean);

  if (state.creatorUploadMode === "link-short") {
    return hasShorts;
  }

  return Boolean(state.creatorUploadMainName && hasShorts);
}

function currentShort() {
  return shorts[state.currentShortId];
}

function currentCreatorMainState(mainKey) {
  return state.creatorMainFlowState[mainKey] || "active";
}

function currentCreatorShortVisibility(shortId) {
  return state.creatorShortVisibilityState[shortId] || "live";
}

function linkedShortIdsForMain(mainKey) {
  return Object.entries(creatorManagerDetailData.shorts)
    .filter(([, detail]) => detail.mainKey === mainKey)
    .map(([shortId]) => shortId);
}

function linkedLiveShortCountForMain(mainKey) {
  return linkedShortIdsForMain(mainKey).filter((shortId) => currentCreatorShortVisibility(shortId) === "live").length;
}

function visibilityStateForAction(actionKind) {
  return actionKind === "delete" ? "deleted" : "hidden";
}

function requestCreatorShortAction(actionKind) {
  const detailTab = state.creatorManagerDetailTab === "main" ? "main" : "shorts";
  const detailShortId = state.creatorManagerDetailShortId || state.currentShortId;

  if (detailTab !== "shorts") {
    return;
  }

  if (currentCreatorShortVisibility(detailShortId) !== "live") {
    return;
  }

  const detail = creatorManagerDetailData.shorts[detailShortId];

  if (!detail) {
    return;
  }

  state.creatorPendingAction = {
    actionKind: actionKind === "delete" ? "delete" : "hide",
    mainKey: detail.mainKey || detailShortId,
    shortId: detailShortId,
    willStopMainFlow: linkedLiveShortCountForMain(detail.mainKey || detailShortId) <= 1,
  };
  state.overlay = "creator-short-action";
  render();
}

function requestCreatorMainAction(actionKind) {
  const detailTab = state.creatorManagerDetailTab === "main" ? "main" : "shorts";
  const mainKey = state.creatorManagerDetailShortId || state.currentShortId;
  const currentState = currentCreatorMainState(mainKey);

  if (detailTab !== "main") {
    return;
  }

  if (currentState !== "active" && currentState !== "paused") {
    return;
  }

  state.creatorPendingAction = {
    actionKind: actionKind === "delete" ? "delete" : "hide",
    linkedShortIds: linkedShortIdsForMain(mainKey),
    mainKey,
    targetType: "main",
  };
  state.overlay = "creator-main-action";
  render();
}

function confirmCreatorShortAction() {
  const pendingAction = state.creatorPendingAction;

  if (!pendingAction || pendingAction.targetType === "main") {
    return;
  }

  state.creatorShortVisibilityState[pendingAction.shortId] = visibilityStateForAction(pendingAction.actionKind);

  if (pendingAction.willStopMainFlow) {
    state.creatorMainFlowState[pendingAction.mainKey] = "paused";
  }

  state.creatorPendingAction = null;
  state.overlay = null;
  render();
}

function confirmCreatorMainAction() {
  const pendingAction = state.creatorPendingAction;

  if (!pendingAction || pendingAction.targetType !== "main") {
    return;
  }

  const nextVisibilityState = visibilityStateForAction(pendingAction.actionKind);

  state.creatorMainFlowState[pendingAction.mainKey] = nextVisibilityState;
  pendingAction.linkedShortIds.forEach((shortId) => {
    state.creatorShortVisibilityState[shortId] = nextVisibilityState;
  });

  state.creatorPendingAction = null;
  state.overlay = null;
  render();
}

function currentTheme() {
  if (state.screen === "creator") {
    return "profile";
  }

  if (state.screen === "library") {
    return "profile";
  }

  if (state.screen === "fan") {
    return "fan";
  }

  return currentShort().theme;
}

function creatorCard(creatorId) {
  const creator = creators[creatorId];

  return `
    <button class="creator-link" data-action="open-creator" data-creator-id="${creator.id}" type="button">
      <span class="creator-avatar"></span>
      <span>
        <p class="creator-name">${creator.name}</p>
        <p class="creator-handle">${creator.handle}</p>
      </span>
    </button>
  `;
}

function feedCreatorMeta(shortId, { clickableCaption = true } = {}) {
  const short = shorts[shortId];
  const creator = creators[short.creatorId];
  const following = state.followingCreatorIds.has(creator.id);
  const captionMarkup = clickableCaption
    ? `<button class="caption-button" data-action="open-short" data-short-id="${short.id}" type="button">${short.caption}</button>`
    : `<p class="caption-button">${short.caption}</p>`;

  return `
    <div class="creator-card">
      <div class="feed-creator-row">
        <button class="creator-link" data-action="open-creator" data-creator-id="${creator.id}" type="button">
          <span class="creator-avatar"></span>
          <span class="creator-name">${creator.name}</span>
        </button>
        <button class="feed-follow-button ${following ? "is-active" : ""}" data-action="toggle-follow" data-creator-id="${creator.id}" type="button">
          ${following ? "Following" : "Follow"}
        </button>
      </div>
      ${captionMarkup}
    </div>
  `;
}

function creatorRows(creatorIds, trailingLabel = "削除") {
  return creatorIds
    .map((creatorId) => {
      const creator = creators[creatorId];

      return `
        <div class="follow-row" data-following-item data-search-text="${`${creator.name} ${creator.handle}`.toLowerCase()}">
          <button class="follow-row-main" data-action="open-creator" data-creator-id="${creator.id}" type="button">
            <span class="follow-meta">
              <span class="mini-avatar"></span>
              <span>
                <p class="follow-name">${creator.name}</p>
                <p class="follow-handle">${creator.handle}</p>
              </span>
            </span>
          </button>
          ${
            trailingLabel
              ? `<button class="follow-row-action" data-action="remove-follow" data-creator-id="${creator.id}" type="button">${trailingLabel}</button>`
              : ""
          }
        </div>
      `;
    })
    .join("");
}

function orderedLibraryShortIds() {
  return Array.from(state.libraryShortIds).reverse();
}

function fanShortTiles(shortIds, action, labelKey = "searchLabel") {
  return shortIds
    .map((shortId) => {
      const short = shorts[shortId];
      const label = labelKey === "title" ? short.title.replace("preview", "main") : short.searchLabel;

      return `
        <button class="pin-card" data-action="${action}" data-short-id="${short.id}" type="button">
          <span class="pin-thumb"></span>
          <b>${label}</b>
        </button>
      `;
    })
    .join("");
}

function fanMediaTiles(shortIds, action) {
  return shortIds
    .map((shortId) => {
      const short = shorts[shortId];
      const creator = creators[short.creatorId];
      const label = action === "open-main" ? `${creator.name} ${short.title.replace("preview", "main")}` : `${creator.name} ${short.searchLabel}`;

      return `
        <button
          aria-label="${label}"
          class="fan-media-tile"
          data-action="${action}"
          data-short-id="${short.id}"
          style="--tile-top: ${short.tile.top}; --tile-mid: ${short.tile.mid}; --tile-bottom: ${short.tile.bottom};"
          type="button"
        >
          <span aria-hidden="true" class="fan-media-frame"></span>
        </button>
      `;
    })
    .join("");
}

function creatorMediaTiles(shortIds) {
  return shortIds
    .map((shortId) => {
      const short = shorts[shortId];
      const creator = creators[short.creatorId];

      return `
        <button
          aria-label="${creator.name} ${short.searchLabel}"
          class="creator-media-tile"
          data-action="open-short"
          data-short-id="${short.id}"
          style="--tile-top: ${short.tile.top}; --tile-mid: ${short.tile.mid}; --tile-bottom: ${short.tile.bottom};"
          type="button"
        >
          <span aria-hidden="true" class="creator-media-frame"></span>
        </button>
      `;
    })
    .join("");
}

function pinAction(shortId) {
  const pinned = state.pinnedShortIds.has(shortId);
  const label = pinned ? "Pinned short" : "Pin short";

  return `
    <div class="pin-rail">
      <button aria-label="${label}" class="side-button ${pinned ? "is-active" : ""}" data-action="toggle-pin" data-short-id="${shortId}" type="button">
        <svg aria-hidden="true" class="side-icon" viewBox="0 0 14 18">
          <path d="M3 1.75h8a1 1 0 0 1 1 1V16L7 12.9 2 16V2.75a1 1 0 0 1 1-1Z"></path>
        </svg>
        <span class="sr-only">${label}</span>
      </button>
    </div>
  `;
}

function fanScreen() {
  const followingCreatorIds = Array.from(state.followingCreatorIds);
  const pinnedShortIds = Array.from(state.pinnedShortIds);
  const libraryShortIds = orderedLibraryShortIds();
  const activeFanTab = state.fanTab === "library" ? "library" : "pinned";
  const stats = [
    { label: "Following", count: followingCreatorIds.length, action: "open-following" },
    { label: "Pinned", count: pinnedShortIds.length },
    { label: "Library", count: libraryShortIds.length },
  ];
  const tabs = [
    { key: "pinned", label: "Pinned", icon: "pinned", count: pinnedShortIds.length },
    { key: "library", label: "Library", icon: "library", count: libraryShortIds.length },
  ];

  let panelMarkup = "";

  if (activeFanTab === "library") {
    panelMarkup = libraryShortIds.length
      ? `<div class="fan-media-grid">${fanMediaTiles(libraryShortIds, "open-main")}</div>`
      : `<p class="fan-empty-state">unlock した main はまだありません。</p>`;
  } else {
    panelMarkup = pinnedShortIds.length
      ? `<div class="fan-media-grid">${fanMediaTiles(pinnedShortIds, "open-short")}</div>`
      : `<p class="fan-empty-state">pin した short はまだありません。</p>`;
  }

  return `
    <section class="screen screen-fan" data-theme="fan">
      <div class="fan-screen">
        <div class="fan-topbar">
          ${createBackAction()}
          <button aria-label="Settings" class="fan-settings-button" type="button">
            <svg aria-hidden="true" class="fan-settings-icon" viewBox="0 0 20 20">
              <line x1="10" y1="1.8" x2="10" y2="4.1"></line>
              <line x1="10" y1="15.9" x2="10" y2="18.2"></line>
              <line x1="1.8" y1="10" x2="4.1" y2="10"></line>
              <line x1="15.9" y1="10" x2="18.2" y2="10"></line>
              <line x1="4.2" y1="4.2" x2="5.9" y2="5.9"></line>
              <line x1="14.1" y1="14.1" x2="15.8" y2="15.8"></line>
              <line x1="14.1" y1="5.9" x2="15.8" y2="4.2"></line>
              <line x1="4.2" y1="15.8" x2="5.9" y2="14.1"></line>
              <circle cx="10" cy="10" r="3.1"></circle>
            </svg>
          </button>
        </div>

        <section class="fan-profile-shell">
          <div class="fan-profile-head">
            <span class="fan-profile-avatar"></span>
            <div class="fan-profile-side">
              <div class="fan-profile-stats">
                ${stats
                  .map(
                    (item) => `
                      ${
                        item.action
                          ? `<button class="fan-profile-stat fan-profile-stat-button" data-action="${item.action}" type="button">
                              <strong>${item.count}</strong>
                              <span>${item.label} <b aria-hidden="true">&gt;</b></span>
                            </button>`
                          : `<div class="fan-profile-stat">
                              <strong>${item.count}</strong>
                              <span>${item.label}</span>
                            </div>`
                      }
                    `,
                  )
                  .join("")}
              </div>
            </div>
          </div>

          <div class="fan-profile-copy">
            <p class="fan-profile-name">My archive</p>
          </div>

          <button class="fan-creator-button" data-action="open-creator-dashboard" type="button">creator管理画面へ</button>

          <div aria-label="Profile sections" class="fan-tabbar" role="tablist">
            ${tabs
              .map(
                (tab) => `
                  <button
                    aria-label="${tab.label}"
                    aria-selected="${activeFanTab === tab.key ? "true" : "false"}"
                    class="fan-tab-button ${activeFanTab === tab.key ? "is-active" : ""}"
                    data-action="set-fan-tab"
                    data-tab="${tab.key}"
                    role="tab"
                    type="button"
                  >
                    <span aria-hidden="true" class="fan-tab-icon fan-tab-icon-${tab.icon}"></span>
                    <span class="sr-only">${tab.label}</span>
                  </button>
                `,
              )
              .join("")}
          </div>

          <div class="fan-panel" role="tabpanel">
            ${panelMarkup}
          </div>
        </section>
      </div>
    </section>
  `;
}

function libraryScreen() {
  const shortIds = orderedLibraryShortIds();

  if (!shortIds.length) {
    return `
      <section class="screen" data-theme="profile">
        <div class="library-screen">
          <div class="profile-topbar">
            ${createBackAction()}
          </div>

          <h2 class="library-heading">library</h2>
        </div>
      </section>
    `;
  }

  return `
    <section class="screen" data-theme="profile">
      <div class="library-screen">
        <div class="profile-topbar">
          ${createBackAction()}
        </div>

        <h2 class="library-heading">library</h2>

        <div class="library-search-wrap">
          <input
            class="library-search-input"
            data-role="library-search-input"
            placeholder="Aoi / soft light"
            type="search"
            value="${state.libraryQuery}"
          />
        </div>

        <div class="profile-grid-head">
          <p class="eyebrow">all mains</p>
          <span data-role="library-count-label">${shortIds.length} items</span>
        </div>

        <div class="library-grid">
          ${shortIds
            .map((shortId, index) => {
              const short = shorts[shortId];
              const creator = creators[short.creatorId];
              const statusLabel = state.lastMainShortId === short.id ? short.progress : "play main";

              return `
                <button
                  class="library-tile"
                  data-action="open-main"
                  data-library-item
                  data-library-index="${index + 1}"
                  data-search-text="${`${creator.name} ${creator.handle} ${short.title} ${short.searchLabel}`.toLowerCase()}"
                  data-short-id="${short.id}"
                  style="--tile-top: ${short.tile.top}; --tile-mid: ${short.tile.mid}; --tile-bottom: ${short.tile.bottom};"
                  type="button"
                >
                  <span class="library-tile-frame"></span>
                  <span class="library-tile-meta">
                    <span class="library-tile-creator">${creator.name}</span>
                    <b>${short.title.replace("preview", "main")}</b>
                    <span>${short.duration}</span>
                    <span class="library-tile-status">${statusLabel}</span>
                  </span>
                </button>
              `;
            })
            .join("")}
        </div>
      </div>
    </section>
  `;
}

function followingScreen() {
  const creatorIds = Array.from(state.followingCreatorIds);

  return `
    <section class="screen" data-theme="profile">
      <div class="library-screen">
        <div class="profile-topbar">
          ${createBackAction()}
        </div>

        <h2 class="library-heading">following</h2>

        <div class="library-search-wrap">
          <div class="search-input-wrap">
            <span class="search-input-icon" aria-hidden="true"></span>
            <input class="search-input" data-role="following-search-input" placeholder="検索" type="search" value="${state.followingQuery}" />
          </div>
        </div>

        <div class="profile-grid-head">
          <p class="eyebrow">all creators</p>
          <span data-role="following-count-label">${creatorIds.length} creators</span>
        </div>

        <div class="follow-list">${creatorRows(creatorIds)}</div>
      </div>
    </section>
  `;
}

function pinnedScreen() {
  const shortIds = Array.from(state.pinnedShortIds);

  if (!shortIds.length) {
    return `
      <section class="screen" data-theme="profile">
        <div class="library-screen">
          <div class="profile-topbar">
            ${createBackAction()}
          </div>

          <h2 class="library-heading">pinned shorts</h2>
        </div>
      </section>
    `;
  }

  return `
    <section class="screen" data-theme="profile">
      <div class="library-screen">
        <div class="profile-topbar">
          ${createBackAction()}
        </div>

        <h2 class="library-heading">pinned shorts</h2>

        <div class="profile-grid-head">
          <p class="eyebrow">all pinned</p>
          <span>${shortIds.length} items</span>
        </div>

        <div class="library-grid">
          ${shortIds
            .map((shortId) => {
              const short = shorts[shortId];
              const creator = creators[short.creatorId];

              return `
                <button
                  class="library-tile"
                  data-action="open-short"
                  data-short-id="${short.id}"
                  style="--tile-top: ${short.tile.top}; --tile-mid: ${short.tile.mid}; --tile-bottom: ${short.tile.bottom};"
                  type="button"
                >
                  <span class="library-tile-frame"></span>
                  <span class="library-tile-meta">
                    <span class="library-tile-creator">${creator.name}</span>
                    <b>${short.searchLabel}</b>
                    <span>${short.duration}</span>
                  </span>
                </button>
              `;
            })
            .join("")}
        </div>
      </div>
    </section>
  `;
}

function feedScreen() {
  const shortId = feedShortByTab[state.feedTab];
  const short = shorts[shortId];

  return `
    <section class="screen screen-feed" data-theme="${short.theme}">
      <div class="screen-content">
        <div class="topbar">
          <div class="segmented">
            <button class="${state.feedTab === "recommended" ? "is-active" : ""}" data-action="set-tab" data-tab="recommended" type="button">おすすめ</button>
            <button class="${state.feedTab === "following" ? "is-active" : ""}" data-action="set-tab" data-tab="following" type="button">フォロー中</button>
          </div>
        </div>

        ${pinAction(short.id)}

        ${unlockButton(short)}

        <div class="creator-block">
          ${feedCreatorMeta(short.id, { clickableCaption: false })}
        </div>
      </div>
      ${renderOverlay()}
    </section>
  `;
}

function creatorScreen() {
  const creator = creators[state.currentCreatorId];
  const shortIds = creatorShorts[creator.id];
  const following = state.followingCreatorIds.has(creator.id);

  return `
    <section class="screen" data-theme="profile">
      <div class="profile-screen creator-profile-screen">
        <div class="profile-topbar">
          ${createBackAction()}
        </div>

        <section class="creator-profile-shell">
          <div class="creator-profile-head">
            <span class="creator-profile-avatar"></span>
            <div class="creator-profile-stats">
              <div class="creator-profile-stat">
                <strong>${creator.stats.shorts}</strong>
                <span>shorts</span>
              </div>
              <div class="creator-profile-stat">
                <strong>${creator.stats.fans}</strong>
                <span>fans</span>
              </div>
              <div class="creator-profile-stat">
                <strong>${creator.stats.views}</strong>
                <span>views</span>
              </div>
            </div>
          </div>

          <div class="creator-profile-copy">
            <p class="creator-profile-handle">${creator.handle}</p>
            <p class="creator-profile-name">${creator.name}</p>
            <p class="creator-profile-bio">${creator.bio}</p>
          </div>

          <div class="creator-profile-actions">
            <button class="follow-button ${following ? "is-active" : ""}" data-action="toggle-follow" data-creator-id="${creator.id}" type="button">
              ${following ? "Following" : "Follow"}
            </button>
          </div>

          <div aria-label="Shorts" class="creator-profile-tabbar" role="tablist">
            <button aria-label="Shorts grid" aria-selected="true" class="creator-profile-tab is-active" role="tab" type="button">
              <svg aria-hidden="true" class="creator-profile-tab-icon" viewBox="0 0 18 18">
                <rect x="2" y="2" width="4" height="4" rx="1"></rect>
                <rect x="7" y="2" width="4" height="4" rx="1"></rect>
                <rect x="12" y="2" width="4" height="4" rx="1"></rect>
                <rect x="2" y="7" width="4" height="4" rx="1"></rect>
                <rect x="7" y="7" width="4" height="4" rx="1"></rect>
                <rect x="12" y="7" width="4" height="4" rx="1"></rect>
                <rect x="2" y="12" width="4" height="4" rx="1"></rect>
                <rect x="7" y="12" width="4" height="4" rx="1"></rect>
                <rect x="12" y="12" width="4" height="4" rx="1"></rect>
              </svg>
            </button>
          </div>

          <div class="creator-media-grid">
            ${creatorMediaTiles(shortIds)}
          </div>
        </section>
      </div>
      ${renderOverlay()}
    </section>
  `;
}

function creatorDashboardScreen() {
  const creator = creators[viewerCreatorId];
  const dashboard = creatorDashboardData[viewerCreatorId];
  const activeTab = state.creatorManagerTab === "main" ? "main" : "shorts";
  const revisionNotice = dashboard.revisionNotice || null;
  const visibleItems = (activeTab === "main" ? dashboard.mains : dashboard.shorts).map((item) =>
    resolveCreatorDashboardItem(item, activeTab),
  );

  return `
    <section class="screen" data-theme="profile">
      <div class="profile-screen creator-profile-screen creator-manager-screen">
        <div class="profile-topbar">
          ${createBackAction()}
          ${createCreatorAddAction()}
        </div>

        <section class="creator-profile-shell">
          <div class="creator-profile-head">
            <span class="creator-profile-avatar"></span>
            <div class="creator-profile-stats">
              ${dashboard.profileStats
                .map(
                  (item) => `
                    <div class="creator-profile-stat">
                      <strong>${item.value}</strong>
                      <span>${item.label}</span>
                    </div>
                  `,
                )
                .join("")}
            </div>
          </div>

          <div class="creator-profile-copy creator-manager-copy">
            <p class="creator-profile-handle">${creator.handle}</p>
            <p class="creator-profile-name">${creator.name}</p>
            <p class="creator-profile-bio">${dashboard.description}</p>
          </div>

          ${
            revisionNotice
              ? `
                <div class="creator-manager-alert" role="status">
                  <span class="creator-manager-alert-badge">差し戻し</span>
                  <div class="creator-manager-alert-copy">
                    <b>${revisionNotice.label}</b>
                    <span>${revisionNotice.detail}</span>
                  </div>
                </div>
              `
              : ""
          }

          <div aria-label="Managed posts" class="creator-profile-tabbar" role="tablist">
            <button
              aria-label="Shorts"
              aria-selected="${activeTab === "shorts" ? "true" : "false"}"
              class="creator-profile-tab creator-manager-tab-button ${activeTab === "shorts" ? "is-active" : ""}"
              data-action="set-creator-manager-tab"
              data-tab="shorts"
              role="tab"
              type="button"
            >
              Shorts
            </button>
            <button
              aria-label="Main"
              aria-selected="${activeTab === "main" ? "true" : "false"}"
              class="creator-profile-tab creator-manager-tab-button ${activeTab === "main" ? "is-active" : ""}"
              data-action="set-creator-manager-tab"
              data-tab="main"
              role="tab"
              type="button"
            >
              Main
            </button>
          </div>

          <div class="creator-manager-grid">
            ${visibleItems
              .map((item) => {
                const short = shorts[item.shortId];

                return `
                  <button
                    aria-label="${short.title}"
                    class="creator-manager-tile is-${item.tone}"
                    data-action="open-creator-manager-detail"
                    data-kind="${activeTab}"
                    data-short-id="${item.shortId}"
                    type="button"
                  >
                    <span
                      aria-hidden="true"
                      class="creator-manager-tile-frame"
                      style="--tile-top: ${short.tile.top}; --tile-mid: ${short.tile.mid}; --tile-bottom: ${short.tile.bottom};"
                    ></span>
                    <span class="creator-manager-tile-overlay">
                      ${
                        item.tone !== "approved"
                          ? `<span class="creator-manager-tile-status is-${item.tone} is-center">${item.status}</span>`
                          : ""
                      }
                    </span>
                  </button>
                `;
              })
              .join("")}
          </div>
        </section>
      </div>
      ${renderOverlay()}
    </section>
  `;
}

function currentCreatorManagerDetail() {
  const detailTab = state.creatorManagerDetailTab === "main" ? "main" : "shorts";
  const detailShortId = state.creatorManagerDetailShortId || state.currentShortId;
  const fallbackShortId = creatorShorts[viewerCreatorId][0];
  const short = shorts[detailShortId] || shorts[fallbackShortId];
  const baseDetail =
    creatorManagerDetailData[detailTab][detailShortId] ||
    creatorManagerDetailData[detailTab][fallbackShortId] ||
    creatorManagerDetailData.shorts[fallbackShortId];
  const detail = {
    ...baseDetail,
    metrics: baseDetail.metrics.map((item) => ({ ...item })),
    settings: baseDetail.settings.map((item) => ({ ...item })),
  };
  const mainKey = detail.mainKey || detailShortId;
  const mainState = currentCreatorMainState(mainKey);
  const shortVisibilityState = currentCreatorShortVisibility(detailShortId);

  if (detailTab === "main") {
    if (mainState === "paused") {
      detail.statusLabel = "新規公開停止";
      detail.statusTone = "paused";
      detail.summary = "公開中のショートがないため、新規流入は停止中です。既存購入者の視聴は継続します。";
      detail.settings = detail.settings.map((item) =>
        item.label === "紐づくショート" ? { label: "公開中ショート", value: "0本" } : item,
      );
    }

    if (mainState === "hidden" || mainState === "deleted") {
      const hidden = mainState === "hidden";
      detail.statusLabel = hidden ? "非公開" : "削除済み";
      detail.statusTone = hidden ? "hidden" : "removed";
      detail.summary = hidden
        ? "この本編を非公開にしたため、紐づくショートも公開面から外れています。"
        : "この本編を削除したため、紐づくショートも合わせて削除されています。";
      detail.settings = detail.settings.map((item) => {
        if (item.label === "レビュー") {
          return { ...item, value: hidden ? "非公開" : "削除済み" };
        }

        if (item.label === "紐づくショート" || item.label === "公開中ショート") {
          return { label: hidden ? "非公開ショート" : "削除済みショート", value: `${linkedShortIdsForMain(mainKey).length}本` };
        }

        return item;
      });
    }
  }

  if (detailTab === "shorts" && shortVisibilityState !== "live") {
    const hidden = shortVisibilityState === "hidden";
    detail.statusLabel = hidden ? "非公開" : "削除済み";
    detail.statusTone = hidden ? "hidden" : "removed";
    detail.summary =
      mainState === "hidden" || mainState === "deleted"
        ? `リンク先本編を${mainState === "hidden" ? "非公開" : "削除"}にしたため、このショートも合わせて${hidden ? "非公開" : "削除"}されています。`
        : mainState === "paused"
        ? `このショートを${hidden ? "非公開" : "削除"}にしたため、本編の新規公開も停止中です。`
        : `このショートは${hidden ? "公開面から外れています" : "公開面から削除されています"}。`;
    detail.settings = detail.settings.map((item) =>
      item.label === "公開範囲" ? { ...item, value: hidden ? "非公開" : "削除済み" } : item,
    );
  }

  return {
    detail,
    detailTab,
    mainState,
    short,
    shortVisibilityState,
  };
}

function resolveCreatorDashboardItem(item, activeTab) {
  const resolvedItem = { ...item };

  if (activeTab === "main") {
    const mainState = currentCreatorMainState(item.shortId);

    if (mainState === "paused") {
      resolvedItem.status = "停止中";
      resolvedItem.tone = "paused";
    }

    if (mainState === "hidden") {
      resolvedItem.status = "非公開";
      resolvedItem.tone = "hidden";
    }

    if (mainState === "deleted") {
      resolvedItem.status = "削除";
      resolvedItem.tone = "removed";
    }
  }

  if (activeTab === "shorts") {
    const shortVisibility = currentCreatorShortVisibility(item.shortId);

    if (shortVisibility === "hidden") {
      resolvedItem.status = "非公開";
      resolvedItem.tone = "hidden";
    }

    if (shortVisibility === "deleted") {
      resolvedItem.status = "削除";
      resolvedItem.tone = "removed";
    }
  }

  return resolvedItem;
}

function creatorLinkedShortItems(mainKey) {
  return linkedShortIdsForMain(mainKey).map((shortId) => {
    const shortDetail = creatorManagerDetailData.shorts[shortId];

    return resolveCreatorDashboardItem(
      {
        shortId,
        status: shortDetail.statusLabel,
        tone: shortDetail.statusTone,
      },
      "shorts",
    );
  });
}

function creatorLinkedMainItem(shortId) {
  const shortDetail = creatorManagerDetailData.shorts[shortId];

  if (!shortDetail?.mainKey) {
    return null;
  }

  const mainDetail = creatorManagerDetailData.main[shortDetail.mainKey];

  if (!mainDetail) {
    return null;
  }

  return resolveCreatorDashboardItem(
    {
      shortId: shortDetail.mainKey,
      status: mainDetail.statusLabel,
      tone: mainDetail.statusTone,
    },
    "main",
  );
}

function creatorPostDetailScreen() {
  const creator = creators[viewerCreatorId];
  const { detail, detailTab, mainState, short, shortVisibilityState } = currentCreatorManagerDetail();
  const showSettingsButton =
    detailTab === "main" ? mainState === "active" || mainState === "paused" : shortVisibilityState === "live";
  const linkedShortItems = detailTab === "main" ? creatorLinkedShortItems(state.creatorManagerDetailShortId || state.currentShortId) : [];
  const linkedMainItem = detailTab === "shorts" ? creatorLinkedMainItem(state.creatorManagerDetailShortId || state.currentShortId) : null;
  const visibleSettings = detail.settings.filter((item) => !(detailTab === "shorts" && item.action === "open-linked-main"));

  return `
    <section class="screen" data-theme="profile">
      <div class="profile-screen creator-post-detail-screen">
        <div class="profile-topbar">
          ${createBackAction()}
          ${
            showSettingsButton
              ? '<button aria-label="投稿操作" class="creator-post-detail-menu-button" data-action="open-creator-post-actions" type="button"><span></span><span></span><span></span></button>'
              : ""
          }
        </div>

        <section class="creator-post-detail-shell">
          <div class="creator-post-detail-author">
            <span class="creator-post-detail-avatar"></span>
            <div class="creator-post-detail-author-copy">
              <p class="creator-post-detail-handle">${creator.handle}</p>
            </div>
          </div>

          <div class="creator-post-detail-media is-${detail.statusTone}">
            <span
              aria-hidden="true"
              class="creator-post-detail-media-frame"
              style="--tile-top: ${short.tile.top}; --tile-mid: ${short.tile.mid}; --tile-bottom: ${short.tile.bottom};"
            ></span>
            <div class="creator-post-detail-media-overlay">
              <div class="creator-post-detail-media-meta">
                <span class="creator-post-detail-kind">${detail.kindLabel}</span>
                <span class="creator-post-detail-status is-${detail.statusTone}">${detail.statusLabel}</span>
              </div>
              <span class="creator-post-detail-play" aria-hidden="true"></span>
              <span class="creator-post-detail-duration">${short.duration}</span>
            </div>
          </div>

          <div class="creator-post-detail-summary">
            <p class="creator-post-detail-summary-text">${detail.summary}</p>
          </div>

          <div class="creator-post-detail-metrics">
            ${detail.metrics
              .map(
                (metric) => `
                  <div class="creator-post-detail-metric">
                    <strong>${metric.value}</strong>
                    <span>${metric.label}</span>
                  </div>
                `,
              )
              .join("")}
          </div>

          <section class="creator-post-detail-section">
            <h3 class="creator-post-detail-section-title">設定</h3>
            <div class="creator-post-detail-setting-list">
              ${visibleSettings
                .map(
                  (item) =>
                    item.action === "open-linked-main"
                      ? `
                          <button class="creator-post-detail-setting-row is-link" data-action="open-creator-linked-main" data-short-id="${item.mainKey}" type="button">
                            <span>${item.label}</span>
                            <strong>${item.value}</strong>
                          </button>
                        `
                      : `
                          <div class="creator-post-detail-setting-row">
                            <span>${item.label}</span>
                            <strong>${item.value}</strong>
                          </div>
                        `,
                )
                .join("")}
            </div>
          </section>

          ${
            detailTab === "main"
              ? `
                <section class="creator-post-detail-section">
                  <h3 class="creator-post-detail-section-title">紐づくショート</h3>
                  <div class="creator-post-detail-linked-grid">
                    ${linkedShortItems
                      .map((item) => {
                        const linkedShort = shorts[item.shortId];

                        return `
                          <button
                            aria-label="${linkedShort.title}"
                            class="creator-post-detail-linked-tile creator-manager-tile is-${item.tone}"
                            data-action="open-creator-manager-detail"
                            data-kind="shorts"
                            data-short-id="${item.shortId}"
                            type="button"
                          >
                            <span
                              aria-hidden="true"
                              class="creator-manager-tile-frame"
                              style="--tile-top: ${linkedShort.tile.top}; --tile-mid: ${linkedShort.tile.mid}; --tile-bottom: ${linkedShort.tile.bottom};"
                            ></span>
                            <span class="creator-manager-tile-overlay">
                              ${item.tone !== "approved" ? `<span class="creator-manager-tile-status is-${item.tone} is-center">${item.status}</span>` : ""}
                            </span>
                          </button>
                        `;
                      })
                      .join("")}
                  </div>
                </section>
              `
              : ""
          }

          ${
            detailTab === "shorts" && linkedMainItem
              ? `
                <section class="creator-post-detail-section">
                  <h3 class="creator-post-detail-section-title">紐づく本編</h3>
                  <div class="creator-post-detail-linked-grid">
                    <button
                      aria-label="${shorts[linkedMainItem.shortId].title}"
                      class="creator-post-detail-linked-tile creator-manager-tile is-${linkedMainItem.tone}"
                      data-action="open-creator-manager-detail"
                      data-kind="main"
                      data-short-id="${linkedMainItem.shortId}"
                      type="button"
                    >
                      <span
                        aria-hidden="true"
                        class="creator-manager-tile-frame"
                        style="--tile-top: ${shorts[linkedMainItem.shortId].tile.top}; --tile-mid: ${shorts[linkedMainItem.shortId].tile.mid}; --tile-bottom: ${shorts[linkedMainItem.shortId].tile.bottom};"
                      ></span>
                      <span class="creator-manager-tile-overlay">
                        ${
                          linkedMainItem.tone !== "approved"
                            ? `<span class="creator-manager-tile-status is-${linkedMainItem.tone} is-center">${linkedMainItem.status}</span>`
                            : ""
                        }
                      </span>
                    </button>
                  </div>
                </section>
              `
              : ""
          }
        </section>
      </div>
      ${renderOverlay()}
    </section>
  `;
}

function creatorUploadScreen() {
  const linkingToExistingMain = state.creatorUploadMode === "link-short" && Boolean(state.creatorUploadTargetMainId);
  const mainSelected = linkingToExistingMain || Boolean(state.creatorUploadMainName);
  const shortUploadSlots = state.creatorUploadShortNames;
  const shortsSelectedCount = shortUploadSlots.filter(Boolean).length;
  const shortsSelected = shortsSelectedCount > 0;
  const uploadReady = isCreatorUploadReady();
  const pageTitle = linkingToExistingMain ? "ショートを追加" : "本編とショートを追加";
  const submitLabel = linkingToExistingMain ? "ショートを追加" : "アップロード";

  return `
    <section class="screen" data-theme="profile">
      <div class="profile-screen creator-upload-screen">
        <div class="profile-topbar">
          ${createBackAction()}
        </div>

        <section class="creator-upload-shell">
          <h2 class="creator-upload-page-title">${pageTitle}</h2>

          ${
            linkingToExistingMain
              ? ""
              : '<input accept="video/*" class="creator-upload-native-input" data-kind="main" data-role="creator-upload-input" type="file" />'
          }
          ${shortUploadSlots
            .map(
              (_, slotIndex) => `
                <input
                  accept="video/*"
                  class="creator-upload-native-input"
                  data-index="${slotIndex}"
                  data-kind="shorts"
                  data-role="creator-upload-input"
                  type="file"
                />
              `,
            )
            .join("")}

          <div class="creator-upload-form">
            ${
              linkingToExistingMain
                ? `
                  <section class="creator-upload-field">
                    <div class="creator-upload-field-head">
                      <div>
                        <p class="creator-upload-field-label">main</p>
                        <h3 class="creator-upload-field-title">紐づけ先の本編</h3>
                      </div>
                      <span class="creator-upload-field-state is-filled">選択済み</span>
                    </div>

                    <p class="creator-upload-file-name">今開いている本編にショートを追加します</p>
                  </section>
                `
                : `
                  <section class="creator-upload-field">
                    <div class="creator-upload-field-head">
                      <div>
                        <p class="creator-upload-field-label">main</p>
                        <h3 class="creator-upload-field-title">本編動画</h3>
                      </div>
                      <span class="creator-upload-field-state ${mainSelected ? "is-filled" : ""}">${mainSelected ? "選択済み" : "未選択"}</span>
                    </div>

                    ${
                      mainSelected
                        ? `<p class="creator-upload-file-name">${escapeHtml(state.creatorUploadMainName)}</p>`
                        : `<p class="creator-upload-empty">本編動画を追加してください</p>`
                    }

                    <button class="creator-upload-pick-button" data-action="pick-creator-upload-files" data-kind="main" type="button">
                      ${mainSelected ? "本編を選び直す" : "本編を追加"}
                    </button>
                  </section>
                `
            }

            <section class="creator-upload-section">
              <div class="creator-upload-section-head">
                <div>
                  <p class="creator-upload-field-label">shorts</p>
                  <h3 class="creator-upload-field-title">ショート動画</h3>
                </div>
                <div class="creator-upload-section-side">
                  <span class="creator-upload-field-state ${shortsSelected ? "is-filled" : ""}">${shortsSelected ? `${shortsSelectedCount}本` : "未選択"}</span>
                  <button aria-label="ショート欄を追加" class="creator-upload-short-add-button" data-action="add-creator-upload-short-slot" type="button">+</button>
                </div>
              </div>

              <div class="creator-upload-short-list">
                ${
                  shortUploadSlots.length
                    ? shortUploadSlots
                        .map(
                          (fileName, slotIndex) => `
                            <section class="creator-upload-field creator-upload-short-field">
                              <div class="creator-upload-field-head">
                                <div>
                                  <p class="creator-upload-field-label">short ${slotIndex + 1}</p>
                                  <h3 class="creator-upload-field-title">ショート動画 ${slotIndex + 1}</h3>
                                </div>
                                <div class="creator-upload-field-side">
                                  <span class="creator-upload-field-state ${fileName ? "is-filled" : ""}">${fileName ? "選択済み" : "未選択"}</span>
                                  <button aria-label="ショート欄を削除" class="creator-upload-short-remove-button" data-action="remove-creator-upload-short-slot" data-index="${slotIndex}" type="button">-</button>
                                </div>
                              </div>

                              ${
                                fileName
                                  ? `<p class="creator-upload-file-name">${escapeHtml(fileName)}</p>`
                                  : `<p class="creator-upload-empty">ショート動画を追加してください</p>`
                              }

                              <button class="creator-upload-pick-button" data-action="pick-creator-upload-files" data-index="${slotIndex}" data-kind="shorts" type="button">
                                ${fileName ? "ショートを選び直す" : "ショートを追加"}
                              </button>
                            </section>
                          `,
                        )
                        .join("")
                    : `<p class="creator-upload-short-empty">ショート動画を追加してください</p>`
                }
              </div>
            </section>
          </div>

          <div class="creator-upload-actions">
            <button class="creator-upload-submit-button" ${uploadReady ? "" : "disabled"} data-action="submit-creator-upload-package" type="button">
              ${submitLabel}
            </button>
          </div>
        </section>
      </div>
    </section>
  `;
}

function handleBack() {
  if (state.overlay) {
    closeOverlay();
    return;
  }

  const previous = state.history.pop();

  if (!previous) {
    switchPrimaryTab(state.rootTab);
    return;
  }

  restore(previous);
}

function mainScreen() {
  const short = currentShort();

  return `
    <section class="screen screen-main" data-theme="${short.theme}">
      <div class="screen-content">
        <div class="topbar">
          <div class="topbar-group">
            ${createBackAction()}
          </div>
        </div>

        ${pinAction(short.id)}

        <div class="unlock-cta is-unlocked">
          <div class="unlock-left">
            <span class="unlock-dot">Play</span>
            <span>
              <p class="unlock-title">Playing main</p>
              <p class="unlock-copy">resume without another confirmation</p>
            </span>
          </div>
          <span class="unlock-pill">${short.progress}</span>
        </div>

        <div class="creator-block">
          <div class="creator-card">
            ${creatorCard(short.creatorId)}
            <p class="caption-button">${short.title} の続き。</p>
          </div>
        </div>
      </div>
    </section>
  `;
}

function navigate(nextState, replace = false) {
  if (!replace) {
    state.history.push(snapshot());
  }

  state.acceptAge = false;
  state.acceptTerms = false;
  state.creatorPendingAction = null;
  state.overlay = null;
  state.overlayShortId = null;

  Object.assign(state, nextState);
  state.currentCreatorId = shorts[state.currentShortId].creatorId;
  render();
}

function openMain(shortId) {
  state.purchasedShortIds.add(shortId);
  state.libraryShortIds.add(shortId);
  state.lastMainShortId = shortId;

  navigate({
    currentCreatorId: shorts[shortId].creatorId,
    currentShortId: shortId,
    screen: "main",
  });
}

function openLibrary() {
  navigate({
    libraryQuery: "",
    rootTab: "fan",
    screen: "library",
  });
}

function openFollowing() {
  navigate({
    rootTab: "fan",
    screen: "following",
  });
}

function openPinned() {
  navigate({
    rootTab: "fan",
    screen: "pinned",
  });
}

function openPaywall(shortId) {
  if (state.hasPurchaseSetup) {
    openMain(shortId);
    return;
  }

  state.overlay = "paywall";
  state.overlayShortId = shortId;
  state.acceptAge = false;
  state.acceptTerms = false;
  render();
}

function creatorShortActionOverlay() {
  const pendingAction = state.creatorPendingAction;

  if (!pendingAction || pendingAction.targetType === "main") {
    return "";
  }

  const short = shorts[pendingAction.shortId];
  const actionLabel = pendingAction.actionKind === "delete" ? "削除" : "非公開";
  const confirmLabel = pendingAction.actionKind === "delete" ? "削除する" : "非公開にする";

  return `
    <div class="overlay-backdrop" data-action="close-overlay"></div>
    <div class="overlay-panel creator-short-action-panel">
      <div class="creator-short-action-head">
        <div>
          <h3 class="creator-short-action-title">
            ${
              pendingAction.willStopMainFlow
                ? `${actionLabel}にすると、本編の新規公開も停止します`
                : `${short.title} を${actionLabel}にしますか`
            }
          </h3>
        </div>
      </div>

      <p class="creator-short-action-copy">
        ${
          pendingAction.willStopMainFlow
            ? "この本編には他に公開中のショートがありません。既存購入者の視聴は維持されます。"
            : pendingAction.actionKind === "delete"
              ? "プロフィールと feed から外れ、公開導線として使えなくなります。"
              : "プロフィールと feed から外れ、公開導線として一時的に止まります。"
        }
      </p>

      <div class="creator-short-action-actions">
        <button class="secondary-action" data-action="close-overlay" type="button">キャンセル</button>
        ${
          pendingAction.willStopMainFlow
            ? `<button class="secondary-action" data-action="open-creator-upload" type="button">ショートを追加してから続ける</button>`
            : ""
        }
        <button class="danger-action" data-action="confirm-creator-short-action" type="button">${confirmLabel}</button>
      </div>
    </div>
  `;
}

function creatorMainActionOverlay() {
  const pendingAction = state.creatorPendingAction;

  if (!pendingAction || pendingAction.targetType !== "main") {
    return "";
  }

  const actionLabel = pendingAction.actionKind === "delete" ? "削除" : "非公開";
  const confirmLabel = pendingAction.actionKind === "delete" ? "削除する" : "非公開にする";
  const linkedShortCount = pendingAction.linkedShortIds.length;

  return `
    <div class="overlay-backdrop" data-action="close-overlay"></div>
    <div class="overlay-panel creator-short-action-panel">
      <div class="creator-short-action-head">
        <div>
          <h3 class="creator-short-action-title">本編を${actionLabel}にすると、紐づくショートも合わせて${actionLabel}になります</h3>
        </div>
      </div>

      <p class="creator-short-action-copy">
        ${
          pendingAction.actionKind === "delete"
            ? `この本編と紐づくショート${linkedShortCount}本がプロフィールと feed から削除され、公開導線として使えなくなります。`
            : `この本編と紐づくショート${linkedShortCount}本がプロフィールと feed から外れます。既存購入者の視聴は維持されます。`
        }
      </p>

      <div class="creator-short-action-actions">
        <button class="secondary-action" data-action="close-overlay" type="button">キャンセル</button>
        <button class="danger-action" data-action="confirm-creator-main-action" type="button">${confirmLabel}</button>
      </div>
    </div>
  `;
}

function creatorPostActionDataset({ action, kind, shortId }) {
  const attributes = [`data-action="${action}"`];

  if (kind) {
    attributes.push(`data-kind="${kind}"`);
  }

  if (shortId) {
    attributes.push(`data-short-id="${shortId}"`);
  }

  return attributes.join(" ");
}

function creatorPostActionIcon(kind) {
  if (kind === "link-short") {
    return `
      <svg aria-hidden="true" class="creator-post-actions-icon-svg" viewBox="0 0 20 20">
        <rect x="4.25" y="4.25" width="11.5" height="11.5" rx="3"></rect>
        <path d="M10 6.3v7.4"></path>
        <path d="M6.3 10h7.4"></path>
      </svg>
    `;
  }

  if (kind === "open-main") {
    return `
      <svg aria-hidden="true" class="creator-post-actions-icon-svg" viewBox="0 0 20 20">
        <rect x="3.75" y="4.75" width="12.5" height="10.5" rx="2.5"></rect>
        <path d="M9 8.1l3.3 1.9L9 11.9z"></path>
      </svg>
    `;
  }

  if (kind === "hide") {
    return `
      <svg aria-hidden="true" class="creator-post-actions-icon-svg" viewBox="0 0 20 20">
        <path d="M2.7 10s2.7-4.3 7.3-4.3S17.3 10 17.3 10 14.6 14.3 10 14.3 2.7 10 2.7 10Z"></path>
        <circle cx="10" cy="10" r="2.1"></circle>
        <path d="M4.1 15.9 15.9 4.1"></path>
      </svg>
    `;
  }

  return `
    <svg aria-hidden="true" class="creator-post-actions-icon-svg" viewBox="0 0 20 20">
      <path d="M6.2 6.5h7.6"></path>
      <path d="M7.3 6.5V5.1h5.4v1.4"></path>
      <path d="M7 6.5l.6 8.2h4.8l.6-8.2"></path>
      <path d="M8.9 8.8v3.8"></path>
      <path d="M11.1 8.8v3.8"></path>
    </svg>
  `;
}

function creatorPostActionRow(item) {
  return `
    <button class="creator-post-actions-row ${item.tone === "danger" ? "is-danger" : ""}" ${creatorPostActionDataset(item)} type="button">
      <span class="creator-post-actions-row-main">
        <span class="creator-post-actions-row-icon ${item.tone === "danger" ? "is-danger" : ""}">
          ${creatorPostActionIcon(item.icon)}
        </span>
        <span class="creator-post-actions-row-label">${item.label}</span>
      </span>
      ${item.chevron ? '<span aria-hidden="true" class="creator-post-actions-row-chevron">&gt;</span>' : ""}
    </button>
  `;
}

function creatorPostActionsOverlay() {
  const { detail, detailTab } = currentCreatorManagerDetail();

  const rows =
    detailTab === "main"
      ? [
          { action: "open-creator-linked-short-upload", chevron: true, icon: "link-short", label: "新しいshortsを紐づける" },
          { action: "request-creator-main-action", icon: "hide", kind: "hide", label: "この本編を非公開にする" },
          { action: "request-creator-main-action", icon: "delete", kind: "delete", label: "この本編を削除する", tone: "danger" },
        ]
      : [
          { action: "open-creator-linked-main", chevron: true, icon: "open-main", label: "紐づく本編を開く", shortId: detail.mainKey },
          { action: "request-creator-short-action", icon: "hide", kind: "hide", label: "このショートを非公開にする" },
          { action: "request-creator-short-action", icon: "delete", kind: "delete", label: "このショートを削除する", tone: "danger" },
        ];

  return `
    <div class="overlay-backdrop" data-action="close-overlay"></div>
    <div class="overlay-panel creator-post-actions-panel">
      <div aria-hidden="true" class="creator-post-actions-handle"></div>

      <div class="creator-post-actions-group">
        <div class="creator-post-actions-list">
          ${rows.map((item) => creatorPostActionRow(item)).join("")}
        </div>
      </div>
    </div>
  `;
}

function paywallOverlay() {
  const short = shorts[state.overlayShortId];

  return `
    <div class="overlay-backdrop" data-action="close-overlay"></div>
    <div class="overlay-panel paywall-panel">
      <div class="paywall-head">
        <div>
          <p class="eyebrow">unlock</p>
          <h3 class="paywall-title">${short.title} の続きを見る</h3>
        </div>
        <span class="mini-badge">${short.price}</span>
      </div>

      <div class="paywall-card">
        <span class="card-badge">Card</span>
        <span>
          <p class="paywall-card-title">Visa ending in 4242</p>
          <p class="paywall-card-copy">支払い方法は保存済みです。</p>
        </span>
      </div>

      <div class="paywall-checks">
        <label class="check-row">
          <input ${state.acceptAge ? "checked" : ""} data-role="accept-age" type="checkbox" />
          <span>18歳以上であり、年齢確認に同意する</span>
        </label>
        <label class="check-row">
          <input ${state.acceptTerms ? "checked" : ""} data-role="accept-terms" type="checkbox" />
          <span>利用規約とポリシーに同意し、確認面なしで main 再生へ進む</span>
        </label>
      </div>

      <div class="paywall-actions">
        <button class="secondary-action" data-action="close-overlay" type="button">閉じる</button>
        <button class="primary-action" ${state.acceptAge && state.acceptTerms ? "" : "disabled"} data-action="confirm-paywall" type="button">
          Unlock ${short.price} | ${short.duration}
        </button>
      </div>
    </div>
  `;
}

function render() {
  root.innerHTML = `${screenMarkup()}${tabBarMarkup()}`;
  updateSearchResults();
  updateLibraryResults();
  updateFollowingResults();
  updatePaywallState();
}

function renderOverlay() {
  if (state.overlay === "creator-post-actions") {
    return creatorPostActionsOverlay();
  }

  if (state.overlay === "creator-main-action") {
    return creatorMainActionOverlay();
  }

  if (state.overlay === "creator-short-action") {
    return creatorShortActionOverlay();
  }

  if (state.overlay === "paywall") {
    return paywallOverlay();
  }

  return "";
}

function restore(snapshotState) {
  state.acceptAge = false;
  state.acceptTerms = false;
  state.creatorPendingAction = null;
  state.overlay = null;
  state.overlayShortId = null;
  state.fanTab = snapshotState.fanTab || "pinned";
  state.followingQuery = snapshotState.followingQuery || "";
  state.searchQuery = snapshotState.searchQuery;
  state.rootTab = snapshotState.rootTab;
  state.screen = snapshotState.screen;
  state.feedTab = snapshotState.feedTab;
  state.libraryQuery = snapshotState.libraryQuery;
  state.creatorManagerDetailShortId = snapshotState.creatorManagerDetailShortId || "rooftop";
  state.creatorManagerDetailTab = snapshotState.creatorManagerDetailTab || "shorts";
  state.creatorManagerTab = snapshotState.creatorManagerTab || "shorts";
  state.creatorUploadMode = snapshotState.creatorUploadMode || "new-package";
  state.creatorUploadMainName = snapshotState.creatorUploadMainName || "";
  state.creatorUploadShortNames = snapshotState.creatorUploadShortNames || [];
  state.creatorUploadTargetMainId = snapshotState.creatorUploadTargetMainId || null;
  state.currentShortId = snapshotState.currentShortId;
  state.currentCreatorId = snapshotState.currentCreatorId;
  render();
}

function screenMarkup() {
  if (state.screen === "creator-post-detail") {
    return creatorPostDetailScreen();
  }

  if (state.screen === "creator-upload") {
    return creatorUploadScreen();
  }

  if (state.screen === "creator-dashboard") {
    return creatorDashboardScreen();
  }

  if (state.screen === "creator") {
    return creatorScreen();
  }

  if (state.screen === "search") {
    return searchScreen();
  }

  if (state.screen === "fan") {
    return fanScreen();
  }

  if (state.screen === "library") {
    return libraryScreen();
  }

  if (state.screen === "following") {
    return followingScreen();
  }

  if (state.screen === "pinned") {
    return pinnedScreen();
  }

  if (state.screen === "main") {
    return mainScreen();
  }

  if (state.screen === "short") {
    return shortScreen();
  }

  return feedScreen();
}

function searchScreen() {
  return `
    <section class="screen" data-theme="profile">
      <div class="search-screen">
        <div class="search-input-wrap">
          <span class="search-input-icon" aria-hidden="true"></span>
          <input class="search-input" data-role="search-input" placeholder="検索" type="search" value="${state.searchQuery}" />
        </div>
        <p class="search-section-label">最近</p>
        <div class="search-results">
          ${Object.values(creators)
            .map(
              (creator) => `
                <button class="search-result" data-action="open-creator" data-creator-id="${creator.id}" data-search-item type="button">
                  <span class="search-meta">
                    <span class="mini-avatar"></span>
                    <span>
                      <p class="search-name">${creator.name}</p>
                      <p class="search-handle">${creator.handle}</p>
                    </span>
                  </span>
                  <span class="section-copy" style="color: var(--text-body);">open</span>
                </button>
              `,
            )
            .join("")}
        </div>
      </div>
    </section>
  `;
}

function shortScreen() {
  const short = currentShort();

  return `
    <section class="screen screen-feed screen-short" data-theme="${currentTheme()}">
      <div class="screen-content">
        <div class="topbar">
          ${createBackAction()}
        </div>

        ${pinAction(short.id)}

        ${unlockButton(short)}

        <div class="creator-block">
          ${feedCreatorMeta(short.id, { clickableCaption: false })}
        </div>
      </div>
      ${renderOverlay()}
    </section>
  `;
}

function snapshot() {
  return {
    creatorManagerDetailShortId: state.creatorManagerDetailShortId,
    creatorManagerDetailTab: state.creatorManagerDetailTab,
    creatorManagerTab: state.creatorManagerTab,
    creatorUploadMode: state.creatorUploadMode,
    creatorUploadMainName: state.creatorUploadMainName,
    creatorUploadShortNames: [...state.creatorUploadShortNames],
    creatorUploadTargetMainId: state.creatorUploadTargetMainId,
    currentCreatorId: state.currentCreatorId,
    currentShortId: state.currentShortId,
    fanTab: state.fanTab,
    feedTab: state.feedTab,
    followingQuery: state.followingQuery,
    libraryQuery: state.libraryQuery,
    rootTab: state.rootTab,
    searchQuery: state.searchQuery,
    screen: state.screen,
  };
}

function switchPrimaryTab(tab) {
  state.history = [];
  state.searchQuery = "";

  if (tab === "feed") {
    navigate(
      {
        currentCreatorId: shorts[feedShortByTab[state.feedTab]].creatorId,
        currentShortId: feedShortByTab[state.feedTab],
        rootTab: "feed",
        screen: "feed",
      },
      true,
    );
    return;
  }

  if (tab === "search") {
    navigate({ rootTab: "search", screen: "search" }, true);
    return;
  }

  navigate({ rootTab: "fan", screen: "fan" }, true);
}

function tabBarMarkup() {
  if (state.screen === "creator-dashboard" || state.screen === "creator-post-detail" || state.screen === "creator-upload") {
    return "";
  }

  const items = [
    { action: "open-feed", ariaLabel: "フィード", icon: "feed", key: "feed" },
    { action: "open-search", ariaLabel: "検索", icon: "search", key: "search" },
    { action: "open-fan", ariaLabel: "マイ", icon: "fan", key: "fan" },
  ];

  return `
    <nav class="tabbar" aria-label="Primary">
      ${items
        .map(
          (item) => `
            <button
              aria-label="${item.ariaLabel || item.label}"
              class="tabbar-button ${state.rootTab === item.key ? "is-active" : ""}"
              data-action="${item.action}"
              type="button"
            >
              ${
                item.icon
                  ? `<span class="tabbar-icon tabbar-icon-${item.icon}" aria-hidden="true"></span><span class="sr-only">${item.ariaLabel}</span>`
                  : `<span class="tabbar-dot"></span><span>${item.label}</span>`
              }
            </button>
          `,
        )
        .join("")}
    </nav>
  `;
}

function toggleSet(targetSet, value) {
  if (targetSet.has(value)) {
    targetSet.delete(value);
    return;
  }

  targetSet.add(value);
}

function unlockButton(short) {
  const unlocked = state.purchasedShortIds.has(short.id);
  const durationLabel = short.duration.replace("分", "m");
  const ctaLabel = unlocked ? "Continue main" : "Unlock";
  const ctaMeta = unlocked ? short.progress : `${short.price} | ${durationLabel}`;

  return `
    <button
      class="unlock-cta ${unlocked ? "is-unlocked" : ""}"
      data-action="${unlocked ? "open-main" : "open-paywall"}"
      data-short-id="${short.id}"
      type="button"
    >
      <span class="unlock-title">${ctaLabel}</span>
      <span class="unlock-pill">${ctaMeta}</span>
    </button>
  `;
}

function unlockShort(shortId, replaceHistory = false) {
  state.hasPurchaseSetup = true;
  state.libraryShortIds.add(shortId);
  state.lastMainShortId = shortId;
  state.purchasedShortIds.add(shortId);

  navigate(
    {
      currentCreatorId: shorts[shortId].creatorId,
      currentShortId: shortId,
      screen: "main",
    },
    replaceHistory,
  );
}

function updatePaywallState() {
  const confirmButton = root.querySelector("[data-action='confirm-paywall']");

  if (!confirmButton) {
    return;
  }

  confirmButton.disabled = !(state.acceptAge && state.acceptTerms);
}

function updateSearchResults() {
  const query = state.searchQuery.trim().toLowerCase();
  const items = root.querySelectorAll("[data-search-item]");

  if (!items.length) {
    return;
  }

  items.forEach((item) => {
    const creator = creators[item.dataset.creatorId];
    const haystack = `${creator.name} ${creator.handle}`.toLowerCase();
    const matches = query === "" || haystack.includes(query);
    item.hidden = !matches;
  });
}

function updateLibraryResults() {
  const query = state.libraryQuery.trim().toLowerCase();
  const items = root.querySelectorAll("[data-library-item]");
  const countLabel = root.querySelector("[data-role='library-count-label']");
  let visibleCount = 0;

  if (!items.length) {
    return;
  }

  items.forEach((item) => {
    const haystack = item.dataset.searchText || "";
    const matches = query === "" || haystack.includes(query);
    item.hidden = !matches;

    if (matches) {
      visibleCount += 1;
    }
  });

  if (countLabel) {
    countLabel.textContent = query === "" ? `${visibleCount} items` : `${visibleCount} matches`;
  }
}

function updateFollowingResults() {
  const query = state.followingQuery.trim().toLowerCase();
  const items = root.querySelectorAll("[data-following-item]");
  const countLabel = root.querySelector("[data-role='following-count-label']");
  let visibleCount = 0;

  if (!items.length) {
    return;
  }

  items.forEach((item) => {
    const haystack = item.dataset.searchText || "";
    const matches = query === "" || haystack.includes(query);
    item.hidden = !matches;

    if (matches) {
      visibleCount += 1;
    }
  });

  if (countLabel) {
    countLabel.textContent = query === "" ? `${visibleCount} creators` : `${visibleCount} matches`;
  }
}
