const js = require("@eslint/js");

module.exports = [
  js.configs.recommended,
  {
    files: ["**/*.js"],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
      globals: {
        console: "readonly",
        document: "readonly",
        window: "readonly",
        fetch: "readonly",
        localStorage: "readonly",
        sessionStorage: "readonly",
        alert: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
      },
    },
    rules: {
      // Error prevention
      "no-unused-vars": ["error", { 
        "argsIgnorePattern": "^_",
        "varsIgnorePattern": "^_" 
      }],
      "no-undef": "error",
      "no-console": "off", // Allow console.log for debugging
      
      // Code quality
      "prefer-const": "error",
      "no-var": "error",
      "eqeqeq": ["error", "always"],
      "curly": ["error", "all"],
      
      // Style consistency
      "indent": ["error", 2],
      "quotes": ["error", "single", { "allowTemplateLiterals": true }],
      "semi": ["error", "always"],
      "comma-trailing": ["error", "never"],
      
      // Best practices
      "no-eval": "error",
      "no-implied-eval": "error",
      "no-with": "error",
      "no-fallthrough": "error",
      "default-case": "error",
    },
  },
];