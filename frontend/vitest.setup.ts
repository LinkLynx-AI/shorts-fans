import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { createElement, type AnchorHTMLAttributes, type ReactNode } from "react";
import { afterEach, vi } from "vitest";

type MockedLinkProps = AnchorHTMLAttributes<HTMLAnchorElement> & {
  children?: ReactNode;
  href: string;
};

const mockedUsePathname = vi.fn(() => "/");

afterEach(() => {
  cleanup();
});

vi.mock("next/navigation", () => ({
  usePathname: mockedUsePathname,
}));

vi.mock("next/link", () => ({
  default: ({ children, href, ...props }: MockedLinkProps) =>
    createElement(
      "a",
      {
        ...props,
        href,
      },
      children,
    ),
}));
