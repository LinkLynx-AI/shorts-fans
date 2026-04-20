"use client";

import { createElement, useEffect, useRef, useState } from "react";

import { cn } from "@/shared/lib";
import type { CardSetupSession } from "../api/contracts";

export type CCBillPaymentWidgetProps = {
  className?: string | undefined;
  onPaymentTokenCreated: (paymentTokenId: string) => void;
  session: CardSetupSession;
};

const ccbillPaymentWidgetScriptSrc = "https://js.ccbill.com/payment-widget/v1/index.js";
let ccbillPaymentWidgetScriptPromise: Promise<void> | null = null;
const ccbillPaymentWidgetScriptStatusAttribute = "data-ccbill-payment-widget-status";

function attachScriptLoadListeners(script: HTMLScriptElement, onError: () => void): Promise<void> {
  return new Promise((resolve, reject) => {
    script.addEventListener(
      "load",
      () => {
        script.dataset.ccbillPaymentWidgetStatus = "loaded";
        resolve();
      },
      { once: true },
    );
    script.addEventListener(
      "error",
      () => {
        script.dataset.ccbillPaymentWidgetStatus = "error";
        onError();
        reject(new Error("ccbill widget script failed"));
      },
      { once: true },
    );
  });
}

function loadCCBillPaymentWidgetScript(): Promise<void> {
  if (typeof window === "undefined") {
    return Promise.resolve();
  }

  if (window.customElements?.get("ccb-payment-widget")) {
    return Promise.resolve();
  }

  const existing = document.querySelector<HTMLScriptElement>('script[data-ccbill-payment-widget="true"]');
  if (existing) {
    const status = existing.getAttribute(ccbillPaymentWidgetScriptStatusAttribute);
    if (status === "loaded") {
      return Promise.resolve();
    }
    if (status === "error") {
      existing.remove();
    } else if (ccbillPaymentWidgetScriptPromise) {
      return ccbillPaymentWidgetScriptPromise;
    } else {
      const scriptPromise = attachScriptLoadListeners(existing, () => {
        ccbillPaymentWidgetScriptPromise = null;
        existing.remove();
      });
      ccbillPaymentWidgetScriptPromise = scriptPromise;

      return scriptPromise;
    }
  } else if (ccbillPaymentWidgetScriptPromise) {
    return ccbillPaymentWidgetScriptPromise;
  }

  const script = document.createElement("script");
  script.async = true;
  script.dataset.ccbillPaymentWidget = "true";
  script.dataset.ccbillPaymentWidgetStatus = "loading";
  script.src = ccbillPaymentWidgetScriptSrc;
  const scriptPromise = attachScriptLoadListeners(script, () => {
    ccbillPaymentWidgetScriptPromise = null;
    script.remove();
  });
  ccbillPaymentWidgetScriptPromise = scriptPromise;
  document.head.appendChild(script);

  return scriptPromise;
}

/**
 * CCBill payment widget を埋め込み、payment token 作成イベントを feature 境界に閉じ込める。
 */
export function CCBillPaymentWidget({
  className,
  onPaymentTokenCreated,
  session,
}: CCBillPaymentWidgetProps) {
  const widgetRef = useRef<HTMLElement | null>(null);
  const [status, setStatus] = useState<"loading" | "ready" | "error">(() =>
    typeof window !== "undefined" && window.customElements?.get("ccb-payment-widget") ? "ready" : "loading",
  );

  useEffect(() => {
    let cancelled = false;

    if (typeof window !== "undefined" && window.customElements?.get("ccb-payment-widget")) {
      void Promise.resolve().then(() => {
        if (!cancelled) {
          setStatus("ready");
        }
      });
      return;
    }

    void loadCCBillPaymentWidgetScript()
      .then(() => {
        if (!cancelled) {
          setStatus("ready");
        }
      })
      .catch(() => {
        if (!cancelled) {
          setStatus("error");
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (status !== "ready" || !widgetRef.current) {
      return;
    }

    const widget = widgetRef.current;
    const handleTokenCreated = (event: Event) => {
      const paymentTokenId = (
        event as CustomEvent<{
          paymentToken?: {
            paymentTokenId?: string;
          };
        }>
      ).detail?.paymentToken?.paymentTokenId;

      if (paymentTokenId) {
        onPaymentTokenCreated(paymentTokenId);
      }
    };

    widget.addEventListener("tokenCreated", handleTokenCreated as EventListener);

    return () => {
      widget.removeEventListener("tokenCreated", handleTokenCreated as EventListener);
    };
  }, [onPaymentTokenCreated, status]);

  if (status === "error") {
    return (
      <div
        className={cn(
          "rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]",
          className,
        )}
        role="alert"
      >
        カード入力フォームを読み込めませんでした。時間を置いてから再度お試しください。
      </div>
    );
  }

  if (status === "loading") {
    return (
      <div
        className={cn(
          "rounded-[20px] border border-[#bae7ff]/90 bg-white/86 px-4 py-4 text-sm text-muted",
          className,
        )}
      >
        カード入力フォームを読み込み中です。
      </div>
    );
  }

  return (
    <div className={className}>
      {createElement("ccb-payment-widget", {
        apiBaseUrl: session.apiBaseUrl,
        apiKey: session.apiKey,
        clientAccount: session.clientAccount,
        currency: session.currency,
        description: "Shorts Fans main unlock",
        initialPeriod: session.initialPeriod,
        initialPrice: session.initialPrice,
        language: "ja",
        ref: widgetRef,
        subAccount: session.subAccount,
        theme: "ccb-light",
      })}
    </div>
  );
}
