import { z } from "zod";

const clientEnvSchema = z.object({
  NEXT_PUBLIC_API_BASE_URL: z.string().url(),
});

export type ClientEnv = z.infer<typeof clientEnvSchema>;
export type OptionalClientEnv = {
  NEXT_PUBLIC_API_BASE_URL?: string;
};

/**
 * クライアント公開環境変数を検証する。
 */
export function parseClientEnv(input: {
  NEXT_PUBLIC_API_BASE_URL: string | undefined;
}): ClientEnv {
  return clientEnvSchema.parse(input);
}

/**
 * クライアント公開環境変数を取得する。
 */
export function getClientEnv(): ClientEnv {
  return parseClientEnv({
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
  });
}

/**
 * クライアント公開環境変数を optional 契約で検証する。
 */
export function parseOptionalClientEnv(input: {
  NEXT_PUBLIC_API_BASE_URL: string | undefined;
}): OptionalClientEnv {
  if (input.NEXT_PUBLIC_API_BASE_URL === undefined) {
    return {};
  }

  return parseClientEnv(input);
}

/**
 * optional なクライアント公開環境変数を取得する。
 */
export function getOptionalClientEnv(): OptionalClientEnv {
  return parseOptionalClientEnv({
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
  });
}
