import { redirect } from "next/navigation";

import { assertAdminUiEnabled } from "./_lib/admin-ui-access";

export default function AdminPage() {
  assertAdminUiEnabled();
  redirect("/admin/creator-reviews");
}
