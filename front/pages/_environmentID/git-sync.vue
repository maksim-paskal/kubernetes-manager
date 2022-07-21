<template>
  <div class="detail-tab">
    <b-alert variant="warning" show>
      <b-button @click="makeAPICall('disableHPA', 'none')">Disable autoscaling</b-button>&nbsp;For proper usage
      you must disable autoscaler
    </b-alert>
    <DropDown id="gitSyncDropdownContainers" style="margin-bottom:10px" ref="gitSyncDropdownContainers"
      default="backend" text="Select POD" :endpoint="`/api/${this.$route.params.environmentID}/containers`" />

    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-spinner v-if="$fetchState.pending" variant="primary" />
    <div v-else-if="selectedContainer">
      <b-card-text>
        Sync is
        <strong>AAA</strong>
      </b-card-text>

      <b-container fluid>
        <b-row class="my-1">
          <b-col sm="3">
            <label for="input-gitOrigin">origin:</label>
          </b-col>
          <b-col sm="9">
            <b-form-input id="input-gitOrigin"></b-form-input>
          </b-col>
        </b-row>
        <b-row class="my-2">
          <b-col sm="3">
            <label for="input-gitBranch">branch:</label>
          </b-col>
          <b-col sm="9">
            <b-form-input id="input-gitBranch"></b-form-input>
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <b-form-textarea rows="5" readonly />
          </b-col>
        </b-row>
      </b-container>
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
      data: {}
    }
  },
  methods: {
    refresh() {
      this.$refs.gitSyncDropdownContainers.isLoaded = false;
      this.$refs.gitSyncDropdownContainers.select("");
    }
  },
  computed: {
    selectedContainer() {
      return this.$store.state.selectedDropdowns.gitSyncDropdownContainers;
    },
  }
}
</script>