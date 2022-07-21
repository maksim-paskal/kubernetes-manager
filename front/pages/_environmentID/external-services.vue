<template>
  <div class="detail-tab">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <b-spinner v-if="callIsLoading" variant="primary" />
    <div v-else>
      <GitlabProjects v-if="user.user && config" ref="externalServicesProjects" podInfo="true" />
      <b-spinner v-else size="sm" variant="primary" />

      <div v-if="GitlabProjectsLoaded">
        <b-button @click="deploySelected()">Deploy selected</b-button>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  computed: {
    GitlabProjectsLoaded() {
      return this.$store.state.componentLoaded.GitlabProjects;
    }
  },
  methods: {
    deploySelected() {
      const services = this.$refs.externalServicesProjects.getSelectedServices();

      this.call('make-deploy-services', {
        Services: services
      });
    }
  }
}
</script>