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

const fixtures = JSON.parse(await readFile(fixturePath, "utf8"));
const searchFixtures = fixtures["GET /api/fan/creators/search"];

if (!searchFixtures) {
  throw new Error("creator search fixture が見つかりません");
}

const recentResponse = searchFixtures.search_recent;
const filteredResponse = searchFixtures.search_filtered;

if (!recentResponse || !filteredResponse) {
  throw new Error("creator search fixture の recent / filtered が不足しています");
}

function writeJson(response, statusCode, body) {
  response.writeHead(statusCode, {
    "Access-Control-Allow-Credentials": "true",
    "Access-Control-Allow-Headers": "Accept, Content-Type",
    "Access-Control-Allow-Methods": "GET, OPTIONS",
    "Access-Control-Allow-Origin": "*",
    "Content-Type": "application/json; charset=utf-8",
  });
  response.end(JSON.stringify(body));
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

const server = http.createServer((request, response) => {
  const requestUrl = new URL(request.url ?? "/", `http://${host}:${port}`);

  if (request.method === "OPTIONS") {
    response.writeHead(204, {
      "Access-Control-Allow-Credentials": "true",
      "Access-Control-Allow-Headers": "Accept, Content-Type",
      "Access-Control-Allow-Methods": "GET, OPTIONS",
      "Access-Control-Allow-Origin": "*",
    });
    response.end();
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/healthz") {
    writeJson(response, 200, { status: "ok" });
    return;
  }

  if (request.method === "GET" && requestUrl.pathname === "/api/fan/creators/search") {
    const query = requestUrl.searchParams.get("q")?.trim() ?? "";
    writeJson(response, 200, buildSearchResponse(query));
    return;
  }

  writeJson(response, 404, {
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
