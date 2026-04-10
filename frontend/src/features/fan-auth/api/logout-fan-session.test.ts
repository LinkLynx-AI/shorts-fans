import { logoutFanSession } from "@/features/fan-auth";

describe("logoutFanSession", () => {
  it("deletes the fan session with credentials included", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      logoutFanSession({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/fan/auth/session"),
      expect.objectContaining({
        credentials: "include",
        method: "DELETE",
      }),
    );
  });

  it("surfaces transport failures as shared API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockRejectedValue(new Error("network down"));

    await expect(
      logoutFanSession({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).rejects.toEqual(
      expect.objectContaining({
        code: "network",
        message: "API request failed before a response was received.",
        name: "ApiError",
      }),
    );
  });
});
