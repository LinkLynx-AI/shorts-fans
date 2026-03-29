import { z } from "zod";

const clientEnvSchema = z.object({
  NEXT_PUBLIC_API_BASE_URL: z.string().url(),
});

export type ClientEnv = z.infer<typeof clientEnvSchema>;

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
