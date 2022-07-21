<template>
  <div style="padding: 10px">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <b-spinner v-if="callIsLoading" variant="primary" />
    <div v-else>
      <GitlabProjects v-if="user.user && config" ref="createNewEnvironmentProjects" />
      <b-spinner v-else size="sm" variant="primary" />

      <div v-if="GitlabProjectsLoaded">
        <b-button size="lg" @click="createEnvironment()">create new environment</b-button>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  layout: "create",
  computed: {
    GitlabProjectsLoaded() {
      return this.$store.state.componentLoaded.GitlabProjects;
    }
  },
  methods: {
    async createEnvironment() {
      const services = this.$refs.createNewEnvironmentProjects.getSelectedServices();

      await this.callEndpoint('/api/make-create-environment', {
        Services: services,
        User: this.user.user,
      }, true);

      if (this.infoText) {
        this.$router.push(`/${this.infoText}/external-services`)
      }
    }
  }
}
</script>