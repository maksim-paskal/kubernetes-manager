<template>
  <div class="detail-tab">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <b-spinner v-if="$fetchState.pending || callIsLoading" variant="primary" />
    <div v-else>
      <b-form-input v-model="dataFilter" autocomplete="off" placeholder="Type to Search" />
      <b-table style="margin-top:5px" striped hover :items="data.Result" :fields="dataFields" :filter="dataFilter">
        <template v-slot:cell(ServiceHost)="row">
          <CopyIcon :text="row.item.ServiceHost" />
          {{ row.item.ServiceHost }}&nbsp;<span v-if="row.item.Labels" class="badge rounded-pill bg-primary">{{
            row.item.Labels
          }}</span>
        </template>
        <template v-slot:cell(Actions)="row">
          <b-button size="sm" variant="outline-danger" @click="call('make-delete-pod', { PodName: row.item.ServiceHost })"
            v-if="row.item.Type == 'pod'">restart
          </b-button>
          <b-button size="sm" variant="outline-primary" @click="showProxyDialog(row)" v-if="row.item.Ports">proxy
          </b-button>
          <b-button size="sm" variant="outline-primary" @click="showShellDialog(row)" v-if="row.item.Type == 'pod'">
            shell</b-button>
          <b-button size="sm" variant="outline-primary" @click="showLogsDialog(row)" v-if="row.item.Type == 'pod'">
            logs</b-button>
          <b-button v-if="/.*mysql.*.svc.cluster.local$/.test(
            row.item.ServiceHost
          )
            " size="sm" variant="outline-primary" target="_blank"
            :href="`${environment.Links.PhpMyAdminURL}?host=${row.item.ServiceHost}`">
            phpmyadmin</b-button>
        </template>
      </b-table>
    </div>
    <b-modal size="xl" centered id="bv-proxy-dialog" title="Create proxy to service" ok-only>
      <WarningDisableMTLS />
      <KubectlLink />
      <div style="margin-top:20px">
        <CopyTextbox :text="proxyDialogText" />
      </div>
      <br />
      <a target="_blank"
        href="https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/#forward-a-local-port-to-a-port-on-the-pod">Link
        to documentation</a>
    </b-modal>
    <b-modal size="xl" centered id="bv-shell-dialog" title="Create shell to pod" ok-only>
      <KubectlLink />
      <div style="margin-top:20px">
        <CopyTextbox :text="shellDialogText" />
      </div>
    </b-modal>
    <b-modal size="xl" centered id="bv-logs-dialog" title="Pod logs" ok-only>
      <PodLogs v-if="this.selectedRow" :pod="this.selectedRow.item.Name" :logsPodURL="environment.Links.LogsPodURL" />
    </b-modal>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Services', true)
    }
  },
  mounted() {
    if (this.$route.hash) {
      this.dataFilter = this.$route.hash.substring(1);
    }
  },
  data() {
    return {
      data: {},
      dataFilter: null,
      dataFields: [
        { key: "ServiceHost", label: "Service Host" },
        { key: "Actions", label: "Actions" },
      ],
      proxyDialogText: "",
      shellDialogText: "",
      selectedRow: null,
    };
  },
  async fetch() {
    const result = await fetch(`/api/${this.$route.params.environmentID}/services`);
    if (result.ok) {
      this.data = await result.json();
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  computed: {
    kubeconfig() {
      return `/tmp/kubeconfig-${this.environment.Cluster}-${this.environment.Namespace}`
    }
  },
  methods: {
    getPodNameLink(podName) {
      let url = this.environment.Links.LogsPodURL

      url = url.replace(
        /__PodName__/g,
        podName
      );

      return url
    },
    showProxyDialog(row) {
      const port = row.item.Ports.split(",")[0];
      let proxyType = "pod";
      if (/.+.svc.cluster.local$/.test(row.item.ServiceHost)) {
        proxyType = "svc";
      }

      this.proxyDialogText = `kubectl --kubeconfig=${this.kubeconfig} -n ${this.environment.Namespace} port-forward ${proxyType}/${row.item.Name} ${port}:${port}`

      this.$bvModal.show('bv-proxy-dialog')
    },
    showShellDialog(row) {
      this.shellDialogText = `kubectl --kubeconfig=${this.kubeconfig} -n ${this.environment.Namespace} exec -it ${row.item.Name} -- sh`

      this.$bvModal.show('bv-shell-dialog')
    },
    showLogsDialog(row) {
      this.selectedRow = row
      this.$bvModal.show('bv-logs-dialog')
    },
  },
};
</script>