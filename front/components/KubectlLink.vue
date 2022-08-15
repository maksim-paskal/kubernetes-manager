<template>
  <div>
    "Save As" this
    <a target="_blank" :href="`/api/${this.environment.ID}/kubeconfig`">file</a>
    to {{ this.kubeconfig }}&nbsp;(token will be expired after 10h)
  </div>
</template>
<script>
export default {
  computed: {
    kubeconfig() {
      return `/tmp/kubeconfig-${this.environment.Cluster}-${this.environment.Namespace}`;
    },
  },
}
</script>