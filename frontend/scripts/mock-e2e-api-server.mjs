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
const authenticatedFanBootstrap = viewerBootstrapFixtures.authenticatedFan;
const unauthenticatedBootstrap = viewerBootstrapFixtures.unauthenticated;
const e2eSessionToken = "e2e-viewer-session";
const existingFanEmail = "fan@example.com";
const signInChallengeToken = "e2e-sign-in-challenge";
const signUpChallengeToken = "e2e-sign-up-challenge";
const fanProfileOverviewResponse = {
  data: {
    fanProfile: {
      counts: {
        following: 3,
        library: 3,
        pinnedShorts: 3,
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
const fanProfileFollowingResponse = {
  data: {
    items: [
      {
        creator: {
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
        viewer: {
          isFollowing: true,
        },
      },
      {
        creator: {
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
        viewer: {
          isFollowing: true,
        },
      },
      {
        creator: {
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
        viewer: {
          isFollowing: true,
        },
      },
    ],
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

if (!searchFixtures) {
  throw new Error("creator search fixture が見つかりません");
}

const recentResponse = searchFixtures.search_recent;
const filteredResponse = searchFixtures.search_filtered;

if (!recentResponse || !filteredResponse) {
  throw new Error("creator search fixture の recent / filtered が不足しています");
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

function isAuthenticatedFanRequest(request) {
  return readCookieValue(request.headers.cookie, "shorts_fans_session") === e2eSessionToken;
}

function writeAuthRequired(request, response, requestId) {
  writeJson(request, response, 401, {
    data: null,
    error: {
      code: "auth_required",
      message: "fan profile requires authentication",
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

  if (request.method === "GET" && requestUrl.pathname === "/api/viewer/bootstrap") {
    const sessionToken = readCookieValue(request.headers.cookie, "shorts_fans_session");
    const body =
      sessionToken === e2eSessionToken ? authenticatedFanBootstrap : unauthenticatedBootstrap;

    writeJson(request, response, 200, body);
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/fan/profile") {
    if (!isAuthenticatedFanRequest(request)) {
      writeAuthRequired(request, response, "req_e2e_fan_profile_auth_required_001");
      return;
    }

    writeJson(request, response, 200, fanProfileOverviewResponse);
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/fan/profile/following") {
    if (!isAuthenticatedFanRequest(request)) {
      writeAuthRequired(request, response, "req_e2e_fan_profile_following_auth_required_001");
      return;
    }

    writeJson(request, response, 200, fanProfileFollowingResponse);
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
