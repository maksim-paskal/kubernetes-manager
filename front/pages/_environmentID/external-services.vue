<template>
  <div class="detail-tab">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <b-spinner v-if="callIsLoading" variant="primary" />
    <div v-else>
      <GitlabProjects v-if="user.user && config" ref="externalServicesProjects" podInfo="true"
        :namespace="environment.Namespace" />
      <b-spinner v-else size="sm" variant="primary" />

      <div v-if="GitlabProjectsLoaded">
        <b-button title="full cycle of deploy process, build images and than deploy" size="lg"
          @click="buildDeploySelected()">Build and Deploy selected</b-button>
        <b-button title="reverts all changes on branch" style="margin-left:30px" @click="deploySelected()">Only Deploy
          selected (without build)</b-button>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('External services', true)
    }
  },
  computed: {
    GitlabProjectsLoaded() {
      return this.$store.state.componentLoaded.GitlabProjects;
    }
  },
  methods: {
    buildDeploySelected() {
      const services = this.$refs.externalServicesProjects.getSelectedServices();

      this.call('make-deploy-services', {
        Services: services,
        Operation: "BUILD",
      });
    },
    deploySelected() {
      const services = this.$refs.externalServicesProjects.getSelectedServices();

      this.call('make-deploy-services', {
        Services: services,
        Operation: "DEPLOY",
      });
    }
  }
}
</script>