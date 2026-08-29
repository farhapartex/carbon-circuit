/** @type {import("prettier").Config} */
const config = {
  plugins: ["prettier-plugin-tailwindcss"],
  tailwindStylesheet: "./src/app/globals.css",
  semi: true,
  singleQuote: false,
  trailingComma: "all",
  printWidth: 80,
};

export default config;
