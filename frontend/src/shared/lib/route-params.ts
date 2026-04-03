/**
 * URL query param から単一値を取り出す。
 */
export function getSingleQueryParam(
  value: string | readonly string[] | string[] | undefined,
): string | undefined {
  if (Array.isArray(value)) {
    return value[0];
  }

  if (typeof value === "string") {
    return value;
  }

  return undefined;
}
