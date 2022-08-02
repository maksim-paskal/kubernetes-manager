<template>
  <div>
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-alert v-if="this.errorText" variant="danger" show>{{
        this.errorText
    }}</b-alert>
    <b-spinner v-else-if="$fetchState.pending" variant="primary" />
    <b-spinner v-else-if="callIsLoading" variant="primary" />
    <div v-else>
      <div v-if="RunningPodsCount == 0">
        Paused
        <br />
        <b-button variant="success" @click="start()">Start</b-button>
      </div>
      <div>{{ namespaceStatus }}</div>
    </div>
  </div>
</template>
<script>
export default {
  props: ["item"],
  data() {
    return {
      RunningPodsCount: -1
    }
  },
  async fetch() {
    const result = await fetch(`/api/${this.item.ID}/pods`);
    if (result.ok) {
      const data = await result.json();
      this.RunningPodsCount = data.Result.PodsTotal
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  methods: {
    async start() {
      await this.callEndpoint(`/api/${this.item.ID}/make-start`)
      if (this.infoText) {
        this.$fetch()
      }
    }
  },
  computed: {
    namespaceStatus() {
      const total = this.RunningPodsCount;
      let required = 0
      if (this.item.NamespaceAnnotations && this.item.NamespaceAnnotations[this.const().LabelRequiredPods]) {
        required = this.item.NamespaceAnnotations[this.const().LabelRequiredPods]
      }

      if (total == 0) {
        return;
      }

      if (!required || total >= required) {
        return "Ready";
      }

      return "Waiting";
    },
  }
}
</script>