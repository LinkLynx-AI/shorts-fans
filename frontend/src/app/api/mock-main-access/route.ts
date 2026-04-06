import { NextResponse } from "next/server";
import { z } from "zod";

import { getShortById } from "@/entities/short";
import {
  buildMockMainAccessEntryContext,
  buildMockMainPlaybackGrantContext,
  getMainPlaybackHref,
  getUnlockSurfaceByShortId,
  type MainPlaybackGrantKind,
} from "@/features/unlock-entry";
import { issueMockSignedToken, verifyMockSignedToken } from "@/shared/lib/mock-signed-token";

const mainAccessRequestSchema = z.object({
  acceptedAge: z.boolean(),
  acceptedTerms: z.boolean(),
  entryToken: z.string().min(1),
  fromShortId: z.string().min(1),
  mainId: z.string().min(1),
});

function buildFallbackHref(fromShortId?: string): string {
  if (fromShortId && getShortById(fromShortId)) {
    return `/shorts/${fromShortId}`;
  }

  return "/";
}

function buildDeniedResponse(fromShortId?: string, status = 403) {
  return NextResponse.json(
    {
      fallbackHref: buildFallbackHref(fromShortId),
    },
    {
      status,
    },
  );
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
      return "purchased";
    case "setup_required":
      if (
        (!unlock.setup.requiresAgeConfirmation || acceptedAge) &&
        (!unlock.setup.requiresTermsAcceptance || acceptedTerms)
      ) {
        return "purchased";
      }

      return null;
    case "unavailable":
      return null;
  }
}

/**
 * main 再生用の signed grant を発行する。
 */
export async function POST(request: Request) {
  const parsedBody = mainAccessRequestSchema.safeParse(await request.json().catch(() => null));

  if (!parsedBody.success) {
    return buildDeniedResponse(undefined, 400);
  }

  const { acceptedAge, acceptedTerms, entryToken, fromShortId, mainId } = parsedBody.data;

  const entryShort = getShortById(fromShortId);

  if (!entryShort || entryShort.canonicalMainId !== mainId) {
    return buildDeniedResponse(fromShortId, 404);
  }

  const hasValidEntryToken = verifyMockSignedToken(
    buildMockMainAccessEntryContext(mainId, fromShortId),
    entryToken,
  );

  if (!hasValidEntryToken) {
    return buildDeniedResponse(fromShortId);
  }

  const grantKind = resolveGrantKind(fromShortId, acceptedAge, acceptedTerms);

  if (!grantKind) {
    return buildDeniedResponse(fromShortId);
  }

  const grantToken = issueMockSignedToken(
    buildMockMainPlaybackGrantContext(mainId, fromShortId, grantKind),
  );

  return NextResponse.json({
    href: getMainPlaybackHref(mainId, fromShortId, grantToken),
  });
}
