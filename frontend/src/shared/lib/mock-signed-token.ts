import { createHmac, randomUUID, timingSafeEqual } from "crypto";

const MOCK_SIGNED_TOKEN_SECRET = "shorts-fans:mock-main-access:v1";
const DEFAULT_TOKEN_TTL_MS = 15 * 60 * 1000;

type SignedTokenPayload = {
  context: string;
  expiresAtMs: number;
  nonce: string;
};

type SignedTokenOptions = {
  nowMs?: number;
  ttlMs?: number;
};

function encodePayload(payload: SignedTokenPayload): string {
  return Buffer.from(JSON.stringify(payload)).toString("base64url");
}

function decodePayload(value: string): SignedTokenPayload | null {
  try {
    const parsed = JSON.parse(Buffer.from(value, "base64url").toString("utf8"));

    if (
      typeof parsed !== "object" ||
      parsed === null ||
      typeof parsed.context !== "string" ||
      typeof parsed.expiresAtMs !== "number" ||
      typeof parsed.nonce !== "string"
    ) {
      return null;
    }

    return parsed as SignedTokenPayload;
  } catch {
    return null;
  }
}

function signValue(value: string): string {
  return createHmac("sha256", MOCK_SIGNED_TOKEN_SECRET).update(value).digest("base64url");
}

function hasValidSignature(encodedPayload: string, providedSignature: string): boolean {
  const expectedSignature = signValue(encodedPayload);

  try {
    return timingSafeEqual(Buffer.from(expectedSignature), Buffer.from(providedSignature));
  } catch {
    return false;
  }
}

function getVerifiedPayload(
  token: string,
  options?: Pick<SignedTokenOptions, "nowMs">,
): SignedTokenPayload | null {
  const [encodedPayload, signature] = token.split(".");

  if (!encodedPayload || !signature || !hasValidSignature(encodedPayload, signature)) {
    return null;
  }

  const payload = decodePayload(encodedPayload);

  if (!payload) {
    return null;
  }

  const nowMs = options?.nowMs ?? Date.now();

  return payload.expiresAtMs >= nowMs ? payload : null;
}

/**
 * server 側で検証できる mock signed token を発行する。
 */
export function issueMockSignedToken(context: string, options?: SignedTokenOptions): string {
  const nowMs = options?.nowMs ?? Date.now();
  const ttlMs = options?.ttlMs ?? DEFAULT_TOKEN_TTL_MS;
  const encodedPayload = encodePayload({
    context,
    expiresAtMs: nowMs + ttlMs,
    nonce: randomUUID(),
  });
  const signature = signValue(encodedPayload);

  return `${encodedPayload}.${signature}`;
}

/**
 * signed token が context に対して有効か判定する。
 */
export function verifyMockSignedToken(
  context: string,
  token: string,
  options?: Pick<SignedTokenOptions, "nowMs">,
): boolean {
  const payload = getVerifiedPayload(token, options);
  return payload?.context === context;
}

/**
 * signed token が有効なら payload を返す。
 */
export function readMockSignedToken(
  token: string,
  options?: Pick<SignedTokenOptions, "nowMs">,
): SignedTokenPayload | null {
  return getVerifiedPayload(token, options);
}
