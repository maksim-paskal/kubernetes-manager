<template>
  <div class="detail-tab">
    <b-alert variant="warning" show>
      <b-button @click="makeAPICall('disableHPA', 'none')">Disable autoscaling</b-button>&nbsp;For proper usage
      you must disable autoscaler
    </b-alert>
    <DropDown id="debugDropdownContainers" style="margin-bottom:10px" ref="debugDropdownContainers" default="backend"
      text="Select POD" :endpoint="`/api/${this.$route.params.environmentID}/containers`" />

    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-spinner v-if="$fetchState.pending" variant="primary" />
    <div v-else-if="selectedContainer">
      <b-card-text>
        XDEBUG is
        <strong>{{ this.debug_enabled }}</strong>
      </b-card-text>
      <b-card-text>extra setting in php-fpm.conf</b-card-text>

      <b-form-textarea spellcheck="false" rows="10" style="width:100%" v-model="data.PhpFpmSettings"
        placeholder="Enter something..." />

      <b-card-text>
        Use
        <strong>ngrok tcp 9000</strong>
      </b-card-text>
      <b-button @click="refresh()">Refresh</b-button>
      <b-dropdown id="dropdown-1" offset="25" text="Templates" class="m-2">
        <b-dropdown-item v-for="(item, index) in this.config.DebugTemplates" :key="index"
          @click="templateAction(item.Data)">{{ item.Display }}</b-dropdown-item>
      </b-dropdown>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  mounted() {
    this.$nuxt.$on('component::DropDown::selected', (data) => {
      if (data.id === 'debugDropdownContainers') {
        this.$fetch()
      }
    })
  },
  async fetch() {
    if (!this.selectedContainer) return;
    location.hash = this.selectedContainer;

    const result = await fetch(`/api/${this.$route.params.environmentID}/container-info?container=${this.selectedContainer}`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  data() {
    return {
      data: {}
    }
  },
  methods: {
    async templateAction(data) {
      this.data.PhpFpmSettings += `${data}\n`;
    },
    refresh() {
      this.$refs.debugDropdownContainers.isLoaded = false;
      this.$refs.debugDropdownContainers.select("");
    }
  },
  computed: {
    debug_enabled() {
      if (this.data.XdebugEnabled == "true") {
        return "ENABLED"
      } else {
        return "DISABLED"
      }
    },
    selectedContainer() {
      return this.$store.state.selectedDropdowns.debugDropdownContainers;
    },
  }
}
</script>