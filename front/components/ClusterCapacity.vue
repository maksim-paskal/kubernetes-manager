<template>
  <div>
    <b-spinner v-if="$fetchState.pending" variant="primary" />
    <div v-else-if="data">
      <div style="cursor: pointer" @click="details()"><i class="bi bi-check-lg text-success"
          style="font-size: 30px;" />Cluster {{ cluster }} have enough resources</div>
      <ul v-if="showDetails">
        <li>Nodes: {{ data.TotalNodes }}</li>
        <li>CPU: {{ data.NodesCPU }}</li>
        <li>Memory: {{ data.NodesMemory }}</li>
        <li>Volumes: {{ data.StorageDisks }}</li>
        <li>Volumes storage: {{ data.StorageSize }}</li>
        <li>Max count of pods: {{ data.NodesMaxPods }}</li>
      </ul>
    </div>
  </div>
</template>
<script>
export default {
  props: ["cluster"],
  watch: {
    cluster() {
      this.$fetch();
    }
  },
  data() {
    return {
      data: {},
      showDetails: false
    };
  },
  async fetch() {
    this.data = {}

    const result = await fetch(`/api/cluster-info?cluster=${this.cluster}`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result
    }
  },
  methods: {
    details() {
      this.showDetails = !this.showDetails
    }
  }
}
</script>