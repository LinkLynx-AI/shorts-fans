import {
  adminUiAccessCookieName,
  adminUiSessionCookieMaxAgeSeconds,
  createAdminUiSessionCookieValue,
  getCookieValue,
  isAdminUiEnabled,
  isAdminUiAccessTokenMatch,
  isAdminUiRequestAllowed,
  isAdminUiSessionCookieMatch,
} from "./admin-ui-access";

describe("admin ui access", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.useRealTimers();
  });

  it("returns true only when the dedicated admin env flag is enabled in development", () => {
    vi.stubEnv("NODE_ENV", "development");
    vi.stubEnv("ADMIN_UI_ENABLED", "1");
    expect(isAdminUiEnabled()).toBe(true);

    vi.stubEnv("NODE_ENV", "production");
    expect(isAdminUiEnabled()).toBe(false);

    vi.stubEnv("NODE_ENV", "development");
    vi.stubEnv("ADMIN_UI_ENABLED", "0");
    expect(isAdminUiEnabled()).toBe(false);

    vi.unstubAllEnvs();
    expect(isAdminUiEnabled()).toBe(false);
  });

  it("allows admin requests only with a signed session cookie when enabled", () => {
    vi.stubEnv("NODE_ENV", "development");
    vi.stubEnv("ADMIN_UI_ENABLED", "1");
    vi.stubEnv("ADMIN_UI_ACCESS_TOKEN", "admin-ui-secret");
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-04-21T00:00:00.000Z"));
    const sessionCookie = createAdminUiSessionCookieValue();

    expect(isAdminUiAccessTokenMatch("admin-ui-secret")).toBe(true);
    expect(isAdminUiAccessTokenMatch("wrong")).toBe(false);
    expect(isAdminUiAccessTokenMatch("")).toBe(false);
    expect(isAdminUiSessionCookieMatch(sessionCookie)).toBe(true);
    expect(isAdminUiSessionCookieMatch("admin-ui-secret")).toBe(false);
    expect(isAdminUiSessionCookieMatch("wrong")).toBe(false);

    expect(getCookieValue(
      `other=value; ${adminUiAccessCookieName}=${sessionCookie}`,
      adminUiAccessCookieName,
    )).toBe(sessionCookie);

    expect(isAdminUiRequestAllowed({
      uiSessionCookie: sessionCookie,
    })).toBe(true);

    vi.setSystemTime(new Date("2026-04-21T08:00:01.000Z"));
    expect(isAdminUiSessionCookieMatch(sessionCookie)).toBe(false);
    expect(isAdminUiRequestAllowed({
      uiSessionCookie: sessionCookie,
    })).toBe(false);

    vi.setSystemTime(new Date("2026-04-21T00:00:00.000Z"));
    expect(isAdminUiRequestAllowed({
      uiSessionCookie: "",
    })).toBe(false);
    expect(isAdminUiRequestAllowed({
      uiSessionCookie: "wrong",
    })).toBe(false);

    vi.stubEnv("ADMIN_UI_ENABLED", "0");
    expect(isAdminUiRequestAllowed({
      uiSessionCookie: sessionCookie,
    })).toBe(false);

    expect(adminUiSessionCookieMaxAgeSeconds).toBe(8 * 60 * 60);
    expect(getCookieValue(
      `${adminUiAccessCookieName}=%E0%A4%A`,
      adminUiAccessCookieName,
    )).toBe("");
  });
});
