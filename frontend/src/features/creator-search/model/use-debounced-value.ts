"use client";

import { startTransition, useEffect, useState } from "react";

/**
 * 入力値を一定時間遅らせて反映する。
 */
export function useDebouncedValue(value: string, delayMs: number): string {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      startTransition(() => {
        setDebouncedValue(value);
      });
    }, delayMs);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [delayMs, value]);

  return debouncedValue;
}
