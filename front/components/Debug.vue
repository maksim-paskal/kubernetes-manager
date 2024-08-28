<template>
  <div>
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>
    <b-spinner v-if="callIsLoading" variant="primary" />
    <div v-else>
      <WarningDisableHPA />

      <DropDown id="debugDropdownContainers" :value="selectedContainer" style="margin-bottom: 10px"
        ref="debugDropdownContainers" default="" text="Select POD"
        :endpoint="`/api/${this.$route.params.environmentID}/containers?filter=kubernetes-manager/debug-containers&annotation=kubernetes-manager/debug-containers`" />

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

        <b-form-textarea spellcheck="false" rows="10" style="width: 100%" v-model="data.PhpFpmSettings"
          placeholder="Enter something..." />
        <div class="mt-3">
          <b-button variant="success" @click="enableXdebug()">Enable XDEBUG</b-button>
          <b-button @click="savePhpConfig()">Set settings in php-fpm.conf</b-button>
          <b-button variant="danger" @click="deletePod()">Delete Pod</b-button>
          <b-dropdown id="dropdown-1" offset="25" text="Templates" class="m-2">
            <b-dropdown-item v-for="(item, index) in this.config.DebugTemplates" :key="index"
              @click="templateAction(item.Data)">{{ item.Display }}</b-dropdown-item>
          </b-dropdown>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Debug', true)
    }
  },
  mounted() {
    this.$nuxt.$on("component::DropDown::selected", (data) => {
      if (data.id === "debugDropdownContainers") {
        this.$fetch();
      }
    });

    this.$refs.debugDropdownContainers.select(this.selectedContainer);
  },
  async fetch() {
    if (!this.selectedContainer) return;
    location.hash = this.selectedContainer;

    const result = await fetch(
      `/api/${this.$route.params.environmentID}/debug-info?container=${this.selectedContainer}`
    );
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result;
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  data() {
    return {
      data: {},
    };
  },
  methods: {
    async templateAction(data) {
      this.data.PhpFpmSettings += `${data}\n`;
    },
    async deletePod() {
      await this.call("make-delete-container", { Container: this.selectedContainer });

      if (!this.errorText) {
        this.$refs.debugDropdownContainers.select("");
      }
    },
    async enableXdebug() {
      await this.call("make-debug-xdebug-init", {
        Container: this.selectedContainer,
      });

      if (!this.errorText) {
        this.$fetch();
      }
    },
    async savePhpConfig() {
      await this.call("make-debug-save-config", {
        Container: this.selectedContainer,
        PhpFpmSettings: this.data.PhpFpmSettings,
      });

      if (!this.errorText) {
        this.$fetch();
      }
    },
  },
  computed: {
    debug_enabled() {
      if (this.data.XdebugEnabled == "true") {
        return "ENABLED";
      } else {
        return "DISABLED";
      }
    },
    selectedContainer() {
      return this.$store.state.selectedDropdowns.debugDropdownContainers;
    },
  },
};
</script>