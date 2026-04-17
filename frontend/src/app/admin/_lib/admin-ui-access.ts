import { notFound } from "next/navigation";

const adminUiEnabledValue = "1";
const developmentNodeEnv = "development";

export function isAdminUiEnabled(): boolean {
  return (
    process.env.NODE_ENV === developmentNodeEnv &&
    process.env.ADMIN_UI_ENABLED === adminUiEnabledValue
  );
}

export function assertAdminUiEnabled(): void {
  if (!isAdminUiEnabled()) {
    notFound();
  }
}
