<template>
  <div class="detail-tab">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <b-spinner v-else-if="$fetchState.pending || callIsLoading" variant="primary" />
    <div v-else>
      <div style="margin-bottom: 15px" v-if="this.user.user">
        <b-button style="margin-right: 10px" v-bind:key="index" v-for="(item, index) in this.data.Result.Actions"
          target="_blank" @click="startAutotest(item)">&nbsp;{{ item.Name
          }}</b-button>
        <b-button @click="showCustomDialog()">Custom autotest</b-button>
      </div>
      <b-card-group v-if="this.data.Result.LastPipelines" style="margin-bottom:15px">
        <b-card :title="getCardTitle(item)" v-bind:key="index" v-for="(item, index) in this.data.Result.LastPipelines">
          <b-card-text>
            <AllureScore v-if="item.Status === 'success'" :allureResults="item.ResultURL" variant="large"
              showFailedTests="true" />
            <div v-else>In progress</div><br />
            <b-button style="margin-top:5px" size="sm" v-if="item.Status === 'success'" variant="outline-primary"
              target="_blank" :href="item.ResultURL">Open Report</b-button>
            <b-button style="margin-top:5px" size="sm" v-else variant="outline-primary" target="_blank"
              :href="item.PipelineURL">Open Pipeline</b-button>
          </b-card-text>
          <template #header>
            <small class="text-muted">&nbsp;{{ item.PipelineRelease }}</small>
          </template>
          <template #footer>
            <small class="text-muted" :title="item.PipelineCreated">{{ item.PipelineCreatedHuman }}&nbsp;ago</small>
          </template>
        </b-card>
      </b-card-group>
      <b-form-input v-model="dataFilter" autocomplete="off" placeholder="Type to Search" />
      <b-table style="margin-top:5px" striped hover :items="data.Result.Pipelines" :fields="dataFields"
        :filter="dataFilter">
        <template v-slot:cell(Score)="row">
          <AllureScore v-if="row.item.Status === 'success'" :allureResults="row.item.ResultURL" />
        </template>
        <template v-slot:cell(Test)="row">
          {{ row.item.Test }}
          <b-button v-if="row.item.PipelineEnv?.CUSTOM_ACTION" @click="showJSON(row.item.PipelineEnv)"
            class="badge rounded-pill bg-primary" :title="JSON.stringify(row.item.PipelineEnv)">custom</b-button>
        </template>
        <template v-slot:cell(PipelineCreated)="row">
          <div :title="row.item.PipelineCreated">{{ row.item.PipelineCreatedHuman }}&nbsp;ago</div>
        </template>
        <template v-slot:cell(Actions)="row">
          <b-button size="sm" variant="outline-primary" target="_blank" :href="row.item.PipelineURL">Open
            Pipeline</b-button>&nbsp;
          <b-button size="sm" v-if="row.item.Status === 'success'" variant="outline-primary" target="_blank"
            :href="row.item.ResultURL">Open Report</b-button>
        </template>
      </b-table>
      <b-button variant="light" title="Get more results" v-if="data.Result.HasMorePipelines"
        @click="getMoreResults()"><em class="bi bi-arrow-clockwise" /></b-button>
    </div>
    <b-modal centered id="bv-custom-dialog" @ok="createCustomRun()" title="Create custom run">
      <AutotestCustomAction ref="autotestCustomAction" v-if="this.data.Result?.CustomAction"
        :customAction="this.data.Result.CustomAction" />
    </b-modal>
    <b-modal id="bv-show-json-dialog" title="Custom report information">
      <pre>{{ jsonObject }}</pre>
    </b-modal>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Autotests', true)
    }
  },
  data() {
    return {
      data: {},
      size: 10,
      jsonObject: {},
      dataFilter: null,
      dataFields: [
        { key: "PipelineID", label: "ID" },
        { key: "PipelineCreated", label: "Created" },
        { key: "PipelineDuration", label: "Duration" },
        { key: "Status", label: "Status" },
        { key: "Test", label: "Test" },
        { key: "PipelineRelease", label: "Release" },
        { key: "Score", label: "Score" },
        { key: "PipelineOwner", label: "Owner" },
        { key: "Actions", label: "Actions" },
      ],
    }
  },
  async fetch() {
    const result = await fetch(`/api/${this.$route.params.environmentID}/autotests?size=${this.size}`);
    if (result.ok) {
      this.data = await result.json();
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  methods: {
    startAutotest(item) {
      this.call('make-start-autotest', { Test: item.Test })
    },
    getCardTitle(item) {
      return `${item.Test}`
    },
    getMoreResults() {
      this.size += 10;
      this.$router.app.refresh();
    },
    showCustomDialog() {
      this.$bvModal.show('bv-custom-dialog')
    },
    showJSON(obj) {
      this.jsonObject = obj
      this.$bvModal.show('bv-show-json-dialog')
    },
    createCustomRun() {
      this.call('make-start-autotest-custom', this.$refs.autotestCustomAction.getCustomActionInput())
    }
  }
}
</script>