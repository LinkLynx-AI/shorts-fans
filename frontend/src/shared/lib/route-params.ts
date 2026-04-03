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

/**
 * 許可された値に含まれる query param のみを返す。
 */
export function getEnumQueryParam<const T extends string>(
  value: string | readonly string[] | string[] | undefined,
  allowedValues: readonly T[],
): T | undefined {
  const singleValue = getSingleQueryParam(value);

  if (!singleValue) {
    return undefined;
  }

  return allowedValues.includes(singleValue as T) ? (singleValue as T) : undefined;
}
