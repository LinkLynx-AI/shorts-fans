import { createHmac, randomBytes, timingSafeEqual } from "crypto";
import { headers } from "next/headers";
import { notFound } from "next/navigation";

const adminUiEnabledValue = "1";
const developmentNodeEnv = "development";
const adminUiSessionCookieVersion = "v1";
const adminUiSessionSignatureContext = "shorts-fans-admin-ui-session";

export const adminUiAccessCookieName = "shorts_fans_admin_ui_token";
export const adminUiAccessHeaderName = "x-shorts-fans-admin-ui-token";
export const adminUiSessionCookieMaxAgeSeconds = 8 * 60 * 60;

export function isAdminUiEnabled(): boolean {
  return (
    process.env.NODE_ENV === developmentNodeEnv &&
    process.env.ADMIN_UI_ENABLED === adminUiEnabledValue
  );
}

export function getAdminUiAccessToken(): string {
  return process.env.ADMIN_UI_ACCESS_TOKEN?.trim() ?? "";
}

export function getCookieValue(cookieHeader: string | null | undefined, cookieName: string): string {
  if (!cookieHeader) {
    return "";
  }

  for (const cookiePart of cookieHeader.split(";")) {
    const [rawName, ...rawValue] = cookiePart.split("=");
    if (rawName?.trim() !== cookieName) {
      continue;
    }

    try {
      return decodeURIComponent(rawValue.join("=").trim());
    } catch {
      return "";
    }
  }

  return "";
}

function safeTimingEqual(leftValue: string, rightValue: string): boolean {
  const left = Buffer.from(leftValue);
  const right = Buffer.from(rightValue);
  return left.length === right.length && timingSafeEqual(left, right);
}

export function isAdminUiAccessTokenMatch(actualToken: string | null | undefined): boolean {
  const expectedToken = getAdminUiAccessToken();
  const normalizedActual = actualToken?.trim() ?? "";
  if (!expectedToken || !normalizedActual) {
    return false;
  }

  return safeTimingEqual(expectedToken, normalizedActual);
}

function signAdminUiSession({ expiresAt, nonce }: { expiresAt: string; nonce: string }): string {
  const accessToken = getAdminUiAccessToken();
  if (!accessToken) {
    return "";
  }

  return createHmac("sha256", accessToken)
    .update(adminUiSessionSignatureContext)
    .update(":")
    .update(expiresAt)
    .update(":")
    .update(nonce)
    .digest("base64url");
}

export function createAdminUiSessionCookieValue(): string {
  const expiresAt = String(Math.floor(Date.now() / 1000) + adminUiSessionCookieMaxAgeSeconds);
  const nonce = randomBytes(24).toString("base64url");
  const signature = signAdminUiSession({ expiresAt, nonce });
  return `${adminUiSessionCookieVersion}.${expiresAt}.${nonce}.${signature}`;
}

export function isAdminUiSessionCookieMatch(cookieValue: string | null | undefined): boolean {
  const normalizedValue = cookieValue?.trim() ?? "";
  const [version, expiresAtValue, nonce, signature, extra] = normalizedValue.split(".");
  if (
    extra !== undefined ||
    version !== adminUiSessionCookieVersion ||
    !expiresAtValue ||
    !nonce ||
    !signature
  ) {
    return false;
  }

  const expiresAt = Number(expiresAtValue);
  if (!Number.isSafeInteger(expiresAt) || expiresAt <= Math.floor(Date.now() / 1000)) {
    return false;
  }

  const expectedSignature = signAdminUiSession({ expiresAt: expiresAtValue, nonce });
  if (!expectedSignature) {
    return false;
  }

  return safeTimingEqual(expectedSignature, signature);
}

export function isAdminUiRequestAllowed({
  uiSessionCookie,
}: {
  uiSessionCookie: string | null | undefined;
}): boolean {
  return isAdminUiEnabled() && isAdminUiSessionCookieMatch(uiSessionCookie);
}

export async function assertAdminUiAccess(): Promise<void> {
  const requestHeaders = await headers();
  if (
    !isAdminUiRequestAllowed({
      uiSessionCookie: getCookieValue(requestHeaders.get("cookie"), adminUiAccessCookieName),
    })
  ) {
    notFound();
  }
}
