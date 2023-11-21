<template>
  <div>
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <b-alert v-else-if="this.errorText" variant="danger" show>{{ this.errorText }}</b-alert>

    <b-spinner v-if="$fetchState.pending || this.isLoading" variant="primary" />
    <div v-else>
      <div v-if="data.length === 0">
        <div style="text-align: center;">
          <h1>You don't have any environments yet</h1>
        </div>
      </div>
      <div v-else>
        <div style="text-align: center;">
          <h1>You have already created <strong>{{ data.length }}</strong> environment. Please delete any unused ones.</h1>
        </div>
        <b-table striped hover :fields="fields" :items="data">
          <template #head(Action)="">
            &nbsp;
          </template>
          <template v-slot:cell(GitBranch)="data">
            <GitBranch :item="data.item" />
          </template>
          <template v-slot:cell(Created)="row">
            <div v-if="row.item.NamespaceCreatedDays > 0" :title="row.item.NamespaceCreated">{{
              row.item.NamespaceCreatedDays }}&nbsp;days&nbsp;ago</div>
            <div v-else>today</div>
          </template>
          <template v-slot:cell(LastScaled)="row">
            <div v-if="row.item.NamespaceLastScaledDays > 0" :title="row.item.NamespaceLastScaled">{{
              row.item.NamespaceLastScaledDays }}&nbsp;days&nbsp;ago</div>
            <div v-else>today</div>
          </template>
          <template v-slot:cell(Action)="row">
            <input class="form-check-input delete-environments-checkbox" :value="row.item.ID" type="checkbox"
              :data-id="row.item.ID" :data-name="getEnvironmentName(row.item)"
              :checked="row.item.NamespaceLastScaledDays > 3" />
          </template>
          <template v-slot:cell(Name)="row">
            <b-link class="text-decoration-none" :to="`/${row.item.ID}/info`">{{ getEnvironmentName(row.item) }}</b-link>
          </template>
        </b-table>
      </div>
      <b-modal ref="delete-modal" title="Deleting environments">
        <template #modal-footer="{ ok, cancel }">
          <b-button size="sm" variant="danger" @click="ok(); deleteSelectedEnvironments()">
            Delete selected
          </b-button>
          <b-button size="sm" @click="cancel()">
            Cancel
          </b-button>
        </template>
        <div>
          <strong>Are you sure to delete selected environments?</strong>
          <ul>
            <li v-bind:key="index" v-for="(item, index) in selectedEnvironments">{{ item.dataset.name }}</li>
          </ul>
        </div>
      </b-modal>
    </div>
  </div>
</template>
<script>
export default {
  async fetch() {
    let url = `/api/environments`;
    if (this.filter) {
      url += `?sortby=lastscaled&filter=` + encodeURIComponent(this.filter);
    }
    const result = await fetch(url);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result;
    }
    else {
      const text = await result.text();
      throw Error(text);
    }
  },
  data() {
    return {
      isLoading: false,
      errorText: "",
      tableFilter: "",
      fields: [
        { key: "Action", sortable: false, class: "text-center" },
        { key: "Name", sortable: false, class: "text-center" },
        { key: "GitBranch", sortable: false, class: "text-center" },
        { key: "Created", sortable: false, class: "text-center" },
        { key: "LastScaled", sortable: false, class: "text-center" }
      ],
      data: [],
      selectedEnvironments: []
    };
  },
  computed: {
    userLabel() {
      return this.const({ user: this.user.user }).LabelCreator;
    },
    filter() {
      return `${this.userLabel}=true`;
    }
  },
  methods: {
    getEnvironmentName(data) {
      if (data.NamespaceAnnotations && data.NamespaceAnnotations[this.const().LabelEnvironmentName]) {
        return data.NamespaceAnnotations[this.const().LabelEnvironmentName];
      }
      return data.Namespace;
    },
    getUserEnvironments() {
      return this.data;
    },
    getSelectedEnvironments() {
      const checkboxCollection = document.getElementsByClassName("delete-environments-checkbox");
      let selectedEnvironments = [];
      for (let i = 0; i < checkboxCollection.length; i++) {
        if (checkboxCollection[i].checked) {
          selectedEnvironments.push(checkboxCollection[i]);
        }
      }
      return selectedEnvironments;
    },
    showDeleteSelectedEnvironments() {
      this.selectedEnvironments = this.getSelectedEnvironments();
      this.$refs['delete-modal'].show();
    },
    async deleteSelectedEnvironments() {
      try {
        this.isLoading = true;
        this.errorText = "";
        this.selectedEnvironments.forEach(async (item) => {
          await this.callEndpoint(`/api/${item.dataset.id}/make-delete`);
        });
        // wait for 3 seconds to refresh data
        setTimeout(() => {
          this.$fetch();
          this.isLoading = false;
        }, 3000);
      }
      catch (e) {
        this.errorText = e.message;
        this.isLoading = false;
      }
    },
  }
}
</script>
