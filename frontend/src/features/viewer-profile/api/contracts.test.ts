import { viewerProfileResponseSchema } from "@/features/viewer-profile/api/contracts";

describe("viewerProfileResponseSchema", () => {
  it("accepts handles that match the shared profile contract", () => {
    const result = viewerProfileResponseSchema.safeParse({
      data: {
        profile: {
          avatar: null,
          displayName: "Mina Rei",
          handle: "@mina.rei_01",
        },
      },
      error: null,
      meta: {
        page: null,
        requestId: "req_viewer_profile_001",
      },
    });

    expect(result.success).toBe(true);
  });

  it("rejects handles that violate the shared profile contract", () => {
    for (const handle of ["@", "@Mina", "@mina-rei"]) {
      const result = viewerProfileResponseSchema.safeParse({
        data: {
          profile: {
            avatar: null,
            displayName: "Mina Rei",
            handle,
          },
        },
        error: null,
        meta: {
          page: null,
          requestId: "req_viewer_profile_001",
        },
      });

      expect(result.success).toBe(false);
    }
  });
});
