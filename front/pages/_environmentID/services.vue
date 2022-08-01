<template>
  <div class="detail-tab">
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-spinner v-if="$fetchState.pending" variant="primary" />
    <div v-else>
      <WarningDisableMTLS />
      <b-form-input v-model="dataFilter" autocomplete="off" placeholder="Type to Search" />
      <b-table striped hover :items="data.Result" :fields="dataFields" :filter="dataFilter">
        <template v-slot:cell(ServiceHost)="row">
          {{ row.item.ServiceHost }}&nbsp;<span v-if="row.item.Labels" class="badge rounded-pill bg-primary">{{
              row.item.Labels
          }}</span>
        </template>
        <template v-slot:cell(Ports)="row">
          <b-button size="sm" variant="outline-primary" @click="showProxyDialog(row)">proxy</b-button>
          <b-button size="sm" target="_blank" v-if="environment.Links.LogsPodURL && row.item.Type == 'pod'"
            variant="outline-primary" :href="getPodNameLink(row.item.Name)">logs</b-button>
          <b-button v-if="
            /.*mysql.*.svc.cluster.local$/.test(
              row.item.ServiceHost
            )
          " size="sm" variant="outline-primary" target="_blank"
            :href="`${environment.Links.PhpMyAdminURL}?host=${row.item.ServiceHost}`">
            phpmyadmin</b-button>&nbsp;{{ row.value }}
        </template>
      </b-table>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  data() {
    return {
      data: {},
      dataFilter: null,
      dataFields: [
        { key: "ServiceHost", label: "Service Host" },
        { key: "Ports", label: "Service Ports" },
      ],
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
      const proxyString =
        `"Save As" this <a target="_blank" href="/api/${this.environment.ID}/kubeconfig">file</a> to ${this.kubeconfig}` +
        `<br/><br/><textarea readonly style="background-color:#eeeeee;border:0px;padding:10px;outline:none;width:100%" onclick="this.focus();this.select()">kubectl --kubeconfig=${this.kubeconfig} -n ${this.environment.Namespace} port-forward ${proxyType}/${row.item.Name} ${port}:${port}</textarea>` +
        `<br/><br/>this will listen localy 127.0.0.1:${port} and forward all requests to service in cluster, to listen other port for example 127.0.0.1:12345 - change end of command from ${port}:${port} to <strong>12345</strong>:${port}`;

      const h = this.$createElement;
      const messageVNode = h("div", { domProps: { innerHTML: proxyString } });

      this.$bvModal.msgBoxOk([messageVNode], {
        title: `Create proxy to remote service: ${proxyType}/${row.item.Name}`,
        size: "xl",
        centered: true,
      });
    },
  },
};
</script>