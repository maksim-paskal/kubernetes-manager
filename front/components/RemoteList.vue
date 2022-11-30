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
            <b-button size="sm" variant="success" @click="serverAction(row, 'power_on')">
              Start
            </b-button>
          </div>
        </template>
        <template v-slot:cell(Actions)="row">
          <b-button size="sm" variant="outline-primary" @click="showConfigDialog(row)">Local
            configuration
          </b-button>
          <b-button size="sm" variant="outline-primary" @click="delayAutopause(row)">Delay autopause for next 3
            hours</b-button>
        </template>
      </b-table>
      <b-modal size="xl" centered id="bv-remote-servers-config-dialog" title="Run this commands in your local terminal"
        ok-only>
        <div v-for="(item, index) in this.links" :key="index">
          <h3>{{ item.Name }}</h3>
          <CopyTextbox :text="item.URL" />
          <br />
        </div>
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
      data: [],
      links: []
    }
  },
  methods: {
    showConfigDialog(row) {
      this.links = []
      this.config.RemoteServersLinks.forEach((item) => {
        let url = item.URL

        url = url.replaceAll("{REMOTE_IP}", row.item.IPv4)
        if (this.user.user) {
          url = url.replaceAll("{USER_LOGIN}", this.user.user)
        }

        this.links.push({
          Name: item.Name,
          URL: url
        })
      })

      this.$bvModal.show('bv-remote-servers-config-dialog')
    },
    async serverAction(row, action) {
      await this.callEndpoint('/api/make-remote-server-action', {
        Cloud: row.item.Cloud,
        ID: row.item.ID,
        Action: action
      }, true);
    },
    async delayAutopause(row) {
      await this.callEndpoint('/api/make-remote-server-delay', {
        Cloud: row.item.Cloud,
        ID: row.item.ID,
        Duration: '3h'
      }, true);
    }
  }
}
</script>
