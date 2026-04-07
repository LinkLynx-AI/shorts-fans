import { redirect } from "next/navigation";

import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { FanAuthEntryShell } from "@/widgets/fan-auth-entry-shell";

export default async function LoginPage() {
  const viewerState = await getFanAuthGateState();

  if (viewerState.hasSession) {
    redirect("/");
  }

  return <FanAuthEntryShell />;
}
