import * as Dialog from "@radix-ui/react-dialog";
import { Slot } from "@radix-ui/react-slot";
import type {
  ComponentPropsWithoutRef,
  ReactElement,
  ReactNode,
} from "react";

import { cn } from "@/shared/lib";

type BottomSheetMenuProps = {
  children: ReactNode;
  description: string;
  title: string;
  trigger: ReactElement;
};

type BottomSheetMenuGroupProps = ComponentPropsWithoutRef<"div">;

type BottomSheetMenuActionProps = ComponentPropsWithoutRef<"button"> & {
  asChild?: boolean;
  tone?: "danger" | "default";
  withDivider?: boolean;
};

type BottomSheetMenuCloseProps = ComponentPropsWithoutRef<typeof Dialog.Close>;

/**
 * モバイル向けの共通 bottom sheet menu を表示する。
 */
export function BottomSheetMenu({
  children,
  description,
  title,
  trigger,
}: BottomSheetMenuProps) {
  return (
    <Dialog.Root>
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-y-0 left-1/2 z-40 w-full max-w-[408px] -translate-x-1/2 bg-[rgba(15,23,42,0.22)] backdrop-blur-[6px]" />
        <Dialog.Content className="fixed bottom-3 left-1/2 z-50 w-[calc(100vw-24px)] max-w-[384px] -translate-x-1/2 rounded-[28px] border border-border bg-white p-[10px_10px_12px] shadow-[0_20px_48px_rgba(15,23,42,0.18)]">
          <Dialog.Title className="sr-only">{title}</Dialog.Title>
          <Dialog.Description className="sr-only">{description}</Dialog.Description>

          <div
            aria-hidden="true"
            className="mx-auto mb-3 h-1 w-10 rounded-full bg-border-strong"
          />

          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

/**
 * bottom sheet menu 内の action group を表示する。
 */
export function BottomSheetMenuGroup({
  className,
  ...props
}: BottomSheetMenuGroupProps) {
  return (
    <div
      className={cn("rounded-[22px] bg-surface-subtle py-1", className)}
      {...props}
    />
  );
}

/**
 * bottom sheet menu の action row を表示する。
 */
export function BottomSheetMenuAction({
  asChild = false,
  className,
  tone = "default",
  withDivider = false,
  ...props
}: BottomSheetMenuActionProps) {
  const Comp = asChild ? Slot : "button";

  return (
    <Comp
      className={cn(
        "flex min-h-[54px] w-full items-center justify-between px-[18px] text-left text-[15px] font-semibold transition disabled:cursor-default disabled:opacity-55",
        tone === "danger"
          ? "text-[#b2394f] hover:bg-[#fff1f3]"
          : "text-foreground hover:bg-white",
        withDivider ? "border-t border-border" : "",
        className,
      )}
      {...(!asChild ? { type: "button" as const } : {})}
      {...props}
    />
  );
}

/**
 * bottom sheet menu 項目から close action を適用する。
 */
export function BottomSheetMenuClose(props: BottomSheetMenuCloseProps) {
  return <Dialog.Close {...props} />;
}
