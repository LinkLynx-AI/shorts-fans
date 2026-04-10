import { ApiError } from "@/shared/api";

import {
  completeCreatorRegistrationAvatarUpload,
  createCreatorRegistrationAvatarUpload,
  uploadCreatorRegistrationAvatarTarget,
} from "./avatar-upload";

function createAvatarFile(name: string, type = "image/png", body = "avatar"): File {
  return new File([body], name, { type });
}

describe("creator registration avatar upload api", () => {
  it("posts avatar file metadata to create the upload target", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            avatarUploadToken: "vcupl_token",
            expiresAt: "2026-04-10T12:15:00Z",
            uploadTarget: {
              fileName: "avatar.png",
              mimeType: "image/png",
              upload: {
                headers: {
                  "Content-Type": "image/png",
                },
                method: "PUT",
                url: "https://raw-bucket.example.com/avatar",
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_avatar_create",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      createCreatorRegistrationAvatarUpload(createAvatarFile("avatar.png"), {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      avatarUploadToken: "vcupl_token",
      expiresAt: "2026-04-10T12:15:00Z",
      uploadTarget: {
        fileName: "avatar.png",
        mimeType: "image/png",
        upload: {
          headers: {
            "Content-Type": "image/png",
          },
          method: "PUT",
          url: "https://raw-bucket.example.com/avatar",
        },
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/viewer/creator-registration/avatar-uploads"),
      expect.objectContaining({
        body: JSON.stringify({
          fileName: "avatar.png",
          fileSizeBytes: 6,
          mimeType: "image/png",
        }),
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("posts avatarUploadToken to complete the upload", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            avatar: {
              durationSeconds: null,
              id: "asset_creator_registration_avatar_fixed",
              kind: "image",
              posterUrl: null,
              url: "https://cdn.example.com/creator-avatar/avatar.png",
            },
            avatarUploadToken: "vcupl_token",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_avatar_complete",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      completeCreatorRegistrationAvatarUpload("vcupl_token", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      avatar: {
        durationSeconds: null,
        id: "asset_creator_registration_avatar_fixed",
        kind: "image",
        posterUrl: null,
        url: "https://cdn.example.com/creator-avatar/avatar.png",
      },
      avatarUploadToken: "vcupl_token",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/viewer/creator-registration/avatar-uploads/complete"),
      expect.objectContaining({
        body: JSON.stringify({
          avatarUploadToken: "vcupl_token",
        }),
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("raises ApiError when the direct upload target returns a non-success status", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response("upload failed", { status: 403 }));

    await expect(
      uploadCreatorRegistrationAvatarTarget({
        fetcher,
        file: createAvatarFile("avatar.png"),
        target: {
          fileName: "avatar.png",
          mimeType: "image/png",
          upload: {
            headers: {
              "Content-Type": "image/png",
            },
            method: "PUT",
            url: "https://raw-bucket.example.com/avatar",
          },
        },
      }),
    ).rejects.toBeInstanceOf(ApiError);

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://raw-bucket.example.com/avatar"),
      expect.objectContaining({
        body: expect.any(File),
        method: "PUT",
      }),
    );
  });
});
