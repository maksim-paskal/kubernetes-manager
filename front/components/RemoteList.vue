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
          <CopyIcon :text="row.item.IPv4" />{{ row.item.IPv4 }}
        </template>
        <template v-slot:cell(Name)="row">
          {{ row.item.Name }}
          <EnvironmentBadges :badges="row.item.FormattedLabels" />
        </template>
        <template v-slot:cell(Status)="row">
          {{ row.item.Status }}
          <div>
            <b-button v-if="row.item.Status == 'Stoped'" size="sm" variant="success"
              @click="serverAction(row, 'PowerOn')">
              Start
            </b-button>
            <b-button v-else size="sm" variant="danger" @click="serverAction(row, 'PowerOff')">
              Stop
            </b-button>
          </div>
        </template>
        <template v-slot:cell(Actions)="row">
          <b-button size="sm" variant="outline-primary" @click="showConfigDialog(row)">Settings</b-button>
          <b-button size="sm" variant="outline-primary" @click="delayAutopause(row)">Delay autopause for next 3
            hours</b-button>
          <div v-if="row.item.Status == 'Running'">The server will work till <strong>{{ getScaleDownDelay(row) }}</strong>
            your local time</div>
        </template>
      </b-table>
      <b-modal size="xl" centered id="bv-remote-servers-config-dialog" title="Run these commands in your local terminal"
        ok-only>
        <b-tabs content-class="mt-3">
          <b-tab :title="item.Name" v-for="(item, index) in this.links" :key="index">
            <p v-if="item.Description">{{ item.Description }}</p>
            <CopyTextbox :text="item.URL" />
          </b-tab>
        </b-tabs>
      </b-modal>
    </div>
  </div>
</template>
<script>
export default {
  mounted() {
    if (this.$route.hash) {
      this.tableFilter = this.$route.hash.substring(1);
    } else {
      this.tableFilter = this.user.user;
    }
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
    getScaleDownDelay(row) {
      const lang = navigator.language | window.navigator.language
      let scaleDownDelay = Date.now();

      if (row.item.Labels?.scaleDownDelay) {
        const epoch = parseInt(row.item.Labels.scaleDownDelay);
        let epochDate = new Date(0);
        epochDate.setUTCSeconds(epoch);
        scaleDownDelay = epochDate;
      }

      let scaleDownDelayDate = new Date(scaleDownDelay)

      if (isNaN(scaleDownDelayDate) || scaleDownDelayDate < Date.now()) {
        scaleDownDelayDate = Date.now()
      }

      const isToday = (new Date().toDateString() == new Date(scaleDownDelayDate).toDateString());

      if (isToday) {
        return `Today at ${new Intl.DateTimeFormat(lang, { timeStyle: 'medium' }).format(scaleDownDelayDate)}`
      }

      const options = { dateStyle: 'full', timeStyle: 'medium' };

      return new Intl.DateTimeFormat(lang, options).format(scaleDownDelayDate)
    },
    showConfigDialog(row) {
      this.links = []
      row.item.Links.forEach((item) => {
        this.links.push({
          Name: item.Name,
          Description: item.Description,
          URL: item.URL
        })
      })

      this.$bvModal.show('bv-remote-servers-config-dialog')
    },
    reload() {
      this.callIsLoading = true;
      // wait for 3 seconds to refresh data
      setTimeout(() => {
        this.$fetch()
        this.callIsLoading = false;
      }, 2000);
    },
    async serverAction(row, action) {
      await this.callEndpoint('/api/make-remote-server-action', {
        Cloud: row.item.Cloud,
        ID: row.item.ID,
        Action: action
      }, true);

      this.reload();
    },
    async delayAutopause(row) {
      await this.callEndpoint('/api/make-remote-server-delay', {
        Cloud: row.item.Cloud,
        ID: row.item.ID,
        Duration: '3h'
      }, true);

      this.reload();
    }
  }
}
</script>
