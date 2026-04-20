import { z } from "zod";

import { requestJson } from "@/shared/api";

const creatorWorkspaceReviewBlockerCodeSchema = z.enum([
  "linked_short_missing",
  "main_asset_not_ready",
  "short_asset_not_ready",
  "main_price_missing",
  "ownership_missing",
  "consent_missing",
]);

const creatorWorkspaceReviewPackageReadinessSchema = z.enum([
  "blocked",
  "conflict",
  "none",
  "ready",
]);

const creatorWorkspaceReviewPackageStatusSchema = z.enum([
  "approved",
  "changes_requested",
  "draft",
  "pending_review",
  "rejected",
]);

const creatorWorkspaceReviewSubmitActionSchema = z.enum([
  "none",
  "resubmit",
  "submit",
]);

const creatorWorkspaceReviewTargetKindSchema = z.enum([
  "main",
  "short",
]);

const creatorWorkspaceReviewObjectStateSchema = z.string().min(1);

const creatorWorkspaceReviewPackageSummarySchema = z.object({
  blockers: z.array(creatorWorkspaceReviewBlockerCodeSchema),
  canonicalMainId: z.string().min(1),
  linkedShortCount: z.number().int().nonnegative(),
  readiness: creatorWorkspaceReviewPackageReadinessSchema,
  reviewStatus: creatorWorkspaceReviewPackageStatusSchema,
  submitAction: creatorWorkspaceReviewSubmitActionSchema,
});

const creatorWorkspaceReviewMainItemSchema = z.object({
  id: z.string().min(1),
  state: creatorWorkspaceReviewObjectStateSchema,
});

const creatorWorkspaceReviewShortItemSchema = z.object({
  canonicalMainId: z.string().min(1),
  id: z.string().min(1),
  state: creatorWorkspaceReviewObjectStateSchema,
});

const creatorWorkspaceReviewTargetStateSchema = z.object({
  reasonCode: z.string().min(1).nullable(),
  state: creatorWorkspaceReviewObjectStateSchema,
});

const creatorWorkspaceReviewTargetSchema = z.object({
  canonicalMainId: z.string().min(1),
  id: z.string().min(1),
  kind: creatorWorkspaceReviewTargetKindSchema,
});

const creatorWorkspaceReviewSurfaceResponseSchema = z.object({
  data: z.object({
    reviewSurface: z.object({
      mains: z.array(creatorWorkspaceReviewMainItemSchema),
      packages: z.array(creatorWorkspaceReviewPackageSummarySchema),
      shorts: z.array(creatorWorkspaceReviewShortItemSchema),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

const creatorWorkspaceReviewDetailResponseSchema = z.object({
  data: z.object({
    reviewSurface: z.object({
      package: creatorWorkspaceReviewPackageSummarySchema,
      review: creatorWorkspaceReviewTargetStateSchema,
      target: creatorWorkspaceReviewTargetSchema,
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspaceReviewPackageSummary = z.output<typeof creatorWorkspaceReviewPackageSummarySchema>;
export type CreatorWorkspaceReviewMainItem = z.output<typeof creatorWorkspaceReviewMainItemSchema>;
export type CreatorWorkspaceReviewShortItem = z.output<typeof creatorWorkspaceReviewShortItemSchema>;
export type CreatorWorkspaceReviewTargetState = z.output<typeof creatorWorkspaceReviewTargetStateSchema>;
export type CreatorWorkspaceReviewTarget = z.output<typeof creatorWorkspaceReviewTargetSchema>;

export type CreatorWorkspaceReviewSurface = {
  mains: readonly CreatorWorkspaceReviewMainItem[];
  packages: readonly CreatorWorkspaceReviewPackageSummary[];
  requestId: string;
  shorts: readonly CreatorWorkspaceReviewShortItem[];
};

export type CreatorWorkspaceItemReviewSurface = {
  package: CreatorWorkspaceReviewPackageSummary;
  requestId: string;
  review: CreatorWorkspaceReviewTargetState;
  target: CreatorWorkspaceReviewTarget;
};

type GetCreatorWorkspaceReviewSurfaceOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  signal?: AbortSignal;
};

function buildCreatorWorkspaceReviewSurfacePath(
  pathname:
    | "/api/creator/workspace/review-surface"
    | "/api/creator/workspace/mains"
    | "/api/creator/workspace/shorts",
  id?: string,
): `/${string}` {
  if (!id) {
    return pathname;
  }

  return `${pathname}/${encodeURIComponent(id.trim())}/review-surface`;
}

/**
 * creator workspace dashboard の審査 status surface を取得する。
 */
export async function getCreatorWorkspaceReviewSurface({
  baseUrl,
  credentials = "include",
  fetcher,
  signal,
}: GetCreatorWorkspaceReviewSurfaceOptions = {}): Promise<CreatorWorkspaceReviewSurface> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspaceReviewSurfacePath("/api/creator/workspace/review-surface"),
    schema: creatorWorkspaceReviewSurfaceResponseSchema,
  });

  return {
    mains: response.data.reviewSurface.mains,
    packages: response.data.reviewSurface.packages,
    requestId: response.meta.requestId,
    shorts: response.data.reviewSurface.shorts,
  };
}

/**
 * creator workspace main detail の審査 status surface を取得する。
 */
export async function getCreatorWorkspaceMainReviewSurface(
  mainId: string,
  {
    baseUrl,
    credentials = "include",
    fetcher,
    signal,
  }: GetCreatorWorkspaceReviewSurfaceOptions = {},
): Promise<CreatorWorkspaceItemReviewSurface> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspaceReviewSurfacePath("/api/creator/workspace/mains", mainId),
    schema: creatorWorkspaceReviewDetailResponseSchema,
  });

  return {
    package: response.data.reviewSurface.package,
    requestId: response.meta.requestId,
    review: response.data.reviewSurface.review,
    target: response.data.reviewSurface.target,
  };
}

/**
 * creator workspace short detail の審査 status surface を取得する。
 */
export async function getCreatorWorkspaceShortReviewSurface(
  shortId: string,
  {
    baseUrl,
    credentials = "include",
    fetcher,
    signal,
  }: GetCreatorWorkspaceReviewSurfaceOptions = {},
): Promise<CreatorWorkspaceItemReviewSurface> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspaceReviewSurfacePath("/api/creator/workspace/shorts", shortId),
    schema: creatorWorkspaceReviewDetailResponseSchema,
  });

  return {
    package: response.data.reviewSurface.package,
    requestId: response.meta.requestId,
    review: response.data.reviewSurface.review,
    target: response.data.reviewSurface.target,
  };
}
