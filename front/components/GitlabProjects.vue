<template>
  <div>
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-alert v-if="this.errorText" variant="danger" show>{{
        this.errorText
    }}</b-alert>
    <b-alert v-if="this.infoText" variant="info" show>{{
        this.infoText
    }}</b-alert>
    <b-spinner v-if="$fetchState.pending" variant="primary" />
    <div v-else>
      <div v-if="projectProfile">
        <b-button variant="outline-primary" @click="selectAllFromMain()">Select all services from main</b-button>
        <b-button variant="outline-primary" @click="clearAllSelection()">Clear selection</b-button>
      </div>
      <b-table striped hover :items="data" :fields="tableFields">
        <template #cell(Service)="data">
          <b-button title="delete service from namespace" v-if="podInfo" :disabled="data.item.GitBranch ? false : true"
            size="sm" @click="
              call('make-delete-service', {
                ProjectID: `${data.item.ProjectID}`,
                Ref: data.item.GitBranch,
              })
            " variant="outline-danger"><em class="bi bi-trash3" /></b-button>&nbsp;
          <em v-if="data.item.Required" class="text-danger bi bi-asterisk" />
          <a target="_blank" :href="data.item.WebURL" style="text-decoration: none">{{ data.item.Description }}</a>
          &nbsp;<span v-if="data.item.GitBranch" title="git tag" class="badge rounded-pill bg-primary">{{
              data.item.GitBranch
          }}</span>
          <span v-else-if="data.item.AdditionalInfo" title="docker tag" class="badge rounded-pill bg-primary">{{
              data.item.AdditionalInfo.PodRunning.Tag
          }}</span>&nbsp;<a v-if="data.item.AdditionalInfo && data.item.AdditionalInfo.PodRunning.GitHash"
            target="_blank" :href="getGitlabCommitURL(data.item)"><span title="git short commit hash"
              class="badge rounded-pill bg-success">{{ data.item.AdditionalInfo.PodRunning.GitHash }}</span></a>
          <a v-if="
            data.item.AdditionalInfo &&
            data.item.AdditionalInfo.Pipelines.LastSuccessPipeline
          " :href="data.item.AdditionalInfo.Pipelines.LastSuccessPipeline" target="_blank"><span
              title="deploy pipeline" class="badge rounded-pill bg-success">pipeline</span></a>
          <div v-if="data.item.TagsList.length > 0">
            <div style="margin-left: 5px" class="badge btn-secondary" v-bind:key="index"
              v-for="(item, index) in data.item.TagsList">
              {{ item }}
            </div>
          </div>
        </template>
        <template #cell(Status)="data">
          <div style="height: 25px">
            <b-spinner v-if="!data.item.AdditionalInfo" variant="primary" />
            <em v-if="
              data.item.AdditionalInfo &&
              data.item.AdditionalInfo.PodRunning.Found
            " class="bi bi-check-circle-fill" style="font-size: 26px; color: green" />
            <a target="_blank" :href="data.item.AdditionalInfo.Pipelines.LastErrorPipeline" v-if="
              data.item.AdditionalInfo &&
              data.item.AdditionalInfo.Pipelines.LastErrorPipeline
            "><em class="bi bi-exclamation-circle-fill" style="font-size: 26px; color: #dc3545" /></a>
            <a target="_blank" :href="data.item.AdditionalInfo.Pipelines.LastRunningPipeline" v-if="
              data.item.AdditionalInfo &&
              data.item.AdditionalInfo.Pipelines.LastRunningPipeline
            "><em class="bi bi-hourglass-split" style="font-size: 26px; color: #1f75cb" /></a>
          </div>
        </template>
        <template #cell(Deploy)="data">
          <DropDown :ref="`gitlabProjects${data.item.ProjectID}`" :id="`gitlabProjects${data.item.ProjectID}`"
            :default="data.item.GitBranch" text="Select branch" :value="data.item.Deploy"
            :endpoint="`/api/project-refs?id=${data.item.ProjectID}`" />
        </template>
      </b-table>
    </div>
  </div>
</template>
<script>
export default {
  props: ["podInfo", "namespace", "projectProfile"],
  watch: {
    projectProfile: function () {
      this.$fetch();
    }
  },
  mounted() {
    this.errorText = "";
    this.infoText = "";

    this.$nuxt.$on("component::DropDown::selected", (dropdownData) => {
      if (dropdownData.id.startsWith("gitlabProjects")) {
        for (let row of this.data) {
          if (
            row.ProjectID == dropdownData.id.replace(/^(gitlabProjects)/, "")
          ) {
            row.Deploy = dropdownData.selected;
            break;
          }
        }
      }
    });
  },
  data() {
    return {
      data: {},
      fieldsPodInfo: [
        {
          key: "Status",
          label: "Status",
          class: "text-center col-deploy-service-status",
        },
        {
          key: "Deploy",
          label: "Deploy",
          class: "text-center col-deploy-service-text",
        },
        { key: "Service", label: "Service", tdClass: "col-lg-9 col-valign" },
      ],
      fields: [
        {
          key: "Deploy",
          label: "Deploy",
          class: "text-center col-deploy-service-text",
        },
        { key: "Service", label: "Service", tdClass: "col-lg-9 col-valign" },
      ],
    };
  },
  computed: {
    tableFields() {
      if (this.podInfo) {
        return this.fieldsPodInfo;
      } else {
        return this.fields;
      }
    },
  },
  methods: {
    selectAllFromMain() {
      this.data.forEach(async (row) => {
        row.Deploy = row.DefaultBranch
      });
    },
    clearAllSelection() {
      this.data.forEach(async (row) => {
        row.Deploy = ""
      });
    },
    getGitBranch(projectID) {
      if (!this.environment) return;
      if (!this.environment.NamespaceAnnotations) return;

      let gitBranch = "";
      Object.keys(this.environment.NamespaceAnnotations).forEach((key) => {
        if (key == `kubernetes-manager/project-${projectID}`) {
          gitBranch = this.environment.NamespaceAnnotations[key];
        }
      });

      return gitBranch;
    },
    getSelectedServices() {
      let selectedServices = [];

      this.data.forEach(async (row) => {
        if (row.Deploy) {
          selectedServices[
            selectedServices.length
          ] = `${row.ProjectID}:${row.Deploy}`;
        }
      });

      return selectedServices.join(";");
    },
    getGitlabCommitURL(obj) {
      if (obj && obj.AdditionalInfo && obj.AdditionalInfo.PodRunning.GitHash) {
        return `${obj.WebURL}/-/tree/${obj.AdditionalInfo.PodRunning.GitHash}`;
      }
    },
  },
  async fetch() {
    this.$store.commit("setComponentLoaded", {
      id: "GitlabProjects",
      loaded: false,
    });
    let params = `profile=${this.projectProfile}`
    if (this.namespace) {
      params = `namespace=${this.namespace}`
    }
    const result = await fetch(`/api/external-services?${params}`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result;

      // set default branch for creation
      if (!this.podInfo) {
        this.data.forEach(async (el) => {
          if (el.SelectedBranch) {
            el.Deploy = el.SelectedBranch;
          }
        });
      }

      if (this.podInfo) {
        this.data.forEach(async (el) => {
          el.GitBranch = this.getGitBranch(el.ProjectID);

          const projectInfo = await fetch(
            `/api/${this.$route.params.environmentID}/project-info?projectID=${el.ProjectID}`
          );
          if (projectInfo.ok) {
            const dataProjectInfo = await projectInfo.json();
            el.AdditionalInfo = dataProjectInfo.Result;
          }
        });
      }

      this.$store.commit("setComponentLoaded", {
        id: "GitlabProjects",
        loaded: true,
      });
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
};
</script>