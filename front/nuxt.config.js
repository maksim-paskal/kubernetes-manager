module.exports = {
  /*
  ** Headers of the page
  */
  head: {
    htmlAttrs: {
      lang: 'en',
    },
    title: 'Kubernetes manager',
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { hid: 'description', name: 'description', content: 'Kubernetes manager' }
    ]
  },
  plugins: [
    '~plugins/app.js'
  ],
  modules: [
    '@nuxtjs/axios',
    '@nuxtjs/sentry'
  ],
  axios: {
    proxy: true
  },
  sentry: {
    dsn: 'https://__setry_id__@__setry_server__/1',
    config: {}, // Additional config
  },
  proxy: {
    '/api/': `${process.env.BACKEND_URL}`
  },
  mode: 'spa',
  css: [
    'bootstrap/dist/css/bootstrap.min.css'
  ],
  /*
  ** Customize the progress bar color
  */
  loading: { color: '#3B8070' },
  /*
  ** Build configuration
  */
  build: {
    /*
    ** Run ESLint on save
    */
    extend(config, { isDev, isClient }) {
      if (isDev && isClient) {
        config.module.rules.push({
          enforce: 'pre',
          test: /\.(js|vue)$/,
          loader: 'eslint-loader',
          exclude: /(node_modules)/
        })
      }
    }
  }
}

