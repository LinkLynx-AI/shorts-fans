import { ApiError } from "@/shared/api";

import {
  CreatorWorkspaceMainPriceApiError,
  updateCreatorWorkspaceMainPrice,
} from "./update-creator-workspace-main-price";

describe("updateCreatorWorkspaceMainPrice", () => {
  it("puts the next price with credentials included", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({
      data: {
        main: {
          id: "main_quiet_rooftop",
          priceJpy: 2400,
        },
      },
      error: null,
      meta: {
        page: null,
        requestId: "req_creator_workspace_main_price_update_001",
      },
    }), { status: 200 }));

    await expect(updateCreatorWorkspaceMainPrice({
      baseUrl: "https://api.example.com",
      fetcher,
      mainId: "main_quiet_rooftop",
      priceJpy: 2400,
    })).resolves.toEqual({
      main: {
        id: "main_quiet_rooftop",
        priceJpy: 2400,
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/mains/main_quiet_rooftop/price"),
      expect.objectContaining({
        body: JSON.stringify({
          priceJpy: 2400,
        }),
        credentials: "include",
        method: "PUT",
      }),
    );
  });

  it("surfaces contract errors as CreatorWorkspaceMainPriceApiError", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({
      data: null,
      error: {
        code: "validation_error",
        message: "priceJpy must be a positive integer",
      },
      meta: {
        page: null,
        requestId: "req_creator_workspace_main_price_update_002",
      },
    }), { status: 422 }));

    await expect(updateCreatorWorkspaceMainPrice({
      baseUrl: "https://api.example.com",
      fetcher,
      mainId: "main_quiet_rooftop",
      priceJpy: 0,
    })).rejects.toEqual(new CreatorWorkspaceMainPriceApiError(
      "validation_error",
      "priceJpy must be a positive integer",
      {
        requestId: "req_creator_workspace_main_price_update_002",
        status: 422,
      },
    ));
  });

  it("surfaces malformed success payloads as ApiError", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({
      data: {
        main: {
          id: "main_quiet_rooftop",
        },
      },
      error: null,
      meta: {
        page: null,
        requestId: "req_creator_workspace_main_price_update_003",
      },
    }), { status: 200 }));

    await expect(updateCreatorWorkspaceMainPrice({
      baseUrl: "https://api.example.com",
      fetcher,
      mainId: "main_quiet_rooftop",
      priceJpy: 2400,
    })).rejects.toBeInstanceOf(ApiError);
  });
});
