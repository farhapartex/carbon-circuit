import { NextResponse, type NextRequest } from "next/server";
import { auth0 } from "@/lib/auth0";
import { blockingStep, tenancyOf } from "@/lib/session";

const publicPaths = ["/marketplace/retirements", "/track"];

const protectedPrefixes = [
  "/batches",
  "/claims",
  "/credits",
  "/dashboard",
  "/facilities",
  "/marketplace",
  "/notifications",
  "/settings",
  "/verifier",
];

const mandatoryOnboardingPaths = [
  "/onboarding/organization",
  "/onboarding/plan",
];

const covers = (prefix: string, pathname: string) =>
  pathname === prefix || pathname.startsWith(`${prefix}/`);

const matches = (prefixes: string[], pathname: string) =>
  prefixes.some((prefix) => covers(prefix, pathname));

export async function proxy(request: NextRequest) {
  const response = await auth0.middleware(request);
  const { pathname, origin } = request.nextUrl;

  if (pathname.startsWith("/auth/")) {
    return response;
  }

  const guarded =
    !matches(publicPaths, pathname) && matches(protectedPrefixes, pathname);
  const onboarding = covers("/onboarding", pathname);

  if (!guarded && !onboarding) {
    return response;
  }

  const session = await auth0.getSession(request);
  if (!session) {
    return NextResponse.redirect(new URL("/auth/login", origin));
  }

  const tenancy = tenancyOf(session);
  if (!tenancy) {
    return response;
  }

  const blocking = blockingStep(tenancy);

  if (blocking) {
    const destination = `/onboarding/${blocking}`;
    if (pathname === destination) {
      return response;
    }
    return NextResponse.redirect(new URL(destination, origin));
  }

  if (onboarding && matches(mandatoryOnboardingPaths, pathname)) {
    return NextResponse.redirect(new URL("/dashboard", origin));
  }

  return response;
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt).*)",
  ],
};
