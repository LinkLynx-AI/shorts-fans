import type { NextRequest } from "next/server";

import {
  adminUiAccessCookieName,
  adminUiAccessHeaderName,
  isAdminUiSessionCookieMatch,
} from "../_lib/admin-ui-access";
import { GET, POST } from "./route";

describe("admin access route", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("sets an opaque httpOnly admin UI cookie for a POST header access token", () => {
    vi.stubEnv("NODE_ENV", "development");
    vi.stubEnv("ADMIN_UI_ENABLED", "1");
    vi.stubEnv("ADMIN_UI_ACCESS_TOKEN", "admin-ui-secret");

    const response = POST(new Request(
      "http://127.0.0.1:3002/admin/access?next=/admin/creator-reviews",
      {
        headers: {
          [adminUiAccessHeaderName]: "admin-ui-secret",
        },
      },
    ) as NextRequest);
    const setCookie = response.headers.get("set-cookie") ?? "";
    const cookieValue = setCookie.match(new RegExp(`${adminUiAccessCookieName}=([^;]+)`))?.[1] ?? "";

    expect(response.status).toBe(303);
    expect(response.headers.get("location")).toBe("http://127.0.0.1:3002/admin/creator-reviews");
    expect(response.headers.get("cache-control")).toBe("no-store");
    expect(response.headers.get("referrer-policy")).toBe("no-referrer");
    expect(setCookie).toContain(`${adminUiAccessCookieName}=`);
    expect(setCookie).not.toContain("admin-ui-secret");
    expect(setCookie).toContain("HttpOnly");
    expect(setCookie.toLowerCase()).toContain("samesite=strict");
    expect(isAdminUiSessionCookieMatch(decodeURIComponent(cookieValue))).toBe(true);
  });

  it("rejects GET query-string tokens without issuing a cookie", () => {
    vi.stubEnv("NODE_ENV", "development");
    vi.stubEnv("ADMIN_UI_ENABLED", "1");
    vi.stubEnv("ADMIN_UI_ACCESS_TOKEN", "admin-ui-secret");

    const response = GET();

    expect(response.status).toBe(404);
    expect(response.headers.get("cache-control")).toBe("no-store");
    expect(response.headers.get("referrer-policy")).toBe("no-referrer");
    expect(response.headers.get("set-cookie")).toBeNull();
  });

  it("rejects invalid POST header tokens without issuing a cookie", () => {
    vi.stubEnv("NODE_ENV", "development");
    vi.stubEnv("ADMIN_UI_ENABLED", "1");
    vi.stubEnv("ADMIN_UI_ACCESS_TOKEN", "admin-ui-secret");

    const response = POST(new Request("http://example.com/admin/access?next=/admin/creator-reviews", {
      headers: {
        [adminUiAccessHeaderName]: "wrong",
      },
    }) as NextRequest);

    expect(response.status).toBe(404);
    expect(response.headers.get("cache-control")).toBe("no-store");
    expect(response.headers.get("referrer-policy")).toBe("no-referrer");
    expect(response.headers.get("set-cookie")).toBeNull();
  });
});
