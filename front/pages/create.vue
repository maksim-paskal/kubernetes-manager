<template>
  <div style="padding: 10px">
    <b-alert v-if="namespace" variant="success" show>Namespace <strong>{{ namespace.split(":")[1] }}</strong> created
    </b-alert>
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <div style="padding: 50px" class="text-center" v-if="isBusy">
      <b-spinner style="width: 10rem; height: 10rem" variant="primary"></b-spinner>
    </div>
    <div v-if="!isBusy">
      <b-table striped hover :items="externalServicesItems" :fields="externalServicesFields">
        <template #cell(Service)="data">
          <em v-if="checkRequired(data.item.ProjectID)" class="text-danger bi bi-asterisk" />
          <a target="_blank" :href="data.item.WebURL" style="text-decoration: none">{{ data.item.Description
          }}</a>&nbsp;<span v-if="data.item.Deploy" title="deploy tag" class="badge rounded-pill bg-primary">{{
    data.item.Deploy
}}</span>
        </template>
        <template #cell(Deploy)="data">
          <b-button :disabled="namespace != null" style="width: 100px" @click="getProjectRefs(data)"
            :variant="deployButtonVariant(data)">
            <em v-if="data.item.Deploy" class="bi bi-plus-lg" />
            <em v-else class="bi bi-dash-lg" />
          </b-button>
        </template>
        <template v-if="namespace" #cell(Status)="data">
          <div style="height: 25px">
            <b-spinner v-if="!data.item.AdditionalInfo" variant="primary" />
            <em v-if="
              data.item.AdditionalInfo &&
              data.item.AdditionalInfo.Pipelines.LastSuccessPipeline
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
      </b-table>
      <hr />
      <div v-if="!namespace" style="margin-top: 20px">
        <b-button variant="light" to="/">Cancel</b-button>
        <b-button variant="light" style="margin-left: 20px" @click="initPage()">Reset</b-button>
        <b-button :disabled="!this.externalServicesSelected" variant="primary" style="margin-left: 20px" class="btn-lg"
          @click="createNewBranch()">
          Create new branch</b-button>
        <b-dropdown id="dropdown-2" style="margin-left: 20px" offset="25" text="Templates">
          <b-dropdown-item v-for="(item, index) in this.config
          ? this.config.ExternalServicesTemplates
          : []" :key="index" @click="externalServiceTemplate(item.Data)">{{ item.Display }}</b-dropdown-item>
        </b-dropdown>
      </div>
      <div v-if="namespace" style="margin-top: 20px">
        <b-button variant="light" to="/">Cancel</b-button>
        <b-button variant="secondary" style="margin-left: 20px" class="btn-lg" @click="initPage()">Refresh</b-button>
        <b-button variant="primary" style="margin-left: 20px" :to="`/#${namespace.split(':')[1]}`">Goto list</b-button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  mounted() {
    this.initPage();
  },
  data() {
    return {
      config: null,
      isBusy: true,
      namespace: null, //"hcloud-dev:aaaaa",
      errorText: null,
      externalServicesSelected: false,
      externalServicesItems: null,
      externalServicesFields: [
        { key: "Status", label: "Status", class: "text-center" },
        { key: "Deploy", label: "Deploy", class: "text-center" },
        { key: "Service", label: "Service", tdClass: "col-lg-10" },
      ],
    };
  },
  methods: {
    async APICall(url) {
      this.isBusy = true;
      this.errorText = null;

      try {
        return await this.$axios.$get(url);
      } catch (e) {
        if (e.response && e.response.data) {
          this.errorText = e.response.data;
        } else {
          this.errorText = e;
        }
      } finally {
        this.isBusy = false;
      }
    },
    async initPage() {
      this.config = await this.APICall("/api/getFrontConfig");

      this.externalServicesSelected = false;
      const externalServicesItems = await this.APICall("/api/getProjects");
      this.externalServicesItems = externalServicesItems.result;

      if (this.namespace) {
        // wait some time to created pipelines
        await new Promise(resolve => setTimeout(resolve, 2000));

        this.externalServicesItems.forEach(async (el) => {
          const data = await this.$axios.$get(
            `/api/getProjectInfo?podInfo=false&projectID=${el.ProjectID}&namespace=${this.namespace}`
          );
          if (data.result) {
            el.AdditionalInfo = data.result;
          }
        });
      }
    },
    deployButtonVariant(data) {
      if (data.item.Deploy) {
        return "success";
      }
      return "outline-secondary";
    },
    getProjectRefs(data) {
      const h = this.$createElement;
      const messageVNode = h("div", {
        domProps: {
          innerHTML:
            "Select branch: <select class='form-select' id='gitlabProjectBranchId' disabled><option value=false>Loading...</option></select>",
        },
      });

      this.$axios
        .$get(`/api/getProjectRefs?projectID=${data.item.ProjectID}`)
        .then((httpData) => {
          const select = document.getElementById("gitlabProjectBranchId");
          select.innerHTML = "";
          for (let row of httpData.result) {
            var opt = document.createElement("option");
            opt.value = row.Name;
            opt.innerHTML = row.Name;
            select.appendChild(opt);
          }
          select.disabled = false;
          select.value = data.item.DefaultBranch;
        });

      this.$bvModal
        .msgBoxConfirm([messageVNode], {
          title: data.item.Name,
          centered: true,
        })
        .then((value) => {
          if (!value) return false;

          const v = document.getElementById("gitlabProjectBranchId").value;
          if (!v) return false;

          for (let row of this.externalServicesItems) {
            if (data.item.ProjectID == row.ProjectID) {
              row.Deploy = v;
              this.externalServicesSelected = true;
              break;
            }
          }
        });
    },
    checkRequired(projectID) {
      if (!this.config.ProjectTemplates) {
        return false;
      }

      for (let row of this.config.ProjectTemplates) {
        if (row.ProjectID == projectID) {
          return row.Required;
        }
      }
      return false;
    },
    externalServiceTemplate(projectIDs) {
      const ids = projectIDs.split(";");

      for (let row of this.externalServicesItems) {
        if (ids.includes(row.ProjectID.toString())) {
          row.Deploy = row.DefaultBranch;
          this.externalServicesSelected = true;
        } else {
          row.Deploy = false;
        }
      }
    },
    async createNewBranch() {
      let selectedServices = [];
      for (let row of this.externalServicesItems) {
        if (row.Deploy) {
          selectedServices[
            selectedServices.length
          ] = `${row.ProjectID}:${row.Deploy}`;
        }
      }

      const query = `services=${encodeURIComponent(
        selectedServices.join(";")
      )}`;

      let namespace = await this.APICall(`/api/createNewBranch?${query}`);
      if (namespace) {
        this.namespace = namespace;
        this.initPage();
      }
    },
  },
};
</script>