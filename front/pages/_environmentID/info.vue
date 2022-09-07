<template>
  <div class="detail-tab">
    <div v-if="this.environment.Links">
      <b-button target="_blank" :href="environment.Links.MetricsURL"><img height="16" alt="grafana"
          src="~assets/grafana.png" />&nbsp;Metrics</b-button>
      <b-button target="_blank" :href="environment.Links.LogsURL"><img height="16" alt="elasticsearch"
          src="~assets/elasticsearch.png" />&nbsp;Logs</b-button>
      <b-button target="_blank" :href="environment.Links.TracingURL"><img height="16" alt="jaeger"
          src="~assets/jaeger.png" />&nbsp;Traces</b-button>
      <b-button target="_blank" :href="environment.Links.SentryURL"><img height="16" alt="sentry"
          src="~assets/sentry.png" />&nbsp;Sentry</b-button>
      <b-button @click="showJSON = !showJSON">{{ showJSON ? "Hide" : "Show" }} info</b-button>
    </div>

    <b-card v-if="showJSON" class="mt-3" header="Namespace information">
      <b-spinner v-if="!environment.ID" variant="primary" />
      <pre v-else class="m-0">{{ environment }}</pre>
    </b-card>

    <b-card class="mt-3" header="Hosts">
      <ul v-if="environment.Hosts && environment.Hosts.length > 0">
        <li v-bind:key="index" v-for="(item, index) in environment.Hosts">
          <a class="text-decoration-none" :href="item" rel="noopener" target="_blank">{{ item }}</a>
        </li>
      </ul>
      <div v-else>Wait when <b-link class="text-decoration-none" to="external-services">external-services</b-link>
        deployed...</div>
    </b-card>

    <b-card class="mt-3" header="Internal Hosts"
      v-if="environment.HostsInternal && environment.HostsInternal.length > 0">
      <WarningDisableMTLS />
      <ul>
        <li v-bind:key="index" v-for="(item, index) in environment.HostsInternal">
          <a class="text-decoration-none" :href="item" rel="noopener" target="_blank">{{ item }}</a>
        </li>
      </ul>
    </b-card>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Info', true)
    }
  },
  data() {
    return {
      showJSON: false,
    };
  },
}
</script>