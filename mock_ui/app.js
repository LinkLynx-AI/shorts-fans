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
  mina: ["rooftop", "mirror"],
  sora: ["afterrain", "poolcut"],
};

const feedShortByTab = {
  following: "softlight",
  recommended: "rooftop",
};

const state = {
  acceptAge: false,
  acceptTerms: false,
  currentCreatorId: "mina",
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

  const { action, creatorId, shortId, tab } = actionButton.dataset;

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

function closeOverlay() {
  state.acceptAge = false;
  state.acceptTerms = false;
  state.overlay = null;
  state.overlayShortId = null;
  state.searchQuery = "";
  render();
}

function createBackAction() {
  return `<button aria-label="Back" class="back-button" data-action="back" type="button">&lt;</button>`;
}

function currentShort() {
  return shorts[state.currentShortId];
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

          <button class="fan-creator-button" type="button">creatorになる</button>

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
  if (state.overlay === "paywall") {
    return paywallOverlay();
  }

  return "";
}

function restore(snapshotState) {
  state.acceptAge = false;
  state.acceptTerms = false;
  state.overlay = null;
  state.overlayShortId = null;
  state.fanTab = snapshotState.fanTab || "pinned";
  state.followingQuery = snapshotState.followingQuery || "";
  state.searchQuery = snapshotState.searchQuery;
  state.rootTab = snapshotState.rootTab;
  state.screen = snapshotState.screen;
  state.feedTab = snapshotState.feedTab;
  state.libraryQuery = snapshotState.libraryQuery;
  state.currentShortId = snapshotState.currentShortId;
  state.currentCreatorId = snapshotState.currentCreatorId;
  render();
}

function screenMarkup() {
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
