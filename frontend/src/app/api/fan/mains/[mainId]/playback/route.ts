import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";

import { getMainById } from "@/entities/main";
import { getShortById } from "@/entities/short";
import { getCurrentViewerBootstrap, viewerSessionCookieName } from "@/entities/viewer";
import {
  parseMockMainPlaybackGrantContext,
} from "@/features/unlock-entry";
import {
  createMockSessionProof,
  readMockSignedToken,
} from "@/shared/lib/mock-signed-token";
import { getMainPlaybackPayloadById } from "@/widgets/main-playback-surface";

const playbackParamsSchema = z.object({
  mainId: z.string().min(1),
});

const playbackQuerySchema = z.object({
  fromShortId: z.string().min(1),
  grant: z.string().min(1),
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

function buildSuccessResponse(payload: NonNullable<ReturnType<typeof getMainPlaybackPayloadById>>) {
  return NextResponse.json({
    data: payload,
    error: null,
    meta: {
      page: null,
      requestId: "req_main_playback_001",
    },
  });
}

/**
 * current session grant 済み main playback payload を返す。
 */
export async function GET(
  request: NextRequest,
  context: { params: Promise<{ mainId: string }> },
) {
  const sessionToken = request.cookies.get(viewerSessionCookieName)?.value;

  if (!sessionToken) {
    return buildErrorResponse(
      401,
      "req_main_playback_auth_required_001",
      "auth_required",
      "main playback requires authentication",
    );
  }

  const currentViewer = await getCurrentViewerBootstrap({ sessionToken }).catch(() => null);

  if (!currentViewer) {
    return buildErrorResponse(
      401,
      "req_main_playback_auth_required_001",
      "auth_required",
      "main playback requires authentication",
    );
  }

  const parsedParams = playbackParamsSchema.safeParse(await context.params);

  if (!parsedParams.success) {
    return buildErrorResponse(
      404,
      "req_main_playback_not_found_001",
      "not_found",
      "main or short was not found",
    );
  }

  const parsedQuery = playbackQuerySchema.safeParse({
    fromShortId: request.nextUrl.searchParams.get("fromShortId"),
    grant: request.nextUrl.searchParams.get("grant"),
  });

  if (!parsedQuery.success) {
    return buildErrorResponse(
      404,
      "req_main_playback_not_found_001",
      "not_found",
      "main or short was not found",
    );
  }

  const { mainId } = parsedParams.data;
  const { fromShortId, grant } = parsedQuery.data;
  const main = getMainById(mainId);
  const entryShort = getShortById(fromShortId);
  const sessionProof = createMockSessionProof(sessionToken);

  if (!main || !entryShort || entryShort.canonicalMainId !== mainId) {
    return buildErrorResponse(
      404,
      "req_main_playback_not_found_001",
      "not_found",
      "main or short was not found",
    );
  }

  const grantPayload = readMockSignedToken(grant);
  const parsedGrantContext = grantPayload
    ? parseMockMainPlaybackGrantContext(grantPayload.context)
    : null;

  if (
    grantPayload?.sessionProof !== sessionProof ||
    !parsedGrantContext ||
    parsedGrantContext.mainId !== mainId ||
    parsedGrantContext.fromShortId !== fromShortId
  ) {
    return buildErrorResponse(
      403,
      "req_main_playback_locked_001",
      "main_locked",
      "main playback is locked",
    );
  }

  const payload = getMainPlaybackPayloadById(
    mainId,
    fromShortId,
    parsedGrantContext.grantKind,
  );

  if (!payload) {
    return buildErrorResponse(
      404,
      "req_main_playback_not_found_001",
      "not_found",
      "main or short was not found",
    );
  }

  return buildSuccessResponse(payload);
}
