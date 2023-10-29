<template>
  <div style="padding: 10px">
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <ul id="progressbar">
      <li title="User settings" @click="stepClick(1)" v-bind:class="stepClass(1, 'active', 'bi bi-person')" />
      <li title="Services" @click="stepClick(2)" v-bind:class="stepClass(2, 'active', 'bi bi-gear')" />
      <li title="Summary" v-bind:class="stepClass(3, 'last', 'bi bi-check-lg')" />
    </ul>

    <b-spinner v-if="!this.config?.Clusters || callIsLoading || $fetchState.pending" variant="primary" />

    <div v-bind:class="(callIsLoading || $fetchState.pending) ? 'hide' : ''">
      <div v-bind:class="(currentStep == 1) ? '' : 'hide'">
        <UserQuotas ref="userQuotas" v-if="this.user.user && this.config.Links" />
      </div>
      <div v-bind:class="(currentStep == 2) ? '' : 'hide'">
        <div style="display:flex;align-items:center;margin-bottom: 10px;">
          Project profile:&nbsp;
          <b-form-select class="form-select" v-model="projectProfile" :options="projectProfiles"
            :disabled="!GitlabProjectsLoaded" style="width:300px" />
          &nbsp;&nbsp;Cluster:&nbsp;
          <b-form-select class="form-select" v-model="clusterName" :options="clusters" :disabled="!GitlabProjectsLoaded"
            style="width:300px" />
        </div>

        <div v-if="user.user && config && projectProfile">
          <GitlabProjects ref="createNewEnvironmentProjects" :projectProfile="projectProfile" />
        </div>
      </div>
      <div v-bind:class="(currentStep == 3) ? '' : 'hide'">
        <b-form-input v-model="environmentName" placeholder="Enter environment name (optional)" autocomplete="off" />
        <ul v-if="this.$refs.createNewEnvironmentProjects">
          <li v-bind:key="index"
            v-for="(item, index) in this.$refs.createNewEnvironmentProjects.getSelectedServicesRaw()">
            {{ item.Description }} ({{ item.Deploy }})
            <BranchCommitsBehind :id="item.ProjectID + item.Deploy" :projectID="item.ProjectID" :branch="item.Deploy" />
          </li>
        </ul>
        <ClusterCapacity v-if="clusterName" :cluster="clusterName" />
      </div>

      <div style="margin-top: 30px">
        <div
          v-if="(callIsLoading || $fetchState.pending) || !(this.$refs.createNewEnvironmentProjects && this.$refs.userQuotas)">
          &nbsp;</div>
        <b-button v-else-if="currentStep == 3" style="margin-left: 20px;margin-right:30px" size="lg"
          @click="createEnvironment()">create new
          environment</b-button>
        <b-button v-else @click="stepNext()">Next</b-button>
      </div>
    </div>
  </div>
</template>
<style scoped>
.hide {
  display: none;
}

.hand {
  cursor: pointer;
}

/*progressbar*/
#progressbar {
  margin-top: 20px;
  margin-bottom: 30px;
  overflow: hidden;
  color: lightgrey;
  padding: 0px;
}

#progressbar li {
  list-style-type: none;
  font-size: 12px;
  width: 33%;
  float: left;
  position: relative;
  text-align: center;
}

/*ProgressBar before any progress*/
#progressbar li:before {
  width: 50px;
  height: 50px;
  line-height: 45px;
  display: block;
  font-size: 24px;
  color: #ffffff;
  background: lightgray;
  border-radius: 50%;
  margin: 0 auto 10px auto;
  padding: 2px;
}

/*ProgressBar connectors*/
#progressbar li:after {
  content: '';
  width: 100%;
  height: 2px;
  background: lightgray;
  position: absolute;
  left: 0;
  top: 25px;
  z-index: -1;
}

/*Color number of the step and the connector before it*/
#progressbar li.active:before,
#progressbar li.active:after {
  background: #007bff;
}

#progressbar li.last:before,
#progressbar li.last:after {
  background: #28a745;
}
</style>
<script>
export default {
  layout: "create",
  head() {
    return {
      title: this.pageTitle('Create')
    }
  },
  computed: {
    GitlabProjectsLoaded() {
      return this.$store.state.componentLoaded.GitlabProjects;
    },
    clusters() {
      let result = []

      if (!this.config?.Clusters) {
        return result;
      }

      this.config.Clusters.forEach((el) => {
        result.push(el.ClusterName)
      })

      return result;
    },
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

      const urlParams = new URLSearchParams(window.location.search);
      const userSelectedProfile = urlParams.get('profile');

      this.projectProfile = userSelectedProfile || this.data[0].Value;
      this.clusterName = this.clusters[0];
      this.projectProfileLoaded = true;
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  data() {
    return {
      currentStep: 1,
      environmentName: "",
      clusterName: "",
      data: [],
      projectProfileLoaded: false,
      projectProfile: "",
      projectProfiles: [],
    }
  },
  methods: {
    stepClick(step) {
      if (this.currentStep > step) {
        this.currentStep = step;
      }
    },
    stepClass(step, name, className) {
      let activeClass = name;
      if (this.currentStep != step) {
        activeClass += " hand";
      }

      return this.currentStep >= step ? `${activeClass} ${className}` : className;
    },
    stepNext() {
      if (this.currentStep == 1) {
        if (this.$refs.userQuotas.getSelectedEnvironments().length > 0) {
          return this.$refs.userQuotas.showDeleteSelectedEnvironments()
        }
      }

      this.currentStep++;
    },

    async createEnvironment() {
      const services = this.$refs.createNewEnvironmentProjects.getSelectedServices();

      await this.callEndpoint('/api/make-create-environment', {
        Profile: this.projectProfile,
        Services: services,
        Cluster: this.clusterName,
        Name: this.environmentName,
      });

      if (this.infoText) {
        this.$router.push(`/${this.infoText}/external-services`)
      }
    }
  }
}
</script>