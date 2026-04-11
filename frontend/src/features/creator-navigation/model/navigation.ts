import type { FanHubTab } from "@/entities/fan-profile";
import type { FeedTab, ShortId } from "@/entities/short";

export type CreatorProfileRouteOrigin =
  | {
      creatorDisplayName?: string | undefined;
      creatorHandle?: `@${string}` | undefined;
      from: "feed";
      tab?: FeedTab | undefined;
    }
  | {
      creatorDisplayName?: string | undefined;
      creatorHandle?: `@${string}` | undefined;
      from: "search";
      q?: string | undefined;
    }
  | {
      creatorDisplayName?: string | undefined;
      creatorHandle?: `@${string}` | undefined;
      from: "short";
      shortFanTab?: FanHubTab | undefined;
      shortId: ShortId;
    };

export type CreatorProfileRouteState = {
  creatorDisplayName?: string | undefined;
  creatorHandle?: `@${string}` | undefined;
  from?: CreatorProfileRouteOrigin["from"] | undefined;
  q?: string | undefined;
  shortFanTab?: FanHubTab | undefined;
  shortId?: ShortId | undefined;
  tab?: FeedTab | undefined;
};

export type CreatorShortDetailRouteState = {
  creatorId?: string | undefined;
  fanTab?: FanHubTab | undefined;
  from?: "creator" | "fan" | undefined;
  profileFrom?: CreatorProfileRouteOrigin["from"] | undefined;
  profileQ?: string | undefined;
  profileShortFanTab?: FanHubTab | undefined;
  profileShortId?: ShortId | undefined;
  profileTab?: FeedTab | undefined;
};

function buildQueryString(
  params: Record<string, string | undefined>,
): string {
  const searchParams = new URLSearchParams();

  for (const [key, value] of Object.entries(params)) {
    if (!value) {
      continue;
    }

    searchParams.set(key, value);
  }

  const queryString = searchParams.toString();

  return queryString.length > 0 ? `?${queryString}` : "";
}

function isValidCreatorRouteId(creatorId: string): boolean {
  return /^[A-Za-z0-9_]+$/.test(creatorId);
}

/**
 * creator profile への href を組み立てる。
 */
export function buildCreatorProfileHref(
  creatorId: string,
  origin: CreatorProfileRouteOrigin,
): string {
  if (origin.from === "search") {
    return `/creators/${creatorId}${buildQueryString({
      creatorDisplayName: origin.creatorDisplayName,
      creatorHandle: origin.creatorHandle,
      from: origin.from,
      q: origin.q?.trim() || undefined,
    })}`;
  }

  if (origin.from === "feed") {
    return `/creators/${creatorId}${buildQueryString({
      from: origin.from,
      tab: origin.tab,
    })}`;
  }

  return `/creators/${creatorId}${buildQueryString({
    from: origin.from,
    shortFanTab: origin.shortFanTab,
    shortId: origin.shortId,
  })}`;
}

/**
 * creator profile の戻り先 href を解決する。
 */
export function resolveCreatorProfileBackHref(
  routeState: CreatorProfileRouteState,
): string {
  if (routeState.from === "search") {
    return `/search${buildQueryString({ q: routeState.q?.trim() || undefined })}`;
  }

  if (routeState.from === "feed") {
    return `/${buildQueryString({ tab: routeState.tab })}`;
  }

  if (routeState.from === "short" && routeState.shortId) {
    return routeState.shortFanTab
      ? buildFanProfileShortDetailHref(routeState.shortId, routeState.shortFanTab)
      : `/shorts/${routeState.shortId}`;
  }

  return "/";
}

/**
 * creator profile 内の short tile 用 href を組み立てる。
 */
export function buildCreatorShortDetailHref(
  shortId: ShortId,
  creatorId: string,
  routeState: CreatorProfileRouteState,
): string {
  return `/shorts/${shortId}${buildQueryString({
    creatorId,
    from: "creator",
    profileFrom: routeState.from,
    profileQ: routeState.q?.trim() || undefined,
    profileShortFanTab: routeState.shortFanTab,
    profileShortId: routeState.shortId,
    profileTab: routeState.tab,
  })}`;
}

/**
 * fan profile 内の pinned short tile 用 href を組み立てる。
 */
export function buildFanProfileShortDetailHref(
  shortId: ShortId,
  fanTab: FanHubTab = "pinned",
): string {
  return `/shorts/${shortId}${buildQueryString({
    fanTab,
    from: "fan",
  })}`;
}

/**
 * short detail の戻り先 href を解決する。
 */
export function resolveShortDetailBackHref(
  routeState: CreatorShortDetailRouteState,
): string {
  if (routeState.from === "fan") {
    return `/fan${buildQueryString({
      tab: routeState.fanTab ?? "pinned",
    })}`;
  }

  if (routeState.from !== "creator" || !routeState.creatorId || !isValidCreatorRouteId(routeState.creatorId)) {
    return "/";
  }

  if (!routeState.profileFrom) {
    return `/creators/${routeState.creatorId}`;
  }

  if (routeState.profileFrom === "short") {
    return routeState.profileShortId
      ? buildCreatorProfileHref(routeState.creatorId, {
          from: "short",
          shortFanTab: routeState.profileShortFanTab,
          shortId: routeState.profileShortId,
        })
      : `/creators/${routeState.creatorId}`;
  }

  if (routeState.profileFrom === "feed") {
    return buildCreatorProfileHref(routeState.creatorId, {
      from: "feed",
      tab: routeState.profileTab,
    });
  }

  return buildCreatorProfileHref(routeState.creatorId, {
    from: "search",
    q: routeState.profileQ,
  });
}
