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

    <div v-if="$fetchState.pending || callIsLoading" style="padding: 50px" class="text-center">
      <b-spinner style="width: 10rem; height: 10rem" variant="primary" />
    </div>
    <div v-else>
      <div style="padding:10px">
        <b-form-input v-model="tableFilter" autocomplete="off" placeholder="Type to Search" />
      </div>
      <b-table striped hover :fields="fields" :items="data" :filter="tableFilter">
        <template v-slot:cell(Address)="row">
          {{ row.item.IPv4 }}
        </template>
        <template v-slot:cell(Status)="row">
          {{ row.item.Status }}
          <div v-if="row.item.Status == 'Stoped'">
            <b-button size="sm" variant="success" @click="serverAction(row,'power_on')">
              Start
            </b-button>
          </div>
        </template>
        <template v-slot:cell(Actions)="row">
          <b-button size="sm" variant="outline-primary" @click="showConfigDialog(row)">Local
            configuration
          </b-button>
        </template>
      </b-table>
      <b-modal size="xl" centered id="bv-remote-servers-config-dialog" title="Local configuration" ok-only>
        <h3>MacOS users</h3>
        <CopyTextbox :text="darwinText" />
        <br />
        <h3>Linux users</h3>
        <CopyTextbox :text="linuxText" />
      </b-modal>
    </div>
  </div>
</template>
<script>
export default {
  mounted() {
    this.tableFilter = this.user.user;
  },
  async fetch() {
    const result = await fetch('/api/remote-servers');
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
      errorText: "",
      infoText: "",
      tableFilter: "",
      fields: [
        { key: "Name", sortable: false, class: "text-center" },
        { key: "Address", sortable: false, class: "text-center" },
        { key: "Status", sortable: false, class: "text-center" },
        { key: "Actions", sortable: false, class: "col-deploy-service-text" },
      ],
      darwinText: "",
      linuxText: "",
      data: []
    }
  },
  methods: {
    showConfigDialog(row) {
      this.darwinText = `sudo curl -sSL https://get.paskal-dev.com/local-dev-darwin.sh | sudo REMOTE_IP=${row.item.IPv4} sh`
      this.linuxText = `sudo curl -sSL https://get.paskal-dev.com/local-dev-linux.sh | sudo REMOTE_IP=${row.item.IPv4} sh`

      this.$bvModal.show('bv-remote-servers-config-dialog')
    },
    async serverAction(row, action) {
      await this.callEndpoint('/api/make-remote-server-action', {
        Cloud: row.item.Cloud,
        ID: row.item.ID,
        Action: action
      }, true);
    }
  }
}
</script>
