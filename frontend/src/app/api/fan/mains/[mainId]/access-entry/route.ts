import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";

import { getShortById } from "@/entities/short";
import { getCurrentViewerBootstrap, viewerSessionCookieName } from "@/entities/viewer";
import {
  buildMockMainAccessEntryContext,
  buildMockMainPlaybackGrantContext,
  getMainPlaybackHref,
  getUnlockSurfaceByShortId,
  type MainPlaybackGrantKind,
} from "@/features/unlock-entry";
import {
  createMockSessionProof,
  issueMockSignedToken,
  verifyMockSignedToken,
} from "@/shared/lib/mock-signed-token";

const accessEntryParamsSchema = z.object({
  mainId: z.string().min(1),
});

const accessEntryRequestSchema = z.object({
  acceptedAge: z.boolean(),
  acceptedTerms: z.boolean(),
  entryToken: z.string().min(1),
  fromShortId: z.string().min(1),
});

function buildErrorResponse(
  status: number,
  requestId: string,
  code: string,
  message: string,
) {
  return NextResponse.json(
    {
      data: null,
      meta: {
        page: null,
        requestId,
      },
      error: {
        code,
        message,
      },
    },
    {
      status,
    },
  );
}

function buildSuccessResponse(href: string) {
  return NextResponse.json({
    data: {
      href,
    },
    meta: {
      page: null,
      requestId: "req_main_access_entry_issued_001",
    },
    error: null,
  });
}

function resolveGrantKind(
  fromShortId: string,
  acceptedAge: boolean,
  acceptedTerms: boolean,
): MainPlaybackGrantKind | null {
  const unlock = getUnlockSurfaceByShortId(fromShortId);

  if (!unlock) {
    return null;
  }

  switch (unlock.unlockCta.state) {
    case "owner_preview":
      return "owner";
    case "continue_main":
    case "unlock_available":
      return "unlocked";
    case "setup_required":
      if (
        (!unlock.setup.requiresAgeConfirmation || acceptedAge) &&
        (!unlock.setup.requiresTermsAcceptance || acceptedTerms)
      ) {
        return "unlocked";
      }

      return null;
    case "unavailable":
      return null;
  }
}

/**
 * main 再生用の signed grant を発行する。
 */
export async function POST(
  request: NextRequest,
  context: { params: Promise<{ mainId: string }> },
) {
  const sessionToken = request.cookies.get(viewerSessionCookieName)?.value;

  if (!sessionToken) {
    return buildErrorResponse(
      401,
      "req_main_access_entry_auth_required_001",
      "auth_required",
      "main playback requires authentication",
    );
  }

  const currentViewer = await getCurrentViewerBootstrap({ sessionToken }).catch(() => null);

  if (!currentViewer) {
    return buildErrorResponse(
      401,
      "req_main_access_entry_auth_required_001",
      "auth_required",
      "main playback requires authentication",
    );
  }

  const parsedParams = accessEntryParamsSchema.safeParse(await context.params);

  if (!parsedParams.success) {
    return buildErrorResponse(
      404,
      "req_main_access_entry_not_found_001",
      "not_found",
      "main or short was not found",
    );
  }

  const parsedBody = accessEntryRequestSchema.safeParse(await request.json().catch(() => null));

  if (!parsedBody.success) {
    return buildErrorResponse(
      400,
      "req_main_access_entry_invalid_request_001",
      "invalid_request",
      "main access entry request was invalid",
    );
  }

  const { mainId } = parsedParams.data;
  const { acceptedAge, acceptedTerms, entryToken, fromShortId } = parsedBody.data;
  const entryShort = getShortById(fromShortId);

  if (!entryShort || entryShort.canonicalMainId !== mainId) {
    return buildErrorResponse(
      404,
      "req_main_access_entry_not_found_001",
      "not_found",
      "main or short was not found",
    );
  }

  const hasValidEntryToken = verifyMockSignedToken(
    buildMockMainAccessEntryContext(mainId, fromShortId),
    entryToken,
  );

  if (!hasValidEntryToken) {
    return buildErrorResponse(
      403,
      "req_main_access_entry_locked_001",
      "main_locked",
      "main access entry could not be issued",
    );
  }

  const grantKind = resolveGrantKind(fromShortId, acceptedAge, acceptedTerms);

  if (!grantKind) {
    return buildErrorResponse(
      403,
      "req_main_access_entry_locked_001",
      "main_locked",
      "main access entry could not be issued",
    );
  }

  const grantToken = issueMockSignedToken(
    buildMockMainPlaybackGrantContext(mainId, fromShortId, grantKind),
    {
      sessionProof: createMockSessionProof(sessionToken),
    },
  );

  return buildSuccessResponse(getMainPlaybackHref(mainId, fromShortId, grantToken));
}
