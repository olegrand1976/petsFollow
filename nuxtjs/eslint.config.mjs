import withNuxt from './.nuxt/eslint.config.mjs'

// Gate CI léger : erreurs structurelles Vue ; le reste en warn pour éviter un big-bang.
export default withNuxt({
  rules: {
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/no-unused-vars': 'warn',
    '@typescript-eslint/no-dynamic-delete': 'warn',
    'no-console': 'off',
    'no-control-regex': 'warn',
    'no-useless-assignment': 'warn',
    'prefer-const': 'warn',
    'import/first': 'warn',
    'vue/multi-word-component-names': 'off',
    'vue/no-v-html': 'warn',
    'vue/require-default-prop': 'off',
    'vue/valid-v-bind': 'error',
    'vue/valid-v-if': 'error',
    'vue/valid-v-for': 'error',
    'vue/no-parsing-error': 'error',
    'vue/no-dupe-keys': 'error',
    'vue/no-duplicate-attributes': 'error',
    'vue/return-in-computed-property': 'error',
  },
}).append({
  ignores: [
    '.nuxt/**',
    '.output/**',
    'dist/**',
    'node_modules/**',
    'coverage/**',
    'playwright-report/**',
    'test-results/**',
    'data/**',
  ],
})
