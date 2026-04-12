/**
 * modal の価格入力値を整数円へ変換する。
 */
export function parseCreatorWorkspaceMainPriceInput(value: string): number | null {
  const normalizedValue = value.trim();

  if (!/^[1-9]\d*$/.test(normalizedValue)) {
    return null;
  }

  return Number.parseInt(normalizedValue, 10);
}

/**
 * modal 初期表示用に価格を文字列化する。
 */
export function formatCreatorWorkspaceMainPriceInput(priceJpy: number): string {
  return String(priceJpy);
}
