import type { UnlockSurfaceModel } from "./unlock-entry";

/**
 * payment bypass は backend から返る unlock surface の flag を唯一の source of truth にする。
 */
export function isDevelopmentPaymentBypassEnabled(
  unlock: Pick<UnlockSurfaceModel, "purchase"> | null | undefined,
) {
  return unlock?.purchase.paymentBypassEnabled === true;
}
