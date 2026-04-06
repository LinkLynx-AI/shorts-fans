import { listFollowingItems } from "@/entities/fan-profile";
import { FollowingShell } from "@/widgets/following-shell";

export default function FollowingPage() {
  return <FollowingShell items={listFollowingItems()} />;
}
