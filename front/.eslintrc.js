module.exports = {
  root: true,
  env: {
    browser: true,
    node: true
  },
  parserOptions: {
    requireConfigFile: false,
    ecmaVersion: 15,
    parser: '@babel/eslint-parser'
  },
  extends: [
    // https://github.com/vuejs/eslint-plugin-vue#priority-a-essential-error-prevention
    // consider switching to `plugin:vue/strongly-recommended` or `plugin:vue/recommended` for stricter rules.
    'plugin:vue/essential'
  ],
  // required to lint *.vue files
  plugins: [
    'vue'
  ],
  // add your custom rules here
  rules: {
    'vue/multi-word-component-names': 0,
  }
}
