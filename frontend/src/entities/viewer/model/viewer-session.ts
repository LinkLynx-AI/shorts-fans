import { viewerSessionCookieName } from "./current-viewer";

/**
 * Cookie header から viewer session token を読み取る。
 */
export function readViewerSessionToken(cookieHeader: string | null | undefined): string | null {
  if (!cookieHeader) {
    return null;
  }

  for (const cookiePart of cookieHeader.split(";")) {
    const trimmedCookiePart = cookiePart.trim();

    if (!trimmedCookiePart.startsWith(`${viewerSessionCookieName}=`)) {
      continue;
    }

    const rawValue = trimmedCookiePart.slice(viewerSessionCookieName.length + 1).trim();

    if (rawValue.length > 0) {
      return rawValue;
    }
  }

  return null;
}

/**
 * Cookie header に viewer session cookie が含まれるかを判定する。
 */
export function hasViewerSession(cookieHeader: string | null | undefined): boolean {
  return readViewerSessionToken(cookieHeader) !== null;
}
