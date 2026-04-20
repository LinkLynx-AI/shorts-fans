import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("CCBillPaymentWidget", () => {
  beforeEach(() => {
    vi.resetModules();
    document.head.innerHTML = "";
    document.body.innerHTML = "";
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("retries script loading after a previous load failure", async () => {
    vi.spyOn(window.customElements, "get").mockReturnValue(undefined);
    const appendChildSpy = vi.spyOn(document.head, "appendChild");

    const { CCBillPaymentWidget } = await import("./ccbill-payment-widget");
    const session = {
      apiBaseUrl: "https://api.ccbill.com",
      apiKey: "frontend-token",
      clientAccount: "900100",
      currency: "JPY" as const,
      initialPeriod: "30",
      initialPrice: "1800.00",
      sessionToken: "signed-session-token",
      subAccount: "1",
    };

    const firstView = render(
      <CCBillPaymentWidget
        onPaymentTokenCreated={() => {}}
        session={session}
      />,
    );

    await screen.findByRole("alert");
    firstView.unmount();

    render(
      <CCBillPaymentWidget
        onPaymentTokenCreated={() => {}}
        session={session}
      />,
    );

    await screen.findByRole("alert");

    await waitFor(() => {
      const widgetScriptAppends = appendChildSpy.mock.calls.filter(([node]) => {
        return node instanceof HTMLScriptElement && node.dataset.ccbillPaymentWidget === "true";
      });

      expect(widgetScriptAppends).toHaveLength(2);
    });
  });
});
