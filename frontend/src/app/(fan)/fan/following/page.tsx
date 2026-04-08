import { cookies } from "next/headers";

import { fetchFanProfileFollowingPage } from "@/entities/fan-profile";
import { viewerSessionCookieName } from "@/entities/viewer";
import {
  FanAuthRequiredDialogTrigger,
  isAuthRequiredApiError,
} from "@/features/fan-auth";
import { FollowingShell } from "@/widgets/following-shell";

export default async function FollowingPage() {
  const cookieStore = await cookies();
  const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
  let response;

  try {
    response = await fetchFanProfileFollowingPage({ sessionToken });
  } catch (error) {
    if (isAuthRequiredApiError(error)) {
      return <FanAuthRequiredDialogTrigger />;
    }

    throw error;
  }

  return <FollowingShell items={response.items} />;
}
