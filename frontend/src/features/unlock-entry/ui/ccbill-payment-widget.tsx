"use client";

import { createElement, useEffect, useRef, useState } from "react";

import { cn } from "@/shared/lib";

type PaymentWidgetSession = {
  apiBaseUrl: string;
  apiKey: string;
  clientAccount: string;
  currency: "JPY";
  initialPeriod: string;
  initialPrice: string;
  subAccount: string;
};

export type CCBillPaymentWidgetProps = {
  className?: string | undefined;
  onPaymentTokenCreated: (paymentTokenId: string) => void;
  session: PaymentWidgetSession;
};

const ccbillPaymentWidgetScriptSrc = "https://js.ccbill.com/payment-widget/v1/index.js";
let ccbillPaymentWidgetScriptPromise: Promise<void> | null = null;

function loadCCBillPaymentWidgetScript(): Promise<void> {
  if (typeof window === "undefined") {
    return Promise.resolve();
  }

  if (window.customElements?.get("ccb-payment-widget")) {
    return Promise.resolve();
  }

  if (ccbillPaymentWidgetScriptPromise) {
    return ccbillPaymentWidgetScriptPromise;
  }

  ccbillPaymentWidgetScriptPromise = new Promise((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>('script[data-ccbill-payment-widget="true"]');

    if (existing) {
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener("error", () => reject(new Error("ccbill widget script failed")), { once: true });
      return;
    }

    const script = document.createElement("script");
    script.async = true;
    script.dataset.ccbillPaymentWidget = "true";
    script.src = ccbillPaymentWidgetScriptSrc;
    script.addEventListener("load", () => resolve(), { once: true });
    script.addEventListener("error", () => reject(new Error("ccbill widget script failed")), { once: true });
    document.head.appendChild(script);
  });

  return ccbillPaymentWidgetScriptPromise;
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
