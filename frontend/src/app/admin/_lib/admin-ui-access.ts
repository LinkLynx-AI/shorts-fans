import { notFound } from "next/navigation";

const adminUiEnabledValue = "1";

export function isAdminUiEnabled(): boolean {
  return process.env.ADMIN_UI_ENABLED === adminUiEnabledValue;
}

export function assertAdminUiEnabled(): void {
  if (!isAdminUiEnabled()) {
    notFound();
  }
}
