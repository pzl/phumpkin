module.exports = {
  root: true,
  env: {
    browser: true,
    node: true
  },
  parserOptions: {
    parser: 'babel-eslint'
  },
  extends: [
    '@nuxtjs',
    'plugin:nuxt/recommended'
  ],
  // add your custom rules here
  rules: {
    'no-tabs': 0,
    'indent': 0,
    'quotes': 0,
    'comma-dangle': 0,
    'spaced-comment': 0,
  }
}
