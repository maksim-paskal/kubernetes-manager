<template>
  <b-spinner v-if="$fetchState.pending" variant="primary" />
  <b-tabs v-else v-model="tabIndex">
    <b-tab v-bind:key="index" v-for="(container, index) in data" :title="container.Name" :active="(index == 0)">
      <PodContainerLog v-if="tabIndex == index" :pod="pod" :container="container.Name" />
    </b-tab>
    <b-tab style="padding:10px" title="kubectl">
      <KubectlLink />
      <div style="margin-top:20px">
        <CopyTextbox :text="kubectlCmd" />
      </div>
    </b-tab>
    <b-tab v-if="logsPodURL" style="padding:10px" title="external">
      <b-button target="_blank" :href="getPodNameLink()">Open</b-button>
    </b-tab>
  </b-tabs>
</template>
<script>
export default {
  props: ["pod", "logsPodURL"],
  watch: {
    pod() {
      this.$fetch();
    }
  },
  data() {
    return {
      tabIndex: 0,
      data: {},
    };
  },
  async fetch() {
    this.data = {}
    const result = await fetch(`/api/${this.$route.params.environmentID}/pod-containers?pod=${this.pod}`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result
    }
  },
  computed: {
    kubeconfig() {
      return `/tmp/kubeconfig-${this.environment.Cluster}-${this.environment.Namespace}`
    },
    kubectlCmd() {
      return `kubectl --kubeconfig=${this.kubeconfig} -n ${this.environment.Namespace} logs ${this.pod}`
    }
  },
  methods: {
    getPodNameLink() {
      let url = this.logsPodURL

      url = url.replace(
        /__PodName__/g,
        this.pod
      );

      return url
    },
  }
}
</script>