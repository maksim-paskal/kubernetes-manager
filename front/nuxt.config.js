module.exports = {
  /*
  ** Headers of the page
  */
  head: {
    htmlAttrs: {
      lang: 'en',
    },
    title: 'Kubernetes Manager',
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1, shrink-to-fit=no' },
      { hid: 'description', name: 'description', content: 'Kubernetes Manager' }
    ],
    link: [
      { rel: 'icon', type: 'image/x-icon', href: '/favicon.png' }
    ]
  },
  components: [
    '~/components'
  ],
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
    '/api/': `${process.env.BACKEND_URL}`,
    '/oauth2/userinfo': `${process.env.BACKEND_URL}`
  },
  ssr: false,
  target: 'static',
  css: [
    'bootstrap/dist/css/bootstrap.min.css',
    'bootstrap-icons/font/bootstrap-icons.css',
    '~/css/main.css'
  ],
  /*
  ** Customize the progress bar color
  */
  loading: { color: '#6c757d', height: '7px', continuous: true },
  /*
  ** Build configuration
  */
  build: {
    publicPath: "/_nuxt/" + process.env.APPVERSION + "/",
    /*
    ** Run ESLint on save
    */
    extend(config, { isDev, isClient }) {
      config.devtool = 'source-map'

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

