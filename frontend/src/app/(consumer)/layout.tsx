import type { ReactNode } from "react";

import { ConsumerShell } from "@/widgets/consumer-shell";

export default function ConsumerLayout({ children }: { children: ReactNode }) {
  return <ConsumerShell>{children}</ConsumerShell>;
}
