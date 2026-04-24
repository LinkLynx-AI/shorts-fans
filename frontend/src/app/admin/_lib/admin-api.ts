const adminAPITokenHeader = "X-Shorts-Fans-Admin-Token";

export type AdminAPIFetcher = typeof fetch;

export function getAdminAPIToken(): string {
  return process.env.ADMIN_API_TOKEN?.trim() ?? "";
}

export function createAdminAPIFetcher(fetcher: AdminAPIFetcher = fetch): AdminAPIFetcher {
  const token = getAdminAPIToken();

  return (input, init) => {
    const headers = new Headers(init?.headers);
    if (token) {
      headers.set(adminAPITokenHeader, token);
    }

    return fetcher(input, {
      ...init,
      headers,
    });
  };
}
