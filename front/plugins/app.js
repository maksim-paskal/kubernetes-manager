import Vue from 'vue'

import {
  BNavbar,
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
  FormCheckboxPlugin,
} from 'bootstrap-vue'

Vue.component('b-navbar', BNavbar)
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
Vue.component('b-form-select', BFormSelect)
Vue.component('b-dropdown', BDropdown)
Vue.component('b-dropdown-item', BDropdownItem)
Vue.component('b-dropdown-divider', BDropdownDivider)

Vue.use(ModalPlugin)
Vue.use(TooltipPlugin)
Vue.use(FormCheckboxPlugin)