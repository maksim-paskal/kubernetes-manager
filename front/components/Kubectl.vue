<template>
  <div>
    <h2>Install</h2>
    Install requred application
    <a target="_blank"
      href="https://kubernetes.io/docs/tasks/tools/install-kubectl/">https://kubernetes.io/docs/tasks/tools/install-kubectl/</a>

    <h2 style="margin-top:30px">Configure</h2>
    <KubectlLink />

    <h2 style="margin-top:30px">Test</h2>
    <CopyTextbox :text="commandTest" />

    <h2 style="margin-top:30px">kubectl Cheat Sheet</h2>
    <a href="https://kubernetes.io/docs/reference/kubectl/cheatsheet/"
      target="_blank">https://kubernetes.io/docs/reference/kubectl/cheatsheet/</a>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Kubectl', true)
    }
  },
  computed: {
    commandTest() {
      return `kubectl --kubeconfig=${this.kubeconfig} -n ${this.environment.Namespace} get pods`;
    },
    kubeconfig() {
      return `/tmp/kubeconfig-${this.environment.Cluster}-${this.environment.Namespace}`;
    },
  },
}
</script>
