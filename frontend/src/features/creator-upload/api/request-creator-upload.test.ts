import { ApiError } from "@/shared/api";

import {
  completeCreatorUploadPackage,
  createCreatorUploadPackage,
  CreatorUploadApiError,
  uploadCreatorUploadTarget,
} from "./request-creator-upload";

function createJsonResponse(body: unknown, init: ResponseInit): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: {
      "Content-Type": "application/json",
    },
  });
}

function createVideoFile(name: string): File {
  return new File(["video"], name, { type: "video/mp4" });
}

describe("request creator upload", () => {
  it("creates an upload package with the selected files", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      createJsonResponse(
        {
          data: {
            expiresAt: "2026-04-08T12:15:00Z",
            packageToken: "cupkg_123",
            uploadTargets: {
              main: {
                fileName: "main.mp4",
                mimeType: "video/mp4",
                role: "main",
                upload: {
                  headers: {
                    "Content-Type": "video/mp4",
                  },
                  method: "PUT",
                  url: "https://raw-bucket.example.com/main",
                },
                uploadEntryId: "main-entry",
              },
              shorts: [
                {
                  fileName: "short.mp4",
                  mimeType: "video/mp4",
                  role: "short",
                  upload: {
                    headers: {
                      "Content-Type": "video/mp4",
                    },
                    method: "PUT",
                    url: "https://raw-bucket.example.com/short",
                  },
                  uploadEntryId: "short-entry",
                },
              ],
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_upload_packages_create_001",
          },
        },
        { status: 200 },
      ),
    );

    const result = await createCreatorUploadPackage({
      baseUrl: "https://api.example.com",
      fetcher,
      mainFile: createVideoFile("main.mp4"),
      shortFiles: [createVideoFile("short.mp4")],
    });

    expect(result.packageToken).toBe("cupkg_123");
    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/upload-packages"),
      expect.objectContaining({
        credentials: "include",
        method: "POST",
      }),
    );
    expect(fetcher).toHaveBeenCalledWith(
      expect.any(URL),
      expect.objectContaining({
        body: JSON.stringify({
          main: {
            fileName: "main.mp4",
            fileSizeBytes: 5,
            mimeType: "video/mp4",
          },
          shorts: [
            {
              fileName: "short.mp4",
              fileSizeBytes: 5,
              mimeType: "video/mp4",
            },
          ],
        }),
      }),
    );
  });

  it("maps backend contract failures to CreatorUploadApiError", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      createJsonResponse(
        {
          data: null,
          error: {
            code: "validation_error",
            message: "main is required",
          },
          meta: {
            page: null,
            requestId: "req_creator_upload_packages_validation_main_001",
          },
        },
        { status: 400 },
      ),
    );

    await expect(
      createCreatorUploadPackage({
        baseUrl: "https://api.example.com",
        fetcher,
        mainFile: createVideoFile("main.mp4"),
        shortFiles: [createVideoFile("short.mp4")],
      }),
    ).rejects.toEqual(
      new CreatorUploadApiError("validation_error", "main is required", {
        requestId: "req_creator_upload_packages_validation_main_001",
        status: 400,
      }),
    );
  });

  it("rejects malformed create-package success payloads that violate the contract", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      createJsonResponse(
        {
          data: {
            expiresAt: "2026-04-08T12:15:00Z",
            packageToken: "cupkg_123",
            uploadTargets: {
              main: {
                fileName: "main.mp4",
                mimeType: "video/mp4",
                role: "short",
                upload: {
                  headers: {
                    "Content-Type": "video/mp4",
                  },
                  method: "PUT",
                  url: "https://raw-bucket.example.com/main",
                },
                uploadEntryId: "main-entry",
              },
              shorts: [],
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_upload_packages_create_001",
          },
        },
        { status: 200 },
      ),
    );

    await expect(
      createCreatorUploadPackage({
        baseUrl: "https://api.example.com",
        fetcher,
        mainFile: createVideoFile("main.mp4"),
        shortFiles: [createVideoFile("short.mp4")],
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });

  it("completes the upload package with metadata and uploaded entry ids", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      createJsonResponse(
        {
          data: {
            main: {
              id: "main_001",
              mediaAsset: {
                id: "asset_main_001",
                mimeType: "video/mp4",
                processingState: "uploaded",
              },
              state: "draft",
            },
            shorts: [
              {
                canonicalMainId: "main_001",
                id: "short_001",
                mediaAsset: {
                  id: "asset_short_001",
                  mimeType: "video/mp4",
                  processingState: "uploaded",
                },
                state: "draft",
              },
            ],
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_upload_packages_complete_001",
          },
        },
        { status: 200 },
      ),
    );

    const result = await completeCreatorUploadPackage({
      baseUrl: "https://api.example.com",
      consentConfirmed: true,
      fetcher,
      mainUploadEntryId: "main-entry",
      ownershipConfirmed: true,
      packageToken: "cupkg_123",
      priceJpy: 1800,
      shorts: [
        {
          caption: null,
          uploadEntryId: "short-entry",
        },
      ],
    });

    expect(result.main.id).toBe("main_001");
    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/upload-packages/complete"),
      expect.objectContaining({
        body: JSON.stringify({
          main: {
            consentConfirmed: true,
            ownershipConfirmed: true,
            priceJpy: 1800,
            uploadEntryId: "main-entry",
          },
          packageToken: "cupkg_123",
          shorts: [{ caption: null, uploadEntryId: "short-entry" }],
        }),
      }),
    );
  });

  it("rejects malformed completion success payloads that violate the contract", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      createJsonResponse(
        {
          data: {
            main: {
              id: "main_001",
              mediaAsset: {
                id: "asset_main_001",
                mimeType: "video/mp4",
                processingState: "ready",
              },
              state: "published",
            },
            shorts: [],
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_upload_packages_complete_001",
          },
        },
        { status: 200 },
      ),
    );

    await expect(
      completeCreatorUploadPackage({
        baseUrl: "https://api.example.com",
        consentConfirmed: true,
        fetcher,
        mainUploadEntryId: "main-entry",
        ownershipConfirmed: true,
        packageToken: "cupkg_123",
        priceJpy: 1800,
        shorts: [
          {
            caption: "preview",
            uploadEntryId: "short-entry",
          },
        ],
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });

  it("uploads a file to the presigned target with backend-provided headers", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 200 }));
    const file = createVideoFile("main.mp4");

    await uploadCreatorUploadTarget({
      fetcher,
      file,
      target: {
        fileName: "main.mp4",
        mimeType: "video/mp4",
        role: "main",
        upload: {
          headers: {
            "Content-Type": "video/mp4",
          },
          method: "PUT",
          url: "https://raw-bucket.example.com/main",
        },
        uploadEntryId: "main-entry",
      },
    });

    expect(fetcher).toHaveBeenCalledWith("https://raw-bucket.example.com/main", {
      body: file,
      headers: {
        "Content-Type": "video/mp4",
      },
      method: "PUT",
    });
  });

  it("raises ApiError when direct upload returns a non-success status", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response("upload failed", { status: 403 }));

    await expect(
      uploadCreatorUploadTarget({
        fetcher,
        file: createVideoFile("main.mp4"),
        target: {
          fileName: "main.mp4",
          mimeType: "video/mp4",
          role: "main",
          upload: {
            headers: {
              "Content-Type": "video/mp4",
            },
            method: "PUT",
            url: "https://raw-bucket.example.com/main",
          },
          uploadEntryId: "main-entry",
        },
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});
