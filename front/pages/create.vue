<template>
  <div style="padding: 10px">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <b-spinner v-if="callIsLoading" variant="primary" />
    <div v-else>
      <GitlabProjects v-if="user.user && config" ref="createNewEnvironmentProjects" />
      <b-spinner v-else size="sm" variant="primary" />

      <div v-if="GitlabProjectsLoaded" style="display:flex;align-items:center">
        <b-button style="margin-left: 20px;margin-right:30px" size="lg" @click="createEnvironment()">create new
          environment</b-button>
        <div class="align-items:center">Cluster:&nbsp;</div>
        <select id="createClusterNameId" class="form-select" style="width:300px">
          <option :key="index" v-for="(item, index) in this.config.Clusters">{{ item.ClusterName }}</option>
        </select>
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
      const cluster = document.getElementById("createClusterNameId").value;

      const services = this.$refs.createNewEnvironmentProjects.getSelectedServices();

      await this.callEndpoint('/api/make-create-environment', {
        Services: services,
        User: this.user.user,
        Cluster: cluster,
      }, true);

      if (this.infoText) {
        this.$router.push(`/${this.infoText}/external-services`)
      }
    }
  }
}
</script>