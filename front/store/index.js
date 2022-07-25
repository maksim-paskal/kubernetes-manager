import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

const store = () => new Vuex.Store({
  state: {
    config: {},
    environment: {},
    user: {},
    selectedDropdowns: Object.create(null),
    componentLoaded: Object.create(null),
  },
  mutations: {
    setEnvironment(state, environment) {
      state.environment = environment
    },
    setConfig(state, config) {
      state.config = config
    },
    setUser(state, user) {
      state.user = user
    },
    clearDropDowns(state) {
      state.selectedDropdowns = Object.create(null)
    },
    setDropDown(state,data) {
      Vue.set(state.selectedDropdowns, data.id, data.selected);
    },
    setComponentLoaded(state,data) {
      Vue.set(state.componentLoaded, data.id, data.loaded);
    }
  }
})

export default store