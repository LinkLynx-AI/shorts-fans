import http from "node:http";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

const host = "127.0.0.1";
const port = 3201;

const fixturePath = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../docs/contracts/fixtures/fan-public-surfaces.json",
);
const viewerBootstrapFixturePath = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../docs/contracts/fixtures/viewer-bootstrap.json",
);
const creatorFollowFixturePath = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../docs/contracts/fixtures/fan-creator-follow.json",
);
const fanProfileFixturePath = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../docs/contracts/fixtures/fan-profile.json",
);

const fixtures = JSON.parse(await readFile(fixturePath, "utf8"));
const viewerBootstrapFixtures = JSON.parse(await readFile(viewerBootstrapFixturePath, "utf8"));
const creatorFollowFixtures = JSON.parse(await readFile(creatorFollowFixturePath, "utf8"));
const fanProfileFixtures = JSON.parse(await readFile(fanProfileFixturePath, "utf8"));
const searchFixtures = fixtures["GET /api/fan/creators/search"];
const creatorProfileHeaderFixtures = fixtures["GET /api/fan/creators/{creatorId}"];
const creatorProfileShortGridFixtures = fixtures["GET /api/fan/creators/{creatorId}/shorts"];
const creatorFollowPutFixtures = creatorFollowFixtures["PUT /api/fan/creators/{creatorId}/follow"];
const creatorFollowDeleteFixtures = creatorFollowFixtures["DELETE /api/fan/creators/{creatorId}/follow"];
const fanProfilePinnedShortFixtures = fanProfileFixtures["GET /api/fan/profile/pinned-shorts"];
const authenticatedCreatorBootstrap = viewerBootstrapFixtures.authenticatedCreator;
const authenticatedFanBootstrap = viewerBootstrapFixtures.authenticatedFan;
const unauthenticatedBootstrap = viewerBootstrapFixtures.unauthenticated;
const e2eSessionToken = "e2e-viewer-session";
const e2eCreatorSessionToken = "e2e-creator-session";
const existingFanEmail = "fan@example.com";
const existingFanPassword = "VeryStrongPass123!";
const passwordResetFanEmail = "resetfan@example.com";
const passwordResetFanPassword = "ResetPass123!";
const signUpConfirmationCode = "123456";
const passwordResetConfirmationCode = "654321";
const passwordByEmail = new Map([
  [existingFanEmail, existingFanPassword],
  [passwordResetFanEmail, passwordResetFanPassword],
]);
const pendingPasswordResetByEmail = new Set();
const pendingSignUpDraftByEmail = new Map();
const creatorSessionTokens = new Set([e2eCreatorSessionToken]);
const fanSessionTokens = new Set([e2eSessionToken]);
const sessionEmailByToken = new Map([
  [e2eSessionToken, existingFanEmail],
]);
const creatorBaseStatsById = {
  creator_aoi_n: {
    fanCount: 19000,
    shortCount: 2,
  },
  creator_mina_rei: {
    fanCount: 24000,
    shortCount: 2,
  },
  creator_sora_vale: {
    fanCount: 16000,
    shortCount: 0,
  },
};
const creatorSummaryById = {
  creator_aoi_n: {
    avatar: {
      durationSeconds: null,
      id: "asset_creator_aoi_avatar",
      kind: "image",
      posterUrl: null,
      url: "https://cdn.example.com/creator/aoi/avatar.jpg",
    },
    bio: "soft light と close framing の short を中心に更新中。",
    displayName: "Aoi N",
    handle: "@aoina",
    id: "creator_aoi_n",
  },
  creator_mina_rei: {
    avatar: {
      durationSeconds: null,
      id: "asset_creator_mina_avatar",
      kind: "image",
      posterUrl: null,
      url: "https://cdn.example.com/creator/mina/avatar.jpg",
    },
    bio: "quiet rooftop と hotel light の preview を軸に投稿。",
    displayName: "Mina Rei",
    handle: "@minarei",
    id: "creator_mina_rei",
  },
  creator_sora_vale: {
    avatar: {
      durationSeconds: null,
      id: "asset_creator_sora_avatar",
      kind: "image",
      posterUrl: null,
      url: "https://cdn.example.com/creator/sora/avatar.jpg",
    },
    bio: "after rain と balcony mood の short をまとめています。",
    displayName: "Sora Vale",
    handle: "@soravale",
    id: "creator_sora_vale",
  },
};
const defaultFollowedCreatorIds = ["creator_mina_rei"];
const followedCreatorIdsBySessionToken = new Map();
const viewerStateBySessionToken = new Map();
let issuedSessionCount = 0;

if (!searchFixtures) {
  throw new Error("creator search fixture が見つかりません");
}

const recentResponse = searchFixtures.search_recent;
const filteredResponse = searchFixtures.search_filtered;
const creatorProfileNotFoundResponse = creatorProfileHeaderFixtures?.creator_profile_header_not_found;
const creatorProfileShortGridNotFoundResponse = creatorProfileShortGridFixtures?.creator_profile_shorts_not_found;
const creatorProfileShortGridEmptyResponse = creatorProfileShortGridFixtures?.creator_profile_shorts_empty;
const creatorProfileShortGridNormalResponse = creatorProfileShortGridFixtures?.creator_profile_shorts_normal;

if (!recentResponse || !filteredResponse) {
  throw new Error("creator search fixture の recent / filtered が不足しています");
}

if (
  !creatorProfileNotFoundResponse ||
  !creatorProfileShortGridNotFoundResponse ||
  !creatorProfileShortGridEmptyResponse ||
  !creatorProfileShortGridNormalResponse
) {
  throw new Error("creator profile fixture が不足しています");
}

if (!authenticatedCreatorBootstrap || !authenticatedFanBootstrap || !unauthenticatedBootstrap) {
  throw new Error(
    "viewer bootstrap fixture の authenticatedCreator / authenticatedFan / unauthenticated が不足しています",
  );
}

if (!creatorFollowPutFixtures || !creatorFollowDeleteFixtures) {
  throw new Error("creator follow fixture が不足しています");
}

if (!fanProfilePinnedShortFixtures) {
  throw new Error("fan profile pinned shorts fixture が不足しています");
}

function buildCorsHeaders(request) {
  const origin = request.headers.origin;

  if (!origin) {
    return {
      "Access-Control-Allow-Headers": "Accept, Content-Type",
      "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
    };
  }

  return {
    "Access-Control-Allow-Credentials": "true",
    "Access-Control-Allow-Headers": "Accept, Content-Type",
    "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
    "Access-Control-Allow-Origin": origin,
    Vary: "Origin",
  };
}

function writeJson(request, response, statusCode, body) {
  response.writeHead(statusCode, {
    ...buildCorsHeaders(request),
    "Content-Type": "application/json; charset=utf-8",
  });
  response.end(JSON.stringify(body));
}

function writeNoContent(request, response, headers = {}) {
  response.writeHead(204, {
    ...buildCorsHeaders(request),
    ...headers,
  });
  response.end();
}

function buildSearchResponse(query) {
  if (query === "") {
    return recentResponse.body;
  }

  if (query.toLowerCase() === "mina") {
    return filteredResponse.body;
  }

  return {
    data: {
      items: [],
      query,
    },
    error: null,
    meta: {
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_creator_search_e2e_empty_001",
    },
  };
}

function isAuthenticatedCreatorSessionToken(sessionToken) {
  return typeof sessionToken === "string" && creatorSessionTokens.has(sessionToken);
}

function isAuthenticatedSessionToken(sessionToken) {
  return typeof sessionToken === "string" && (
    creatorSessionTokens.has(sessionToken) ||
    fanSessionTokens.has(sessionToken)
  );
}

function createE2ESessionToken() {
  issuedSessionCount += 1;
  return `${e2eSessionToken}-${issuedSessionCount}`;
}

function buildDefaultViewerState(sessionToken = null) {
  const bootstrapFixture = isAuthenticatedCreatorSessionToken(sessionToken)
    ? authenticatedCreatorBootstrap
    : authenticatedFanBootstrap;

  return structuredClone(bootstrapFixture.data.currentViewer);
}

function getViewerState(sessionToken) {
  if (!isAuthenticatedSessionToken(sessionToken)) {
    return null;
  }

  const existingViewerState = viewerStateBySessionToken.get(sessionToken);

  if (existingViewerState) {
    return existingViewerState;
  }

  const nextViewerState = buildDefaultViewerState(sessionToken);

  viewerStateBySessionToken.set(sessionToken, nextViewerState);

  return nextViewerState;
}

function buildViewerBootstrapResponse(sessionToken) {
  const viewerState = getViewerState(sessionToken);

  if (!viewerState) {
    return structuredClone(unauthenticatedBootstrap);
  }

  const bootstrapBody = structuredClone(
    viewerState.activeMode === "creator" ? authenticatedCreatorBootstrap : authenticatedFanBootstrap,
  );

  bootstrapBody.data.currentViewer = structuredClone(viewerState);

  return bootstrapBody;
}

function normalizeCreatorHandleInput(value) {
  if (typeof value !== "string") {
    return null;
  }

  const normalized = value.trim().replace(/^@/, "").toLowerCase();
  if (normalized === "" || !/^[a-z0-9._]+$/.test(normalized)) {
    return null;
  }

  return normalized;
}

function getFollowedCreatorIds(sessionToken) {
  if (!isAuthenticatedSessionToken(sessionToken)) {
    return null;
  }

  const existingFollowedCreatorIds = followedCreatorIdsBySessionToken.get(sessionToken);

  if (existingFollowedCreatorIds) {
    return existingFollowedCreatorIds;
  }

  const nextFollowedCreatorIds = new Set(defaultFollowedCreatorIds);

  followedCreatorIdsBySessionToken.set(sessionToken, nextFollowedCreatorIds);

  return nextFollowedCreatorIds;
}

function isCreatorFollowed(creatorId, sessionToken) {
  return getFollowedCreatorIds(sessionToken)?.has(creatorId) ?? false;
}

function resolveCreatorFanCount(creatorId, sessionToken) {
  const baseStats = creatorBaseStatsById[creatorId];

  if (!baseStats) {
    return null;
  }

  return baseStats.fanCount + (isCreatorFollowed(creatorId, sessionToken) ? 1 : 0);
}

function buildFanProfileOverviewResponse(sessionToken) {
  return {
    data: {
      fanProfile: {
        counts: {
          following: getFollowedCreatorIds(sessionToken)?.size ?? defaultFollowedCreatorIds.length,
          library: 2,
          pinnedShorts: 1,
        },
        title: "My archive",
      },
    },
    error: null,
    meta: {
      page: null,
      requestId: "req_e2e_fan_profile_overview_001",
    },
  };
}

function buildFanProfileFollowingResponse(sessionToken) {
  const items = Object.values(creatorSummaryById)
    .filter((creator) => isCreatorFollowed(creator.id, sessionToken))
    .map((creator) => ({
      creator,
      viewer: {
        isFollowing: true,
      },
    }));

  return {
    data: {
      items,
    },
    error: null,
    meta: {
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_e2e_fan_profile_following_001",
    },
  };
}

function buildFanProfilePinnedShortsResponse() {
  return fanProfilePinnedShortFixtures.pinned_populated.body;
}

function buildCreatorFollowMutationResponse(method, creatorId, sessionToken) {
  const authRequiredResponse =
    method === "PUT"
      ? creatorFollowPutFixtures.follow_auth_required
      : creatorFollowDeleteFixtures.unfollow_auth_required;

  if (!isAuthenticatedSessionToken(sessionToken)) {
    return authRequiredResponse;
  }

  if (!creatorBaseStatsById[creatorId]) {
    return method === "PUT"
      ? creatorFollowPutFixtures.follow_not_found
      : creatorFollowDeleteFixtures.unfollow_not_found;
  }

  const followedCreatorIds = getFollowedCreatorIds(sessionToken);

  if (!followedCreatorIds) {
    return authRequiredResponse;
  }

  if (method === "PUT") {
    const responseFixture = followedCreatorIds.has(creatorId)
      ? creatorFollowPutFixtures.follow_repeat
      : creatorFollowPutFixtures.follow_success;

    followedCreatorIds.add(creatorId);

    const body = structuredClone(responseFixture.body);

    body.data.stats.fanCount = resolveCreatorFanCount(creatorId, sessionToken);

    return {
      body,
      status: responseFixture.status,
    };
  }

  const responseFixture = followedCreatorIds.has(creatorId)
    ? creatorFollowDeleteFixtures.unfollow_success
    : creatorFollowDeleteFixtures.unfollow_repeat;

  followedCreatorIds.delete(creatorId);

  const body = structuredClone(responseFixture.body);

  body.data.stats.fanCount = resolveCreatorFanCount(creatorId, sessionToken);

  return {
    body,
    status: responseFixture.status,
  };
}

function buildCreatorProfileHeaderResponse(creatorId, sessionToken) {
  const resolvedFanCount = resolveCreatorFanCount(creatorId, sessionToken);

  if (creatorId === "creator_aoi_n") {
    return {
      status: 200,
      body: {
        data: {
          profile: {
            creator: {
              id: "creator_aoi_n",
              displayName: "Aoi N",
              handle: "@aoina",
              avatar: {
                id: "asset_creator_aoi_avatar",
                kind: "image",
                url: "https://cdn.example.com/creator/aoi/avatar.jpg",
                posterUrl: null,
                durationSeconds: null,
              },
              bio: "soft light と close framing の short を中心に更新中。",
            },
            stats: {
              shortCount: creatorBaseStatsById.creator_aoi_n.shortCount,
              fanCount: resolvedFanCount ?? creatorBaseStatsById.creator_aoi_n.fanCount,
            },
            viewer: {
              isFollowing: isCreatorFollowed(creatorId, sessionToken),
            },
          },
        },
        error: null,
        meta: {
          page: null,
          requestId: "req_creator_profile_header_aoi_001",
        },
      },
    };
  }

  if (creatorId === "creator_mina_rei") {
    const body = structuredClone(creatorProfileHeaderFixtures.creator_profile_header_normal.body);

    body.data.profile.stats.fanCount = resolvedFanCount ?? creatorBaseStatsById.creator_mina_rei.fanCount;
    body.data.profile.viewer.isFollowing = isCreatorFollowed(creatorId, sessionToken);
    return {
      body,
      status: 200,
    };
  }

  if (creatorId === "creator_sora_vale") {
    return {
      status: 200,
      body: {
        data: {
          profile: {
            creator: {
              id: "creator_sora_vale",
              displayName: "Sora Vale",
              handle: "@soravale",
              avatar: {
                id: "asset_creator_sora_avatar",
                kind: "image",
                url: "https://cdn.example.com/creator/sora/avatar.jpg",
                posterUrl: null,
                durationSeconds: null,
              },
              bio: "after rain と balcony mood の short をまとめています。",
            },
            stats: {
              shortCount: creatorBaseStatsById.creator_sora_vale.shortCount,
              fanCount: resolvedFanCount ?? creatorBaseStatsById.creator_sora_vale.fanCount,
            },
            viewer: {
              isFollowing: isCreatorFollowed(creatorId, sessionToken),
            },
          },
        },
        error: null,
        meta: {
          page: null,
          requestId: "req_creator_profile_header_sora_001",
        },
      },
    };
  }

  return creatorProfileNotFoundResponse;
}

function buildCreatorProfileShortGridResponse(creatorId) {
  if (creatorId === "creator_aoi_n") {
    return {
      status: 200,
      body: {
        data: {
          items: [
            {
              canonicalMainId: "main_aoi_blue_balcony",
              creatorId: "creator_aoi_n",
              id: "softlight",
              media: {
                durationSeconds: 18,
                id: "asset_short_aoi_softlight",
                kind: "video",
                posterUrl: "https://cdn.example.com/shorts/aoi-softlight-poster.jpg",
                url: "https://cdn.example.com/shorts/aoi-softlight.mp4",
              },
              previewDurationSeconds: 18,
            },
            {
              canonicalMainId: "main_aoi_blue_balcony",
              creatorId: "creator_aoi_n",
              id: "balcony",
              media: {
                durationSeconds: 15,
                id: "asset_short_aoi_balcony",
                kind: "video",
                posterUrl: "https://cdn.example.com/shorts/aoi-balcony-poster.jpg",
                url: "https://cdn.example.com/shorts/aoi-balcony.mp4",
              },
              previewDurationSeconds: 15,
            },
          ],
        },
        error: null,
        meta: {
          page: {
            hasNext: false,
            nextCursor: null,
          },
          requestId: "req_creator_profile_shorts_aoi_001",
        },
      },
    };
  }

  if (creatorId === "creator_mina_rei") {
    return creatorProfileShortGridNormalResponse;
  }

  if (creatorId === "creator_sora_vale") {
    return creatorProfileShortGridEmptyResponse;
  }

  return creatorProfileShortGridNotFoundResponse;
}

function readCookieValue(cookieHeader, cookieName) {
  if (!cookieHeader) {
    return null;
  }

  for (const cookiePart of cookieHeader.split(";")) {
    const trimmedCookiePart = cookiePart.trim();

    if (!trimmedCookiePart.startsWith(`${cookieName}=`)) {
      continue;
    }

    const value = trimmedCookiePart.slice(cookieName.length + 1).trim();

    if (value.length > 0) {
      return value;
    }
  }

  return null;
}

async function readJsonBody(request) {
  const chunks = [];

  for await (const chunk of request) {
    chunks.push(chunk);
  }

  const rawBody = Buffer.concat(chunks).toString("utf8");

  if (rawBody.length === 0) {
    return null;
  }

  try {
    return JSON.parse(rawBody);
  } catch {
    return null;
  }
}

function isValidEmail(value) {
  return typeof value === "string" && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}

function isValidDisplayName(value) {
  return typeof value === "string" && value.trim().length > 0;
}

function isValidHandle(value) {
  return typeof value === "string" && /^@?[a-z0-9._]+$/i.test(value);
}

function isValidPassword(value) {
  return typeof value === "string" && value.length >= 8;
}

function isAuthenticatedFanRequest(request) {
  return isAuthenticatedSessionToken(readCookieValue(request.headers.cookie, "shorts_fans_session"));
}

function writeAuthRequired(request, response, requestId, message = "fan profile requires authentication") {
  writeJson(request, response, 401, {
    data: null,
    error: {
      code: "auth_required",
      message,
    },
    meta: {
      page: null,
      requestId,
    },
  });
}

const server = http.createServer((request, response) => {
  const requestUrl = new URL(request.url ?? "/", `http://${host}:${port}`);

  if (request.method === "OPTIONS") {
    response.writeHead(204, {
      ...buildCorsHeaders(request),
    });
    response.end();
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/healthz") {
    writeJson(request, response, 200, { status: "ok" });
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/fan/creators/search") {
    const query = requestUrl.searchParams.get("q")?.trim() ?? "";
    writeJson(request, response, 200, buildSearchResponse(query));
    return;
  }

  if (
    (request.method === "PUT" || request.method === "DELETE") &&
    requestUrl.pathname.startsWith("/api/fan/creators/")
  ) {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");
    const pathnameParts = requestUrl.pathname.split("/").filter(Boolean);
    const creatorId = pathnameParts[3];
    const lastSegment = pathnameParts.at(-1);

    if (!creatorId || lastSegment !== "follow") {
      writeJson(request, response, 404, creatorProfileNotFoundResponse.body);
      return;
    }

    const followMutationResponse = buildCreatorFollowMutationResponse(
      request.method,
      creatorId,
      sessionToken,
    );

    writeJson(request, response, followMutationResponse.status, followMutationResponse.body);
    return;
  }

  if (request.method === "GET" && requestUrl.pathname.startsWith("/api/fan/creators/")) {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");
    const pathnameParts = requestUrl.pathname.split("/").filter(Boolean);
    const creatorId = pathnameParts[3];
    const lastSegment = pathnameParts.at(-1);

    if (!creatorId) {
      writeJson(request, response, 404, creatorProfileNotFoundResponse.body);
      return;
    }

    if (lastSegment === "shorts") {
      const shortGridResponse = buildCreatorProfileShortGridResponse(creatorId);

      writeJson(request, response, shortGridResponse.status, shortGridResponse.body);
      return;
    }

    const headerResponse = buildCreatorProfileHeaderResponse(creatorId, sessionToken);

    writeJson(request, response, headerResponse.status, headerResponse.body);
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/viewer/bootstrap") {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");
    const body = buildViewerBootstrapResponse(sessionToken);

    writeJson(request, response, 200, body);
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/viewer/creator-registration") {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");

    if (!isAuthenticatedSessionToken(sessionToken)) {
      writeAuthRequired(
        request,
        response,
        "req_e2e_creator_registration_auth_required_001",
        "creator registration requires authentication",
      );
      return;
    }

    void readJsonBody(request).then((body) => {
      if (typeof body?.displayName !== "string" || body.displayName.trim() === "") {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_display_name",
            message: "display name is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_creator_registration_invalid_display_name_001",
          },
        });
        return;
      }
      if (normalizeCreatorHandleInput(body?.handle) === null) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_handle",
            message: "handle is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_creator_registration_invalid_handle_001",
          },
        });
        return;
      }

      const viewerState = getViewerState(sessionToken);

      if (viewerState) {
        viewerState.canAccessCreatorMode = true;
        viewerState.activeMode = "fan";
      }

      writeNoContent(request, response);
    });
    return;
  }

  if (request.method === "PUT" && requestUrl.pathname === "/api/viewer/active-mode") {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");

    if (!isAuthenticatedSessionToken(sessionToken)) {
      writeAuthRequired(
        request,
        response,
        "req_e2e_viewer_active_mode_auth_required_001",
        "viewer mode switch requires authentication",
      );
      return;
    }

    void readJsonBody(request).then((body) => {
      const nextActiveMode = body?.activeMode;

      if (nextActiveMode !== "fan" && nextActiveMode !== "creator") {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_active_mode",
            message: "active mode is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_viewer_active_mode_invalid_001",
          },
        });
        return;
      }

      const viewerState = getViewerState(sessionToken);

      if (!viewerState) {
        writeAuthRequired(
          request,
          response,
          "req_e2e_viewer_active_mode_auth_required_002",
          "viewer mode switch requires authentication",
        );
        return;
      }

      if (nextActiveMode === "creator" && !viewerState.canAccessCreatorMode) {
        writeJson(request, response, 403, {
          data: null,
          error: {
            code: "creator_mode_unavailable",
            message: "creator mode is not available",
          },
          meta: {
            page: null,
            requestId: "req_e2e_viewer_active_mode_unavailable_001",
          },
        });
        return;
      }

      viewerState.activeMode = nextActiveMode;

      writeNoContent(request, response);
    });
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/fan/profile") {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");

    if (!isAuthenticatedFanRequest(request)) {
      writeAuthRequired(request, response, "req_e2e_fan_profile_auth_required_001");
      return;
    }

    writeJson(request, response, 200, buildFanProfileOverviewResponse(sessionToken));
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/fan/profile/following") {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");

    if (!isAuthenticatedFanRequest(request)) {
      writeAuthRequired(request, response, "req_e2e_fan_profile_following_auth_required_001");
      return;
    }

    writeJson(request, response, 200, buildFanProfileFollowingResponse(sessionToken));
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/fan/profile/pinned-shorts") {
    if (!isAuthenticatedFanRequest(request)) {
      writeAuthRequired(request, response, "req_e2e_fan_profile_pinned_auth_required_001");
      return;
    }

    writeJson(request, response, 200, buildFanProfilePinnedShortsResponse());
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/sign-in") {
    void readJsonBody(request).then((body) => {
      const email = body?.email;
      const password = body?.password;

      if (!isValidEmail(email)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_email",
            message: "email is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_in_invalid_email_001",
          },
        });
        return;
      }

      if (!isValidPassword(password)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_password",
            message: "password is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_in_invalid_password_001",
          },
        });
        return;
      }

      if (email === "pendingfan@example.com") {
        writeJson(request, response, 403, {
          data: null,
          error: {
            code: "confirmation_required",
            message: "email confirmation is required",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_in_confirmation_required_001",
          },
        });
        return;
      }

      const expectedPassword = typeof email === "string" ? passwordByEmail.get(email) : null;

      if (!expectedPassword || password !== expectedPassword) {
        writeJson(request, response, 401, {
          data: null,
          error: {
            code: "invalid_credentials",
            message: "email or password is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_in_invalid_credentials_001",
          },
        });
        return;
      }

      const nextSessionToken = createE2ESessionToken();

      fanSessionTokens.add(nextSessionToken);
      viewerStateBySessionToken.set(nextSessionToken, buildDefaultViewerState());
      sessionEmailByToken.set(nextSessionToken, email);

      writeNoContent(request, response, {
        "Set-Cookie": `shorts_fans_session=${nextSessionToken}; Path=/; HttpOnly; SameSite=Lax`,
      });
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/sign-up") {
    void readJsonBody(request).then((body) => {
      if (!isValidEmail(body?.email)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_email",
            message: "email is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_invalid_email_001",
          },
        });
        return;
      }

      if (!isValidDisplayName(body?.displayName)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_display_name",
            message: "display name is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_invalid_display_name_001",
          },
        });
        return;
      }

      if (!isValidHandle(body?.handle)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_handle",
            message: "handle is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_invalid_handle_001",
          },
        });
        return;
      }

      if (!isValidPassword(body?.password)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_password",
            message: "password is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_invalid_password_001",
          },
        });
        return;
      }

      if (typeof body.email === "string") {
        pendingSignUpDraftByEmail.set(body.email, {
          displayName: body.displayName,
          handle: body.handle,
          password: body.password,
        });
      }

      writeJson(request, response, 200, {
        data: {
          deliveryDestinationHint: "f***@example.com",
          nextStep: "confirm_sign_up",
        },
        error: null,
        meta: {
          page: null,
          requestId: "req_e2e_sign_up_accepted_001",
        },
      });
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/sign-up/confirm") {
    void readJsonBody(request).then((body) => {
      if (!isValidEmail(body?.email)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_email",
            message: "email is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_confirm_invalid_email_001",
          },
        });
        return;
      }

      if (body?.confirmationCode === "000000") {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "confirmation_code_expired",
            message: "confirmation code has expired",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_confirm_expired_code_001",
          },
        });
        return;
      }

      if (body?.confirmationCode !== signUpConfirmationCode) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_confirmation_code",
            message: "confirmation code is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_confirm_invalid_code_001",
          },
        });
        return;
      }

      if (typeof body.email !== "string" || !pendingSignUpDraftByEmail.has(body.email)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_confirmation_code",
            message: "confirmation code is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_confirm_missing_draft_001",
          },
        });
        return;
      }

      const nextSessionToken = createE2ESessionToken();
      const pendingDraft =
        typeof body.email === "string" ? pendingSignUpDraftByEmail.get(body.email) : undefined;

      if (
        typeof body.email === "string" &&
        pendingDraft &&
        typeof pendingDraft.displayName === "string" &&
        typeof pendingDraft.handle === "string" &&
        typeof pendingDraft.password === "string"
      ) {
        passwordByEmail.set(body.email, pendingDraft.password);
        pendingSignUpDraftByEmail.delete(body.email);
      }

      fanSessionTokens.add(nextSessionToken);
      viewerStateBySessionToken.set(nextSessionToken, buildDefaultViewerState());
      if (typeof body.email === "string") {
        sessionEmailByToken.set(nextSessionToken, body.email);
      }

      writeNoContent(request, response, {
        "Set-Cookie": `shorts_fans_session=${nextSessionToken}; Path=/; HttpOnly; SameSite=Lax`,
      });
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/password-reset") {
    void readJsonBody(request).then((body) => {
      if (!isValidEmail(body?.email)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_email",
            message: "email is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_password_reset_invalid_email_001",
          },
        });
        return;
      }

      if (typeof body.email === "string") {
        pendingPasswordResetByEmail.add(body.email);
      }

      writeJson(request, response, 200, {
        data: {
          deliveryDestinationHint: "f***@example.com",
          nextStep: "confirm_password_reset",
        },
        error: null,
        meta: {
          page: null,
          requestId: "req_e2e_password_reset_accepted_001",
        },
      });
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/password-reset/confirm") {
    void readJsonBody(request).then((body) => {
      if (body?.confirmationCode !== passwordResetConfirmationCode) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_confirmation_code",
            message: "confirmation code is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_password_reset_confirm_invalid_code_001",
          },
        });
        return;
      }

      if (!isValidPassword(body?.newPassword)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "password_policy_violation",
            message: "password does not satisfy the policy",
          },
          meta: {
            page: null,
            requestId: "req_e2e_password_reset_confirm_policy_001",
          },
        });
        return;
      }

      if (typeof body?.email !== "string" || !pendingPasswordResetByEmail.has(body.email)) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_confirmation_code",
            message: "confirmation code is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_password_reset_confirm_missing_request_001",
          },
        });
        return;
      }

      pendingPasswordResetByEmail.delete(body.email);

      if (passwordByEmail.has(body.email)) {
        passwordByEmail.set(body.email, body.newPassword);
      }

      writeNoContent(request, response);
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/re-auth") {
    void readJsonBody(request).then((body) => {
      if (!isAuthenticatedFanRequest(request)) {
        writeAuthRequired(request, response, "req_e2e_reauth_auth_required_001", "fan re-auth requires authentication");
        return;
      }

      const currentSessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");
      const sessionEmail =
        typeof currentSessionToken === "string" ? sessionEmailByToken.get(currentSessionToken) : undefined;
      const expectedPassword =
        typeof sessionEmail === "string" ? passwordByEmail.get(sessionEmail) : undefined;

      if (!expectedPassword || body?.password !== expectedPassword) {
        writeJson(request, response, 401, {
          data: null,
          error: {
            code: "invalid_credentials",
            message: "password is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_reauth_invalid_credentials_001",
          },
        });
        return;
      }

      const rotatedSessionToken = createE2ESessionToken();
      const currentViewerState = getViewerState(currentSessionToken);

      fanSessionTokens.add(rotatedSessionToken);
      if (currentViewerState) {
        viewerStateBySessionToken.set(rotatedSessionToken, structuredClone(currentViewerState));
      }
      if (sessionEmail) {
        sessionEmailByToken.set(rotatedSessionToken, sessionEmail);
      }

      writeNoContent(request, response, {
        "Set-Cookie": `shorts_fans_session=${rotatedSessionToken}; Path=/; HttpOnly; SameSite=Lax`,
      });
    });
    return;
  }

  writeJson(request, response, 404, {
    data: null,
    error: {
      code: "not_found",
      message: "fixture not found",
    },
    meta: {
      page: null,
      requestId: "req_e2e_not_found_001",
    },
  });
});

server.listen(port, host, () => {
  process.stdout.write(`mock e2e api listening on http://${host}:${port}\n`);
});
