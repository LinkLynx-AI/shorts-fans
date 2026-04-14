"use client";

import {
  useEffect,
  useRef,
} from "react";

import { useFanAuthDialog } from "../model/fan-auth-dialog-context";

/**
 * auth_required surface に到達したとき shared fan auth modal を開く。
 */
export function FanAuthRequiredDialogTrigger() {
  const hasOpenedRef = useRef(false);
  const { openFanAuthDialog } = useFanAuthDialog();

  useEffect(() => {
    if (hasOpenedRef.current) {
      return;
    }

    hasOpenedRef.current = true;
    openFanAuthDialog({
      closeBehavior: "back",
    });
  }, [openFanAuthDialog]);

  return null;
}
