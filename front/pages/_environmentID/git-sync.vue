<template>
  <div class="detail-tab">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>
    <b-spinner v-if="callIsLoading" variant="primary" />
    <div v-else>
      <WarningDisableHPA />
      <DropDown id="gitSyncDropdownContainers" :value="selectedContainer" style="margin-bottom:10px"
        ref="gitSyncDropdownContainers" text="Select POD"
        :endpoint="`/api/${this.$route.params.environmentID}/containers?filter=kubernetes-manager/debug-containers&annotation=kubernetes-manager/debug-containers`" />

      <b-alert v-if="$fetchState.error" variant="danger" show>{{
          $fetchState.error.message
      }}</b-alert>
      <b-spinner v-if="$fetchState.pending" variant="primary" />
      <div v-else-if="selectedContainer">
        <b-card-text>
          Sync is
          <strong>{{ this.gitSync_enabled }}</strong>
        </b-card-text>

        <b-container fluid>
          <b-row class="my-1">
            <b-col sm="3">
              <label for="input-gitOrigin">origin:</label>
            </b-col>
            <b-col sm="9">
              <b-form-input v-model="data.GitOrigin" id="input-gitOrigin"></b-form-input>
            </b-col>
          </b-row>
          <b-row class="my-2">
            <b-col sm="3">
              <label for="input-gitBranch">branch:</label>
            </b-col>
            <b-col sm="9">
              <b-form-input v-model="data.GitBranch" id="input-gitBranch"></b-form-input>
            </b-col>
          </b-row>
          <b-row v-if="this.isShowPublicKey">
            <b-col>
              <b-form-textarea rows="5" readonly :value="data.PublicKey" />
            </b-col>
          </b-row>
        </b-container>
        <div class="mt-3">
          <b-button variant="success" @click="gitSyncInit()">Init</b-button>
          <b-button @click="showPublicKey()">{{ isShowPublicKey ? "Hide" : "Show" }} public key</b-button>
          <b-button variant="danger" @click="deletePod()">Delete pod</b-button>
          <b-button @click="fetchGit()">Fetch</b-button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  mounted() {
    this.$nuxt.$on('component::DropDown::selected', (data) => {
      if (data.id === 'gitSyncDropdownContainers') {
        this.$fetch()
      }
    })

    this.$refs.gitSyncDropdownContainers.select(this.selectedContainer);
  },
  async fetch() {
    if (!this.selectedContainer) return;
    location.hash = this.selectedContainer;

    const result = await fetch(`/api/${this.$route.params.environmentID}/git-sync?container=${this.selectedContainer}`);
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
      isShowPublicKey: false,
      data: {}
    }
  },
  methods: {
    showPublicKey() {
      this.isShowPublicKey = !this.isShowPublicKey
    },
    async gitSyncInit() {
      await this.call('make-git-sync-init', {
        Container: this.selectedContainer,
        GitOrigin: this.data.GitOrigin,
        GitBranch: this.data.GitBranch,
      })

      if (!this.errorText) {
        this.$fetch()
      }
    },
    async deletePod() {
      await this.call('make-delete-pod', { Container: this.selectedContainer })

      if (!this.errorText) {
        this.$refs.gitSyncDropdownContainers.select("");
      }
    },
    async fetchGit() {
      await this.call('make-git-sync-fetch', { Container: this.selectedContainer })

      if (!this.errorText) {
        this.$fetch()
      }
    },
  },
  computed: {
    gitSync_enabled() {
      if (this.data.GitSyncEnabled == "true") {
        return "ENABLED"
      } else {
        return "DISABLED"
      }
    },
    selectedContainer() {
      return this.$store.state.selectedDropdowns.gitSyncDropdownContainers;
    },
  }
}
</script>