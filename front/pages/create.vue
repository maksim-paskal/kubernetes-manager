<template>
  <div style="padding: 10px">
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <b-spinner v-if="callIsLoading || $fetchState.pending" variant="primary" />
    <div v-else>
      <div style="display:flex;align-items:center;margin-bottom: 10px;">
        Project profile:&nbsp;
        <b-form-select class="form-select" v-model="projectProfile" :options="projectProfiles"
          :disabled="!GitlabProjectsLoaded" style="width:300px" />
        &nbsp;&nbsp;Cluster:&nbsp;
        <select id="createClusterNameId" class="form-select" :disabled="!GitlabProjectsLoaded" style="width:300px">
          <option :key="index" v-for="(item, index) in this.config.Clusters">{{ item.ClusterName }}</option>
        </select>
      </div>

      <div v-if="user.user && config && projectProfile">
        <GitlabProjects ref="createNewEnvironmentProjects" :projectProfile="projectProfile" />

        <div v-if="GitlabProjectsLoaded">
          <b-button style="margin-left: 20px;margin-right:30px" size="lg" @click="createEnvironment()">create new
            environment</b-button>
        </div>
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
  async fetch() {
    if (this.projectProfileLoaded) {
      return;
    }
    const result = await fetch(`/api/project-profiles`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result;

      this.data.forEach(async (el) => {
        this.projectProfiles.push({
          text: el.Name,
          value: el.Value
        });
      })

      this.projectProfile = this.data[0].Value;
      this.projectProfileLoaded = true;
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  data() {
    return {
      data: [],
      projectProfileLoaded: false,
      projectProfile: "",
      projectProfiles: [],
    }
  },
  methods: {
    async createEnvironment() {
      const cluster = document.getElementById("createClusterNameId").value;

      const services = this.$refs.createNewEnvironmentProjects.getSelectedServices();

      await this.callEndpoint('/api/make-create-environment', {
        Profile: this.projectProfile,
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