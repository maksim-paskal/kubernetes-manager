<template>
  <div class="detail-tab">
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <b-spinner v-else-if="$fetchState.pending" variant="primary" />
    <div v-else>
      <b-button target="_blank" :href="data.Result?.IssuesExternalLink">Open Issues</b-button>
      <b-form-input style="margin-top:10px" v-model="dataFilter" autocomplete="off" placeholder="Type to Search" />
      <div v-if="data.Result.IssuesProjects" style="margin-top: 5px">
        <span style="cursor: pointer;" class="badge rounded-pill bg-primary" @click="displayOnlyItem(null)">all</span>
        <span v-for="(item, index) in data.Result.IssuesProjects" :key="index" style="margin-left: 5px;cursor: pointer;"
          class="badge rounded-pill bg-primary" @click="displayOnlyItem(item)">{{ item }}</span>
      </div>
      <b-table style="margin-top:5px" striped hover :items="data.Result?.Items" :fields="dataFields"
        :filter="dataFilter">
        <template v-slot:cell(Issues)="row">
          <span :class="getLevelClass(row.item.Level)" />
          <span v-if="row.item.IsNew" class="badge-margin badge rounded-pill bg-primary">New</span>
          (<a class="text-decoration-none" target="_blank" :href="row.item.ProjectURL">{{ row.item.Project
            }}</a>)&nbsp;<a class="text-decoration-none" target="_blank" :href="row.item.Link">{{
              row.item.Title }}</a>&nbsp;{{ row.item.Culprit }}
          <div>{{ row.item.Description }}</div>
          <div :title="row.item.LastSeen">{{ row.item.LastSeenShort }}&nbsp;ago</div>
        </template>
      </b-table>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Issues', true)
    }
  },
  data() {
    return {
      data: {},
      dataFilter: null,
      dataFields: [
        { key: "Issues" },
      ],
    };
  },
  methods: {
    displayOnlyItem(item) {
      this.dataFilter = item ? "/" + item + "/": "";
    },
    getLevelClass(level) {
      if (level === "error" || level === "fatal") {
        return "bi bi-circle-fill text-danger";
      }
      if (level === "warning") {
        return "bi bi-circle-fill text-warning";
      }
      return "bi bi-circle-fill text-primary";
    },
  },
  async fetch() {
    const result = await fetch(`/api/${this.$route.params.environmentID}/issues`);
    if (result.ok) {
      this.data = await result.json();
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
}
</script>