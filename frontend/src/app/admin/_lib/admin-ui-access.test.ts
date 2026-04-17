import { isAdminUiEnabled } from "./admin-ui-access";

describe("admin ui access", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("returns true only when the dedicated admin env flag is enabled", () => {
    vi.stubEnv("ADMIN_UI_ENABLED", "1");
    expect(isAdminUiEnabled()).toBe(true);

    vi.stubEnv("ADMIN_UI_ENABLED", "0");
    expect(isAdminUiEnabled()).toBe(false);

    vi.unstubAllEnvs();
    expect(isAdminUiEnabled()).toBe(false);
  });
});
