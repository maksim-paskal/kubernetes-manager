<template>
  <b-alert v-if="$fetchState.error" variant="danger" show>{{
    $fetchState.error.message
  }}</b-alert>
  <div v-else-if="$fetchState.pending" style="padding: 50px" class="text-center">
    <b-spinner style="width: 10rem; height: 10rem" variant="primary" />
  </div>
  <div v-else>
    <div style="padding:10px">
      <b-form-input v-model="tableFilter" autocomplete="off" placeholder="Type to Search" />
    </div>
    <b-table striped hover :fields="fields" :items="data" :filter="tableFilter">
      <template v-slot:cell(GitBranch)="data">
        <GitBranch :item="data.item" />
      </template>
      <template v-slot:cell(Status)="data">
        <EnvironmentStatus :item="data.item" />
      </template>
      <template v-slot:cell(Hosts)="data">
        <ul>
          <li v-bind:key="index" v-for="(item, index) in data.item.Hosts">
            <a class="text-decoration-none" :href="item" rel="noopener" target="_blank">{{ item }}</a>
          </li>
        </ul>
      </template>
      <template v-slot:cell(Name)="row">
        <b-link class="text-decoration-none" :to="`/${row.item.ID}/info`">{{ getEnvironmentName(row.item) }}</b-link>
        <EnvironmentBadges :badges="row.item.NamespaceBadges" />
      </template>
    </b-table>
  </div>
</template>
<script>
export default {
  props: ["filter"],
  async fetch() {

    let url = `/api/environments`;

    if (this.filter) {
      url += `?filter=` + encodeURIComponent(this.filter);
    }

    const result = await fetch(url);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  data() {
    return {
      tableFilter: "",
      fields: [
        { key: "Name", sortable: false, class: "text-center" },
        { key: "Status", sortable: false, class: "text-center" },
        { key: "GitBranch", sortable: false },
        { key: "Hosts", sortable: false },
      ],
      data: []
    }
  },
  methods: {
    getEnvironmentName(data) {
      if (data.NamespaceAnnotations && data.NamespaceAnnotations[this.const().LabelEnvironmentName]) {
        return data.NamespaceAnnotations[this.const().LabelEnvironmentName]
      }

      return data.Namespace
    }
  }
}
</script>
