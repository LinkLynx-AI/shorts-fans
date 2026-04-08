import { expect, test, type Page } from "@playwright/test";

const viewerSessionCookieBase = {
  domain: "127.0.0.1",
  name: "shorts_fans_session",
  path: "/",
} as const;

function createViewerSessionToken(): string {
  return `e2e-viewer-session-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

async function addViewerSession(page: Page, value = createViewerSessionToken()) {
  await page.context().addCookies([
    {
      ...viewerSessionCookieBase,
      value,
    },
  ]);
}

function buildViewerSessionCookieHeader(value = createViewerSessionToken()) {
  return `${viewerSessionCookieBase.name}=${value}`;
}

async function expectMainShortcutNavigation(page: Page, options: {
  buttonName: RegExp;
  destinationPattern: RegExp;
  path: string;
}) {
  await page.goto(options.path);

  const [response] = await Promise.all([
    page.waitForResponse((candidate) =>
      candidate.request().method() === "POST" &&
      candidate.url().includes("/api/mock-main-access"),
    ),
    page.getByRole("button", { name: options.buttonName }).click(),
  ]);

  expect(response.status()).toBe(200);
  await expect(page).toHaveURL(options.destinationPattern);
}

test("fan shell routes render and unlock flow works", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole("link", { name: "おすすめ" })).toHaveAttribute("aria-current", "page");
  await expect(page.getByRole("button", { name: /Unlock/i })).toBeVisible();
  await expect(page.getByText("Mina Rei")).toBeVisible();

  await page.getByRole("button", { name: /Unlock/i }).click();
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole("heading", { name: "続けるにはログインが必要です" })).toBeVisible();

  await page.goto("/?tab=following");
  await expect(page).toHaveURL(/\/login$/);

  await page.goto("/fan");
  await expect(page).toHaveURL(/\/login$/);

  await addViewerSession(page);
  await page.goto("/");
  await page.getByRole("button", { name: /Unlock/i }).click();
  const paywall = page.getByRole("dialog", { name: "quiet rooftop preview の続きを見る" });
  await expect(paywall).toBeVisible();
  await expect(paywall.getByRole("button", { name: "Unlock ¥1,800 | 8分" })).toBeDisabled();
  await page.getByLabel("18歳以上であり、年齢確認に同意する").check();
  await page.getByLabel("利用規約とポリシーに同意し、確認面なしで main 再生へ進む").check();
  await paywall.getByRole("button", { name: "Unlock ¥1,800 | 8分" }).click();
  await expect(page).toHaveURL(/\/mains\/main_mina_quiet_rooftop\?fromShortId=rooftop&grant=/);
  await expect(page.getByText("Playing main")).toBeVisible();
  await page.getByRole("button", { name: "Back" }).click();
  await expect(page).toHaveURL(/\/$/);

  await page.getByRole("link", { name: "検索" }).click();
  await expect(page).toHaveURL(/\/search$/);
  await page.getByRole("searchbox").fill("mina");
  await expect(page.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
    "href",
    "/creators/creator_mina_rei?from=search&q=mina",
  );
  await page.getByRole("link", { name: /Mina Rei/i }).click();
  await expect(page).toHaveURL(/\/creators\/creator_mina_rei\?from=search&q=mina$/);
  await expect(page.getByRole("heading", { name: /Mina Rei creator profile/i })).toHaveCount(1);
  await expect(page.getByRole("button", { name: "Following" })).toBeVisible();
  await expect(page.getByText("Unlock")).toHaveCount(0);
  await page.getByRole("link", { name: /Back/i }).click();
  await expect(page).toHaveURL(/\/search\?q=mina$/);
  await expect(page.getByRole("searchbox")).toHaveValue("mina");

  await page.goto("/");
  await page.getByRole("link", { name: /Mina Rei/i }).click();
  await expect(page).toHaveURL(/\/creators\/creator_mina_rei\?from=feed&tab=recommended$/);
  await page.getByRole("link", { name: /Mina Rei preview 0:16/i }).click();
  await expect(page).toHaveURL(/\/shorts\/rooftop\?creatorId=creator_mina_rei&from=creator&profileFrom=feed&profileTab=recommended$/);
  await page.getByRole("link", { name: /Back/i }).click();
  await expect(page).toHaveURL(/\/creators\/creator_mina_rei\?from=feed&tab=recommended$/);

  await page.getByRole("link", { name: "マイ" }).click();
  await expect(page).toHaveURL(/\/fan$/);
  await expect(page.getByRole("heading", { name: "My archive" })).toBeVisible();

  await page.getByRole("link", { name: "Following" }).click();
  await expect(page).toHaveURL(/\/fan\/following$/);
  await expect(page.getByRole("heading", { name: "following" })).toBeVisible();

  await page.goto("/shorts/rooftop");
  await expect(page.getByRole("link", { name: /Back/i })).toBeVisible();
  await expect(page.getByText("quiet rooftop preview.")).toBeVisible();

  await page.goto("/mains/main_mina_quiet_rooftop");
  await expect(page.getByRole("heading", { name: "指定された surface はまだ用意されていません。" })).toBeVisible();

  await page.goto("/mains/main_mina_quiet_rooftop?fromShortId=rooftop&grant=invalid");
  await expect(page.getByRole("heading", { name: "この main はまだ unlock されていません。" })).toBeVisible();

  await page.goto("/creators/creator_mina_rei");
  await expect(page.getByRole("heading", { name: /Mina Rei creator profile/i })).toHaveCount(1);

  await page.goto("/creators/creator_sora_vale");
  await expect(page.getByText("まだ公開中の short はありません。")).toBeVisible();
  await expect(page.getByRole("button", { name: "Follow" })).toBeEnabled();

  await page.goto("/creators/creator_mina_rei?from=twitter&tab=other");
  await expect(page).toHaveURL(/\/creators\/creator_mina_rei\?from=twitter&tab=other$/);
  await expect(page.getByRole("heading", { name: /Mina Rei creator profile/i })).toHaveCount(1);
  await expect(page.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");

  await page.goto("/shorts/rooftop?from=share&profileFrom=other");
  await expect(page).toHaveURL(/\/shorts\/rooftop\?from=share&profileFrom=other$/);
  await expect(page.getByText("quiet rooftop preview.")).toBeVisible();
  await expect(page.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
});

test("unauthenticated viewers can sign in from the shared auth modal opened by the profile button", async ({ page }) => {
  await page.goto("/");

  await page.getByRole("link", { name: "マイ" }).click();
  await expect(page.getByRole("dialog", { name: "続けるにはログインが必要です" })).toBeVisible();

  await page.getByRole("textbox", { name: "Email" }).fill("fan@example.com");
  await page.getByRole("button", { name: "サインインを続ける" }).click();

  await expect(page).toHaveURL(/\/fan$/);
  await expect(page.getByRole("heading", { name: "My archive" })).toBeVisible();

  await page.getByRole("link", { name: "フィード" }).click();
  await expect(page).toHaveURL(/\/$/);
  await page.getByRole("link", { name: "マイ" }).click();
  await expect(page).toHaveURL(/\/fan$/);
  await expect(page.getByRole("dialog", { name: "続けるにはログインが必要です" })).toHaveCount(0);
});

test("unauthenticated viewers can open the shared auth modal from creator profile follow without auto-following", async ({ page }) => {
  await page.goto("/creators/creator_aoi_n");

  await page.getByRole("button", { name: "Follow" }).click();
  await expect(page.getByRole("dialog", { name: "続けるにはログインが必要です" })).toBeVisible();

  await page.getByRole("textbox", { name: "Email" }).fill("fan@example.com");
  await page.getByRole("button", { name: "サインインを続ける" }).click();

  await expect(page).toHaveURL(/\/creators\/creator_aoi_n$/);
  await expect(page.getByRole("button", { name: "Follow" })).toBeVisible();
});

test("authenticated viewers can follow and unfollow from creator profile", async ({ page }) => {
  await addViewerSession(page);
  await page.goto("/creators/creator_aoi_n");

  await expect(page.getByRole("button", { name: "Follow" })).toBeVisible();

  const [followResponse] = await Promise.all([
    page.waitForResponse((candidate) =>
      candidate.request().method() === "PUT" &&
      candidate.url().includes("/api/fan/creators/creator_aoi_n/follow"),
    ),
    page.getByRole("button", { name: "Follow" }).click(),
  ]);

  expect(followResponse.status()).toBe(200);
  await expect(page.getByRole("button", { name: "Following" })).toBeVisible();

  const [unfollowResponse] = await Promise.all([
    page.waitForResponse((candidate) =>
      candidate.request().method() === "DELETE" &&
      candidate.url().includes("/api/fan/creators/creator_aoi_n/follow"),
    ),
    page.getByRole("button", { name: "Following" }).click(),
  ]);

  expect(unfollowResponse.status()).toBe(200);
  await expect(page.getByRole("button", { name: "Follow" })).toBeVisible();
});

test("unauthenticated viewers can sign up from the shared auth modal and enter the fan hub", async ({ page }) => {
  await page.goto("/");

  await page.getByRole("link", { name: "マイ" }).click();
  await page.getByRole("button", { name: "サインアップへ" }).click();
  await page.getByRole("textbox", { name: "Email" }).fill("newfan@example.com");
  await page.getByRole("button", { name: "新規登録を続ける" }).click();

  await expect(page).toHaveURL(/\/fan$/);
  await expect(page.getByRole("heading", { name: "My archive" })).toBeVisible();
});

test("invalid grant response does not leak protected playback data", async ({ request }) => {
  const response = await request.get("/mains/main_mina_quiet_rooftop?fromShortId=rooftop&grant=invalid", {
    headers: {
      Cookie: buildViewerSessionCookieHeader(),
    },
  });
  const body = await response.text();

  expect(body).toContain("この main はまだ unlock されていません。");
  expect(body).not.toContain("quiet rooftop main");
  expect(body).not.toContain("cdn.example.com/mains/");
});

test("main access route returns auth_required for unauthenticated viewers", async ({ request }) => {
  const response = await request.post("/api/mock-main-access", {
    data: {
      acceptedAge: true,
      acceptedTerms: true,
      entryToken: "invalid",
      fromShortId: "rooftop",
      mainId: "main_mina_quiet_rooftop",
    },
  });

  expect(response.status()).toBe(401);
  await expect(response.json()).resolves.toEqual({
    data: null,
    error: {
      code: "auth_required",
      message: "main playback requires authentication",
    },
    meta: {
      page: null,
      requestId: "req_mock_main_access_auth_required_001",
    },
  });
});

test("stale session cookies are treated as unauthenticated on protected fan surfaces", async ({ page }) => {
  await page.context().addCookies([
    {
      ...viewerSessionCookieBase,
      value: "stale-e2e-session",
    },
  ]);

  await page.goto("/fan");
  await expect(page).toHaveURL(/\/login$/);
  await page.goto("/shorts/softlight");
  await page.getByRole("button", { name: /Continue main/i }).click();
  await expect(page).toHaveURL(/\/login$/);
});

test("main access route rejects direct setup bypass requests after authentication", async ({ request }) => {
  const response = await request.post("/api/mock-main-access", {
    data: {
      acceptedAge: true,
      acceptedTerms: true,
      entryToken: "invalid",
      fromShortId: "rooftop",
      mainId: "main_mina_quiet_rooftop",
    },
    headers: {
      Cookie: buildViewerSessionCookieHeader(),
    },
  });

  expect(response.status()).toBe(403);
  await expect(response.json()).resolves.toEqual({
    fallbackHref: "/shorts/rooftop",
  });
});

test("authenticated viewers can continue purchased main playback from short detail", async ({ page }) => {
  await addViewerSession(page);

  await expectMainShortcutNavigation(page, {
    buttonName: /Continue main/i,
    destinationPattern: /\/mains\/main_aoi_blue_balcony\?fromShortId=softlight&grant=/,
    path: "/shorts/softlight",
  });
});

test("authenticated viewers can open Aoi creator profile from main playback", async ({ page }) => {
  await addViewerSession(page);
  await page.goto("/shorts/softlight");
  await page.getByRole("button", { name: /Continue main/i }).click();
  await expect(page).toHaveURL(/\/mains\/main_aoi_blue_balcony\?fromShortId=softlight&grant=/);

  await page.getByRole("link", { name: /Aoi N/i }).click();

  await expect(page).toHaveURL(/\/creators\/creator_aoi_n$/);
  await expect(page.getByRole("heading", { name: /Aoi N creator profile/i })).toHaveCount(1);
  await expect(page.getByRole("button", { name: "Follow" })).toBeVisible();
});

test("authenticated viewers can unlock a purchased-required main from short detail", async ({ page }) => {
  await addViewerSession(page);

  await expectMainShortcutNavigation(page, {
    buttonName: /Unlock/i,
    destinationPattern: /\/mains\/main_sora_after_rain\?fromShortId=afterrain&grant=/,
    path: "/shorts/afterrain",
  });
});

test("authenticated viewers can open owner preview main playback from short detail", async ({ page }) => {
  await addViewerSession(page);

  await expectMainShortcutNavigation(page, {
    buttonName: /Owner preview/i,
    destinationPattern: /\/mains\/main_aoi_blue_balcony\?fromShortId=balcony&grant=/,
    path: "/shorts/balcony",
  });
});

test("signed grants cannot be replayed against a different main context", async ({ page, request }) => {
  await addViewerSession(page);
  await page.goto("/");
  await page.getByRole("button", { name: /Unlock/i }).click();
  await page.getByLabel("18歳以上であり、年齢確認に同意する").check();
  await page.getByLabel("利用規約とポリシーに同意し、確認面なしで main 再生へ進む").check();
  await page.getByRole("button", { name: "Unlock ¥1,800 | 8分" }).click();
  await expect(page).toHaveURL(/\/mains\/main_mina_quiet_rooftop\?fromShortId=rooftop&grant=/);

  const mainUrl = new URL(page.url());
  const grant = mainUrl.searchParams.get("grant");

  expect(grant).toBeTruthy();

  const response = await request.get(`/mains/main_aoi_blue_balcony?fromShortId=softlight&grant=${grant}`, {
    headers: {
      Cookie: buildViewerSessionCookieHeader(),
    },
  });
  const body = await response.text();

  expect(body).toContain("この main はまだ unlock されていません。");
  expect(body).not.toContain("quiet rooftop main");
  expect(body).not.toContain("cdn.example.com/mains/");
});

test("undefined routes fall back to the shared not-found page", async ({ page }) => {
  await page.goto("/missing-route");
  await expect(page.getByRole("heading", { name: "指定された surface はまだ用意されていません。" })).toBeVisible();
  await page.getByRole("link", { name: "feed に戻る" }).click();
  await expect(page).toHaveURL(/\/$/);
});
