import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const aliasOnlyImports = {
  patterns: [
    {
      group: ["../*", "../**"],
      message:
        "Reach across directories with the @/ alias so tier boundaries stay enforceable.",
    },
  ],
};

const forbidTiers = (tiers, message) => ({
  patterns: [
    ...aliasOnlyImports.patterns,
    { group: tiers.map((tier) => `@/components/${tier}/*`), message },
  ],
});

const designSystemGuards = [
  {
    selector: "JSXAttribute[name.name='style']",
    message:
      "Style through Tailwind tokens. An inline style bypasses the design system.",
  },
  {
    selector:
      "Literal[value=/#(?:[0-9a-fA-F]{3,4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/]",
    message:
      "Raw hex belongs in the token definitions only. Use a design token.",
  },
];

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    files: ["src/**/*.{ts,tsx}"],
    rules: {
      "no-restricted-syntax": ["error", ...designSystemGuards],
    },
  },
  {
    files: ["src/components/ui/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            ...forbidTiers(
              ["shared", "features", "layout"],
              "A primitive never imports from a higher tier.",
            ).patterns,
            {
              group: ["@/stores/*", "@/hooks/*"],
              message:
                "A primitive is stateless. Pass what it needs in as props.",
            },
          ],
        },
      ],
    },
  },
  {
    files: ["src/components/shared/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        forbidTiers(
          ["features", "layout"],
          "A shared component never imports a domain-specific or layout component.",
        ),
      ],
    },
  },
  {
    files: ["src/components/layout/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        forbidTiers(
          ["features"],
          "A layout is composed from primitives and shared components only.",
        ),
      ],
    },
  },
  {
    files: ["src/components/features/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        forbidTiers(
          ["layout"],
          "A domain component never imports a layout. Layouts compose domains, not the reverse.",
        ),
      ],
    },
  },
  globalIgnores([".next/**", "out/**", "build/**", "next-env.d.ts"]),
]);

export default eslintConfig;
