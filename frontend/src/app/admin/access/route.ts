import { NextResponse, type NextRequest } from "next/server";

import {
  adminUiAccessHeaderName,
  adminUiAccessCookieName,
  adminUiSessionCookieMaxAgeSeconds,
  createAdminUiSessionCookieValue,
  isAdminUiAccessTokenMatch,
  isAdminUiEnabled,
} from "../_lib/admin-ui-access";

const noStoreHeaders = {
  "Cache-Control": "no-store",
  "Referrer-Policy": "no-referrer",
};

function notFoundResponse(): NextResponse {
  return new NextResponse("not found", {
    headers: noStoreHeaders,
    status: 404,
  });
}

function resolveSafeAdminRedirectPath(value: string | null): string {
  if (!value || !value.startsWith("/admin") || value.startsWith("//") || value.includes("://")) {
    return "/admin";
  }

  return value;
}

export function GET(): NextResponse {
  return notFoundResponse();
}

export function POST(request: NextRequest): NextResponse {
  const requestURL = new URL(request.url);
  if (!isAdminUiEnabled()) {
    return notFoundResponse();
  }

  const token = request.headers.get(adminUiAccessHeaderName)?.trim() ?? "";
  if (!isAdminUiAccessTokenMatch(token)) {
    return notFoundResponse();
  }

  const redirectURL = new URL(resolveSafeAdminRedirectPath(requestURL.searchParams.get("next")), request.url);
  const response = NextResponse.redirect(redirectURL, {
    headers: noStoreHeaders,
    status: 303,
  });
  response.cookies.set(adminUiAccessCookieName, createAdminUiSessionCookieValue(), {
    httpOnly: true,
    maxAge: adminUiSessionCookieMaxAgeSeconds,
    path: "/admin",
    sameSite: "strict",
  });

  return response;
}
