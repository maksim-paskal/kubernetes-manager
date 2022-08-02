import Vue from 'vue'

import {
  BNavbar,
  BNavbarNav,
  BAlert,
  BFormInput,
  BTable,
  BButton,
  BSpinner,
  ModalPlugin,
  BCard,
  BCardText,
  BTabs,
  BTab,
  BForm,
  BFormTextarea,
  BContainer,
  BRow,
  BCol,
  BInputGroup,
  TooltipPlugin,
  BFormSelect,
  BDropdown,
  BDropdownItem,
  BDropdownDivider,
  BDropdownForm,
  FormCheckboxPlugin,
  BLink
} from 'bootstrap-vue'

Vue.component('b-navbar', BNavbar)
Vue.component('b-navbar-nav', BNavbarNav)
Vue.component('b-alert', BAlert)
Vue.component('b-form-input', BFormInput)
Vue.component('b-table', BTable)
Vue.component('b-button', BButton)
Vue.component('b-spinner', BSpinner)
Vue.component('b-card', BCard)
Vue.component('b-card-text', BCardText)
Vue.component('b-tabs', BTabs)
Vue.component('b-tab', BTab)
Vue.component('b-form-textarea', BFormTextarea)
Vue.component('b-container', BContainer)
Vue.component('b-row', BRow)
Vue.component('b-col', BCol)
Vue.component('b-input-group', BInputGroup)
Vue.component('b-form', BForm)
Vue.component('b-form-select', BFormSelect)
Vue.component('b-dropdown', BDropdown)
Vue.component('b-dropdown-item', BDropdownItem)
Vue.component('b-dropdown-divider', BDropdownDivider)
Vue.component('b-dropdown-form', BDropdownForm)
Vue.component('b-link', BLink)

Vue.use(ModalPlugin)
Vue.use(TooltipPlugin)
Vue.use(FormCheckboxPlugin)

const KubernetesManager = "kubernetes-manager"

Vue.mixin({
  data() {
    return {
      callIsLoading: "",
      errorText: "",
      infoText: "",
    }
  },
  computed: {
    config() {
      return this.$store.state.config
    },
    environment() {
      return this.$store.state.environment;
    },
    user() {
      return this.$store.state.user
    },
  },
  methods: {
    const(data) {
      return {
        LabelEnvironmentName: `${KubernetesManager}/environment-name`,
        LabelRequiredPods: `${KubernetesManager}/requiredRunningPodsCount`,
        LabelLikedPrefix: `${KubernetesManager}/user-liked-`,
        LabelLiked: `${KubernetesManager}/user-liked-${data && data.user ? data.user : "unknown"}`,
        LabelCreator: `${KubernetesManager}/user-creator-${data && data.user ? data.user : "unknown"}`,
      }
    },
    namespaceAnnotation(name) {
      if (this.environment.NamespaceAnnotations && this.environment.NamespaceAnnotations[name]) {
        return this.environment.NamespaceAnnotations[name]
      }

      return ""
    },
    namespaceLabel(name) {
      if (this.environment.NamespaceLabels && this.environment.NamespaceLabels[name]) {
        return this.environment.NamespaceLabels[name]
      }

      return ""
    },
    async loadUser(endpoint) {
      try {
        const info = await fetch(`${endpoint}`);
        if (info.ok) {
          const data = await info.json();
          this.$store.commit("setUser", data);
        } else {
          throw Error(await info.text());
        }
      } catch (e) {
        this.errorText = e.message
      }
    },
    async loadEnvironment(ID) {
      try {
        const info = await fetch(`/api/${ID}/info`);
        if (info.ok) {
          const data = await info.json();
          this.$store.commit("setEnvironment", data.Result);
        } else {
          throw Error(await info.text());
        }
      } catch (e) {
        this.errorText = e.message
      }
    },
    async loadConfig() {
      try {
        const config = await fetch("/api/front-config")
        if (config.ok) {
          const data = await config.json()
          this.$store.commit("setConfig", data.Result)
        } else {
          throw Error(await config.text())
        }
      } catch (e) {
        this.errorText = e.message
      }
    },
    async call(op, data, onTop = true) {
      if (op == "make-delete" || op == "make-delete-service") {
        const realy = await this.$bvModal.msgBoxConfirm("Realy?");
        if (!realy) return;
      }

      return this.callEndpoint(`/api/${this.$route.params.environmentID}/${op}`, data, onTop)
    },
    async callEndpoint(endpoint, data, onTop) {
      this.infoText = null
      this.errorText = null
      this.callIsLoading = true

      try {
        this.infoText = null
        const result = await fetch(endpoint, {
          method: "POST",
          body: JSON.stringify(data),
        });
        if (result.ok) {
          const resultData = await result.json()
          this.infoText = resultData.Result
        } else {
          throw Error(await result.text())
        }
      } catch (e) {
        this.errorText = e.message
      } finally {
        this.callIsLoading = false

        if (onTop) {
          window.scrollTo(0, 0)
        }
      }
    },
  }
})