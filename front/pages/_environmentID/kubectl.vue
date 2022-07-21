<template>
  <div class="detail-tab">
    <h2>Install</h2>
    Install requred application
    <a target="_blank"
      href="https://kubernetes.io/docs/tasks/tools/install-kubectl/">https://kubernetes.io/docs/tasks/tools/install-kubectl/</a>

    <h2 style="margin-top:30px">Configure</h2>
    "Save As" this
    <a target="_blank" :href="`/api/${this.environment.ID}/kubeconfig`">file</a>
    to {{ this.kubeconfig }}&nbsp;(token will be expired after 10h)

    <h2 style="margin-top:30px">Test</h2>
    <b-form-textarea readonly v-model="commandTest"
      style="background-color:#eeeeee;border:0px;padding:10px;outline:none;width:100%" />
  </div>
</template>
<script>
export default {
  layout: "details",
  computed: {
    commandTest() {
      return `kubectl --kubeconfig=${this.kubeconfig} -n ${this.environment.Namespace} get pods`
    },
    kubeconfig() {
      return `/tmp/kubeconfig-${this.environment.Cluster}-${this.environment.Namespace}`
    },
  }
}
</script>
