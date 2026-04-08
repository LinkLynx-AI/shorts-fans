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

const fixtures = JSON.parse(await readFile(fixturePath, "utf8"));
const viewerBootstrapFixtures = JSON.parse(await readFile(viewerBootstrapFixturePath, "utf8"));
const searchFixtures = fixtures["GET /api/fan/creators/search"];
const creatorProfileHeaderFixtures = fixtures["GET /api/fan/creators/{creatorId}"];
const creatorProfileShortGridFixtures = fixtures["GET /api/fan/creators/{creatorId}/shorts"];
const authenticatedFanBootstrap = viewerBootstrapFixtures.authenticatedFan;
const unauthenticatedBootstrap = viewerBootstrapFixtures.unauthenticated;
const e2eSessionToken = "e2e-viewer-session";
const existingFanEmail = "fan@example.com";
const signInChallengeToken = "e2e-sign-in-challenge";
const signUpChallengeToken = "e2e-sign-up-challenge";

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

if (!authenticatedFanBootstrap || !unauthenticatedBootstrap) {
  throw new Error("viewer bootstrap fixture の authenticatedFan / unauthenticated が不足しています");
}

function buildCorsHeaders(request) {
  const origin = request.headers.origin;

  if (!origin) {
    return {
      "Access-Control-Allow-Headers": "Accept, Content-Type",
      "Access-Control-Allow-Methods": "GET, POST, DELETE, OPTIONS",
    };
  }

  return {
    "Access-Control-Allow-Credentials": "true",
    "Access-Control-Allow-Headers": "Accept, Content-Type",
    "Access-Control-Allow-Methods": "GET, POST, DELETE, OPTIONS",
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

function buildCreatorProfileHeaderResponse(creatorId, sessionToken) {
  const isAuthenticated = sessionToken === e2eSessionToken;

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
              shortCount: 2,
              fanCount: 19000,
            },
            viewer: {
              isFollowing: isAuthenticated,
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

    body.data.profile.viewer.isFollowing = isAuthenticated;
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
              shortCount: 0,
              fanCount: 16000,
            },
            viewer: {
              isFollowing: false,
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
    const body =
      sessionToken === e2eSessionToken ? authenticatedFanBootstrap : unauthenticatedBootstrap;

    writeJson(request, response, 200, body);
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/sign-in/challenges") {
    void readJsonBody(request).then((body) => {
      const email = body?.email;

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

      if (email !== existingFanEmail) {
        writeJson(request, response, 404, {
          data: null,
          error: {
            code: "email_not_found",
            message: "email was not found",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_in_email_not_found_001",
          },
        });
        return;
      }

      writeJson(request, response, 200, {
        data: {
          challengeToken: signInChallengeToken,
          expiresAt: "2026-04-07T12:00:00Z",
        },
        error: null,
        meta: {
          page: null,
          requestId: "req_e2e_sign_in_challenge_001",
        },
      });
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/sign-up/challenges") {
    void readJsonBody(request).then((body) => {
      const email = body?.email;

      if (!isValidEmail(email)) {
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

      if (email === existingFanEmail) {
        writeJson(request, response, 409, {
          data: null,
          error: {
            code: "email_already_registered",
            message: "email is already registered",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_conflict_001",
          },
        });
        return;
      }

      writeJson(request, response, 200, {
        data: {
          challengeToken: signUpChallengeToken,
          expiresAt: "2026-04-07T12:00:00Z",
        },
        error: null,
        meta: {
          page: null,
          requestId: "req_e2e_sign_up_challenge_001",
        },
      });
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/sign-in/session") {
    void readJsonBody(request).then((body) => {
      if (body?.email !== existingFanEmail) {
        writeJson(request, response, 404, {
          data: null,
          error: {
            code: "email_not_found",
            message: "email was not found",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_in_session_missing_email_001",
          },
        });
        return;
      }

      if (body?.challengeToken !== signInChallengeToken) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_challenge",
            message: "challenge is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_in_invalid_challenge_001",
          },
        });
        return;
      }

      writeNoContent(request, response, {
        "Set-Cookie": `shorts_fans_session=${e2eSessionToken}; Path=/; HttpOnly; SameSite=Lax`,
      });
    });
    return;
  }

  if (request.method === "POST" && requestUrl.pathname === "/api/fan/auth/sign-up/session") {
    void readJsonBody(request).then((body) => {
      if (body?.email === existingFanEmail) {
        writeJson(request, response, 409, {
          data: null,
          error: {
            code: "email_already_registered",
            message: "email is already registered",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_session_conflict_001",
          },
        });
        return;
      }

      if (body?.challengeToken !== signUpChallengeToken) {
        writeJson(request, response, 400, {
          data: null,
          error: {
            code: "invalid_challenge",
            message: "challenge is invalid",
          },
          meta: {
            page: null,
            requestId: "req_e2e_sign_up_invalid_challenge_001",
          },
        });
        return;
      }

      writeNoContent(request, response, {
        "Set-Cookie": `shorts_fans_session=${e2eSessionToken}; Path=/; HttpOnly; SameSite=Lax`,
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
