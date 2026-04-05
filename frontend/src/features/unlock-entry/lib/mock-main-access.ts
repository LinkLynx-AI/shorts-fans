const MOCK_MAIN_ACCESS_STORAGE_KEY = "shorts-fans:mock-main-access";

type MockMainAccessMap = Record<string, string>;

function readMockMainAccessMap(): MockMainAccessMap {
  if (typeof window === "undefined") {
    return {};
  }

  const stored = window.sessionStorage.getItem(MOCK_MAIN_ACCESS_STORAGE_KEY);

  if (!stored) {
    return {};
  }

  try {
    const parsed = JSON.parse(stored);
    return typeof parsed === "object" && parsed !== null ? (parsed as MockMainAccessMap) : {};
  } catch {
    return {};
  }
}

function writeMockMainAccessMap(value: MockMainAccessMap): void {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.setItem(MOCK_MAIN_ACCESS_STORAGE_KEY, JSON.stringify(value));
}

function createGrantToken(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  return `grant_${Math.random().toString(36).slice(2, 12)}`;
}

/**
 * main 再生用の mock access grant を発行して保存する。
 */
export function issueMockMainAccessGrant(mainId: string): string {
  const nextToken = createGrantToken();
  const current = readMockMainAccessMap();

  writeMockMainAccessMap({
    ...current,
    [mainId]: nextToken,
  });

  return nextToken;
}

/**
 * main 再生用の grant が有効か判定する。
 */
export function hasMockMainAccessGrant(mainId: string, grantToken: string): boolean {
  if (!grantToken) {
    return false;
  }

  return readMockMainAccessMap()[mainId] === grantToken;
}
