<template>
  <div>
    <b-navbar sticky fixed="top" toggleable="lg" type="dark" variant="dark">
      <b-input-group>
        <b-button variant="outline-secondary" to="/create"><svg xmlns="http://www.w3.org/2000/svg" width="20"
            height="20" fill="currentColor" class="bi bi-plus-lg" viewBox="0 0 16 16">
            <path fill-rule="evenodd"
              d="M8 2a.5.5 0 0 1 .5.5v5h5a.5.5 0 0 1 0 1h-5v5a.5.5 0 0 1-1 0v-5h-5a.5.5 0 0 1 0-1h5v-5A.5.5 0 0 1 8 2Z">
            </path>
          </svg>&nbsp;Create</b-button>&nbsp;
        <b-button :disabled="isBusy" variant="outline-secondary" @click="getIngress()"><svg
            xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" class="bi bi-arrow-repeat"
            viewBox="0 0 16 16">
            <path
              d="M11.534 7h3.932a.25.25 0 0 1 .192.41l-1.966 2.36a.25.25 0 0 1-.384 0l-1.966-2.36a.25.25 0 0 1 .192-.41zm-11 2h3.932a.25.25 0 0 0 .192-.41L2.692 6.23a.25.25 0 0 0-.384 0L.342 8.59A.25.25 0 0 0 .534 9z">
            </path>
            <path fill-rule="evenodd"
              d="M8 3c-1.552 0-2.94.707-3.857 1.818a.5.5 0 1 1-.771-.636A6.002 6.002 0 0 1 13.917 7H12.9A5.002 5.002 0 0 0 8 3zM3.1 9a5.002 5.002 0 0 0 8.757 2.182.5.5 0 1 1 .771.636A6.002 6.002 0 0 1 2.083 9H3.1z">
            </path>
          </svg>&nbsp;Refresh</b-button>&nbsp;
        <b-dropdown variant="outline-secondary" text="Menu">
          <b-dropdown-item v-for="(item, index) in this.config ? this.config.Links.Others : []" :key="index"
            target="_blank" :href="item.URL">
            <em style="font-size: 24px" class="bi bi-box-arrow-up-right" />&nbsp;{{ item.Name }}
          </b-dropdown-item>
          <b-dropdown-item target="_blank" v-if="this.config && this.config.Links.SlackURL"
            :href="this.config.Links.SlackURL">
            <em style="font-size: 24px" class="bi bi-slack" />&nbsp;Report Issue
          </b-dropdown-item>
          <b-dropdown-divider></b-dropdown-divider>
          <b-dropdown-item v-if="this.config" target="_blank"
            href="https://github.com/maksim-paskal/kubernetes-manager"><em style="font-size: 24px"
              class="bi bi-github" />&nbsp;{{
                  this.config.Version
              }}</b-dropdown-item>
        </b-dropdown>&nbsp;
        <b-form-input v-model="filter" :disabled="isBusy" autocomplete="off" placeholder="Type to Search" />&nbsp;
      </b-input-group>
    </b-navbar>

    <div style="padding: 50px" class="text-center" v-if="isBusy">
      <b-spinner style="width: 10rem; height: 10rem" variant="primary"></b-spinner>
    </div>

    <b-modal ref="info-modal" :id="infoModal.id" :title="infoModal.title" size="xl" ok-only>
      <div style="padding: 50px" class="text-center" v-if="infoModal.loading">
        <b-spinner style="width: 10rem; height: 10rem" variant="primary"></b-spinner>
      </div>

      <b-alert v-if="
        infoModal.info && !infoModal.info.Stdout && !infoModal.info.Stderr
      " variant="danger" show>
        <pre>{{ infoModal.info }}</pre>
      </b-alert>

      <b-alert v-if="infoModal.info && infoModal.info.Stderr" variant="danger" show>
        <pre>{{ infoModal.info.Stderr }}</pre>
      </b-alert>

      <b-alert v-if="infoModal.info && infoModal.info.Stdout" variant="info" show>
        <pre>{{ infoModal.info.Stdout }}</pre>
      </b-alert>

      <b-card no-body v-if="!infoModal.loading">
        <b-tabs card @input="showTab" v-model="tabIndex">
          <b-tab key="tab0" title="info">
            <b-button v-if="
              this.infoModal.content.FrontConfig &&
              this.infoModal.content.FrontConfig.Links.MetricsURL
            " :href="
  getNamespacedString(
    this.infoModal.content.FrontConfig.Links.MetricsURL
  )
" target="_blank">
              <img height="16" alt="grafana" src="~assets/grafana.png" />&nbsp;Metrics
            </b-button>
            <b-button v-if="
              this.infoModal.content.FrontConfig &&
              this.infoModal.content.FrontConfig.Links.LogsURL
            " :href="
  getNamespacedString(
    this.infoModal.content.FrontConfig.Links.LogsURL
  )
" target="_blank">
              <img height="16" alt="elasticsearch" src="~assets/elasticsearch.png" />&nbsp;Logs
            </b-button>
            <b-button v-if="
              this.infoModal.content.FrontConfig &&
              this.infoModal.content.FrontConfig.Links.TracingURL
            " :href="
  getNamespacedString(
    this.infoModal.content.FrontConfig.Links.TracingURL
  )
" target="_blank">
              <img height="16" alt="jaeger" src="~assets/jaeger.png" />&nbsp;Traces
            </b-button>
            <b-button v-if="
              this.infoModal.content.FrontConfig &&
              this.infoModal.content.FrontConfig.Links.SentryURL
            " :href="
  getNamespacedString(
    this.infoModal.content.FrontConfig.Links.SentryURL
  )
" target="_blank">
              <img height="16" alt="sentry" src="~assets/sentry.png" />&nbsp;Sentry
            </b-button>

            <b-card class="mt-3" header="Namespace information">
              <pre class="m-0">{{ infoModal.content }}</pre>
            </b-card>
          </b-tab>
          <b-tab key="tab1" title="services" :disabled="this.infoModal.content.RunningPodsCount == 0">
            <b-alert variant="warning" show>
              <b-button @click="makeAPICall('disableMTLS', 'none')">Disable mTLS verification</b-button>&nbsp;For proper
              usage you must disable mutual TLS verification
            </b-alert>
            <div>
              <b-form-input v-model="serviceFilter" autocomplete="off" placeholder="Type to Search" />
              <b-table striped hover :items="tab1Data" :fields="tab1DataFields" :filter="serviceFilter">
                <template v-slot:cell(ServiceHost)="row">
                  {{ row.item.ServiceHost }}&nbsp;<span v-if="row.item.Labels" class="badge rounded-pill bg-primary">{{
                      row.item.Labels
                  }}</span>
                </template>
                <template v-slot:cell(Ports)="row">
                  <b-button size="sm" variant="outline-primary" @click="showProxyDialog(row)">proxy</b-button>
                  <b-button v-if="
                    /.*mysql.*.svc.cluster.local$/.test(
                      row.item.ServiceHost
                    ) &&
                    infoModal.content.FrontConfig &&
                    infoModal.content.FrontConfig.Links.PhpMyAdminURL
                  " size="sm" variant="outline-primary" target="_blank"
                    :href="`${infoModal.content.FrontConfig.Links.PhpMyAdminURL}?host=${row.item.ServiceHost}`">
                    phpmyadmin</b-button>&nbsp;{{ row.value }}
                </template>
              </b-table>
              <br />
              <b-button @click="showTab(1, true, true)"><em class="bi bi-arrow-clockwise" />&nbsp;Refresh</b-button>
            </div>
          </b-tab>
          <b-tab key="tab2" title="mongo" :disabled="true">
            <pre>{{ tab2Data }}</pre>
            <div v-if="tab2Data != null">
              <b-button @click="showTab(2, true, true)"><em class="bi bi-arrow-clockwise" />&nbsp;Refresh</b-button>
              <b-button v-if="tab2Data.result == 'found'" variant="danger" @click="makeAPICall('exec', 'mongoDelete')">
                <em class="bi bi-x-octagon-fill" />&nbsp;Delete
              </b-button>
              <b-button variant="outline-primary" @click="makeAPICall('exec', 'mongoMigrations')">Migrations</b-button>
            </div>
          </b-tab>
          <b-tab key="tab3" title="settings">
            <b-card bg-variant="light" header="Pause branch" class="text-center">
              <div v-if="this.config">
                branch autopause
                {{ this.config.Batch.ScaleDownHourMinPeriod }}:00 -
                {{ this.config.Batch.ScaleDownHourMaxPeriod }}:00
                {{ this.config.Batch.BatchSheduleTimezone }}
              </div>
              <br />
              <b-button @click="makeAPICall('scaleNamespace', 'none', '&replicas=0')"><em
                  class="bi bi-pause-fill" />&nbsp;Pause</b-button>
              <b-button variant="success" @click="makeAPICall('scaleNamespace', 'none', '&replicas=1')"><em
                  class="bi bi-play-fill" />&nbsp;Start</b-button>
              <b-button @click="makeAPICall('scaleDownDelay', 'none', '&duration=3h')">Delay autopause for next 3 hours
              </b-button>
            </b-card>
            <br />
            <b-card bg-variant="light" header="Autoscaling" class="text-center">
              <b-button @click="makeAPICall('disableHPA', 'none')">Disable autoscaling</b-button>
            </b-card>
            <b-card bg-variant="light" header="Envoy Control Plane" class="text-center">
              <b-button @click="makeAPICall('disableMTLS', 'none')">Disable mTLS verification</b-button>
            </b-card>
            <br />
            <b-card bg-variant="light" header="Danger Zone" class="text-center">
              <b-button variant="danger" @click="makeAPICall('deleteALL', 'projectID=')"><em
                  class="bi bi-x-octagon-fill" />&nbsp;Delete All</b-button>
            </b-card>
          </b-tab>
          <b-tab key="tab4" title="debug" :disabled="
            !hasDefaultPod || this.infoModal.content.RunningPodsCount == 0
          ">
            <b-alert variant="warning" show v-if="podsNamesSelectedTotal > 1">
              <b-button @click="makeAPICall('disableHPA', 'none')">Disable autoscaling</b-button>&nbsp;For proper usage
              you must disable autoscaler
            </b-alert>
            <b-form-select class="form-select" v-model="podsNamesSelected" v-on:change="podsNamesChange"
              :options="podsNames" />
            <div v-if="!podsNamesSelected">
              <b-button @click="showTab(4, true, true)"><em class="bi bi-arrow-clockwise" />&nbsp;Refresh</b-button>
            </div>
            <div v-else>
              <b-card-text>
                XDEBUG is
                <strong>{{ this.debug_enabled }}</strong>
              </b-card-text>
              <b-card-text>extra setting in php-fpm.conf</b-card-text>

              <b-container class="bv-example-row">
                <b-row>
                  <b-col cols="9">
                    <b-form-textarea spellcheck="false" rows="10" v-model="debug_text"
                      placeholder="Enter something..." />
                  </b-col>
                  <b-col cols="2">
                    <b-dropdown id="dropdown-1" offset="25" text="Templates" class="m-2">
                      <b-dropdown-item v-for="(item, index) in this.config
                      ? this.filterTemplates(this.config.DebugTemplates)
                      : []" :key="index" @click="templateAction(item.Data)">{{ item.Display }}</b-dropdown-item>
                    </b-dropdown>
                  </b-col>
                </b-row>
              </b-container>

              <b-card-text>
                Use
                <strong>ngrok tcp 9000</strong>
              </b-card-text>
              <b-button @click="showTab(4, true, true)"><em class="bi bi-arrow-clockwise" />&nbsp;Refresh</b-button>
              <b-button variant="success" @click="makeAPICall('exec', 'xdebugEnable')">Enable XDEBUG</b-button>
              <b-button @click="setPhpSettings()">Set settings in php-fpm.conf</b-button>
              <b-button variant="danger" @click="makeAPICall('deletePod', 'none')"><em
                  class="bi bi-x-octagon-fill" />&nbsp;Delete Pod</b-button>
            </div>
          </b-tab>
          <b-tab key="tab5" title="git-sync" :disabled="
            !hasDefaultPod || this.infoModal.content.RunningPodsCount == 0
          ">
            <b-alert variant="warning" show v-if="podsNamesSelectedTotal > 1">
              <b-button @click="makeAPICall('disableHPA', 'none')">Disable autoscaling</b-button>&nbsp;For proper usage
              you must disable autoscaler
            </b-alert>
            <b-form-select class="form-select" v-model="podsNamesSelected" v-on:change="podsNamesChange"
              :options="podsNames" />
            <div v-if="!podsNamesSelected">
              <b-button @click="showTab(5, true, true)"><em class="bi bi-arrow-clockwise" />&nbsp;Refresh</b-button>
            </div>
            <div v-else>
              <b-card-text>
                Sync is
                <strong>{{ this.gitSyncEnabled }}</strong>
              </b-card-text>
              <b-container fluid>
                <b-row class="my-1">
                  <b-col sm="3">
                    <label for="input-gitOrigin">origin:</label>
                  </b-col>
                  <b-col sm="9">
                    <b-form-input id="input-gitOrigin" :state="gitOriginState" v-model="gitOrigin"></b-form-input>
                  </b-col>
                </b-row>
                <b-row class="my-2">
                  <b-col sm="3">
                    <label for="input-gitBranch">branch:</label>
                  </b-col>
                  <b-col sm="9">
                    <b-form-input id="input-gitBranch" :state="gitBranchState" v-model="gitBranch"></b-form-input>
                  </b-col>
                </b-row>
                <b-row v-if="this.gitSyncShowPublicKey">
                  <b-col>
                    <b-form-textarea rows="5" v-model="gitSyncPublicKey" readonly />
                  </b-col>
                </b-row>
              </b-container>

              <br />
              <b-button @click="showTab(5, true, true)"><em class="bi bi-arrow-clockwise" />&nbsp;Refresh</b-button>
              <b-button variant="success" @click="enableGit()">Init</b-button>
              <b-button @click="showPublicKey()">Show public key</b-button>
              <b-button variant="danger" @click="makeAPICall('deletePod', 'none')"><em
                  class="bi bi-x-octagon-fill" />&nbsp;Delete Pod</b-button>
              <b-button @click="makeAPICall('exec', 'gitFetch')">Fetch</b-button>
              <b-button @click="makeAPICall('exec', 'clearCache')">Clear Cache</b-button>
            </div>
          </b-tab>

          <b-tab key="tab6" title="kubectl">
            <b-card no-body>
              <b-tabs pills card vertical>
                <b-tab title="1. Install" active>
                  <b-card-text>
                    Install requred application
                    <a target="_blank"
                      href="https://kubernetes.io/docs/tasks/tools/install-kubectl/">https://kubernetes.io/docs/tasks/tools/install-kubectl/</a>
                  </b-card-text>
                </b-tab>
                <b-tab title="2. Configure">
                  <b-card-text>
                    "Save As" this
                    <a target="_blank" :href="`/getKubeConfig?cluster=${infoModal.content.Cluster}`">file</a>
                    to /tmp/kubeconfig-{{ infoModal.content.Cluster }}
                  </b-card-text>
                </b-tab>
                <b-tab title="3. Test">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig-{{
                      infoModal.content.Cluster
                  }}
                    -n {{ infoModal.content.NamespaceName }} get
                    pods</b-card-text>
                </b-tab>
                <b-tab title="4. Shell">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig-{{
                      infoModal.content.Cluster
                  }}
                    -n {{ infoModal.content.NamespaceName }} exec -it `kubectl
                    --kubeconfig=/tmp/kubeconfig-{{
                        infoModal.content.Cluster
                    }}
                    -n {{ infoModal.content.NamespaceName }} get pods -l{{
                        this.defaultPodInfo[0]
                    }}
                    -o jsonpath='{.items[0].metadata.name}'` -c
                    {{ this.defaultPodInfo[1] }} sh</b-card-text>
                </b-tab>
                <b-tab title="5. Logs">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig-{{
                      infoModal.content.Cluster
                  }}
                    -n {{ infoModal.content.NamespaceName }} logs `kubectl
                    --kubeconfig=/tmp/kubeconfig-{{
                        infoModal.content.Cluster
                    }}
                    -n {{ infoModal.content.NamespaceName }} get pods -l{{
                        this.defaultPodInfo[0]
                    }}
                    -o jsonpath='{.items[0].metadata.name}'` -c
                    {{ this.defaultPodInfo[1] }}</b-card-text>
                </b-tab>
                <b-tab title="6. Clear memcahed">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig-{{
                      infoModal.content.Cluster
                  }}
                    -n {{ infoModal.content.NamespaceName }} delete `kubectl
                    --kubeconfig=/tmp/kubeconfig-{{
                        infoModal.content.Cluster
                    }}
                    -n {{ infoModal.content.NamespaceName }} get pods
                    -l=app=memcached -o name`</b-card-text>
                </b-tab>
              </b-tabs>
            </b-card>
          </b-tab>

          <b-tab key="tab7" title="external services" :disabled="this.infoModal.content.RunningPodsCount == 0">
            <div>
              <b-form-input v-model="externalServiceFilter" autocomplete="off" placeholder="Type to Search" />
              <b-table striped hover :items="tab7Data" :fields="tab7DataFields" :filter="externalServiceFilter">
                <template #cell(Service)="data">
                  <a target="_blank" :href="data.item.WebURL" style="text-decoration: none">{{ data.item.Description
                  }}</a>&nbsp;<span v-if="data.item.AdditionalInfo" title="docker tag"
                    class="badge rounded-pill bg-primary">{{ data.item.AdditionalInfo.PodRunning.Tag }}</span>&nbsp;<a
                    v-if="
                      data.item.AdditionalInfo &&
                      data.item.AdditionalInfo.PodRunning.GitHash
                    " target="_blank" :href="getGitlabCommitURL(data.item)"><span title="git short commit hash"
                      class="badge rounded-pill bg-success">{{ data.item.AdditionalInfo.PodRunning.GitHash }}</span></a>
                  <div v-if="data.item.TagsList.length > 0">
                    <div style="margin-left: 5px" class="badge btn-secondary" v-bind:key="index"
                      v-for="(item, index) in data.item.TagsList">
                      {{ item }}
                    </div>
                  </div>
                </template>
                <template #cell(Status)="data">
                  <div style="height: 25px">
                    <b-spinner v-if="!data.item.AdditionalInfo" variant="primary" />
                    <em v-if="
                      data.item.AdditionalInfo &&
                      data.item.AdditionalInfo.PodRunning.Found
                    " class="bi bi-check-circle-fill" style="font-size: 26px; color: green" />
                    <a target="_blank" :href="
                      data.item.AdditionalInfo.Pipelines.LastErrorPipeline
                    " v-if="
  data.item.AdditionalInfo &&
  data.item.AdditionalInfo.Pipelines.LastErrorPipeline
"><em class="bi bi-exclamation-circle-fill" style="font-size: 26px; color: #dc3545" /></a>
                    <a target="_blank" :href="
                      data.item.AdditionalInfo.Pipelines.LastRunningPipeline
                    " v-if="
  data.item.AdditionalInfo &&
  data.item.AdditionalInfo.Pipelines.LastRunningPipeline
"><em class="bi bi-hourglass-split" style="font-size: 26px; color: #1f75cb" /></a>
                  </div>
                </template>
                <template #cell(Deploy)="data">
                  <b-button style="width: 100px" @click="getProjectRefs(data)" :variant="deployButtonVariant(data)">
                    <div v-if="data.item.Deploy && data.item.Deploy.length <= 6">
                      {{ data.item.Deploy }}
                    </div>
                    <div v-else-if="data.item.Deploy" :title="data.item.Deploy">
                      {{ data.item.Deploy.substr(0, 3) }}...
                    </div>
                    <em v-else class="bi bi-dash-lg" />
                  </b-button>
                </template>
              </b-table>
              <br />
              <b-button @click="showTab(7, true, true)"><em class="bi bi-arrow-clockwise" />&nbsp;Refresh</b-button>
              <b-button variant="success" :disabled="!tab7DataSelected" @click="deploySelectedServices()">Deploy
                Selected
              </b-button>
              <b-dropdown id="dropdown-2" offset="25" text="Templates" class="m-2">
                <b-dropdown-item v-for="(item, index) in this.config
                ? this.filterTemplates(
                  this.config.ExternalServicesTemplates
                )
                : []" :key="index" @click="externalServiceTemplate(item.Data)">{{ item.Display }}</b-dropdown-item>
              </b-dropdown>
            </div>
          </b-tab>
        </b-tabs>
      </b-card>
    </b-modal>

    <div style="padding: 5px" v-if="!isBusy && items == null">
      <b-alert variant="warning" show>No available namespaces founded</b-alert>
    </div>

    <div style="padding: 5px" v-if="error != null">
      <b-alert variant="danger" show>{{ error }}</b-alert>
    </div>

    <b-table striped hover :fields="fields" :items="items" v-if="!isBusy && items != null" :filter="filter">
      <template v-slot:cell(Status)="data">
        <b-spinner v-if="data.item.RunningPodsCount < 0" variant="primary"></b-spinner>
        <div v-else>
          <div v-if="data.item.RunningPodsCount == 0">
            Paused
            <br />
            <b-button variant="success" @click="
              unpauseNamespace(
                data.item.Namespace,
                data.item.IngressAnotations['kubernetes-manager/version']
              )
            ">Start</b-button>
          </div>
          <div>{{ getNamespaceStatus(data.item) }}</div>
        </div>
      </template>
      <template v-slot:cell(Hosts)="data">
        <ul>
          <li v-bind:key="index" v-for="(item, index) in data.value">
            <a :href="item" rel="noopener" target="_blank">{{ item }}</a>
          </li>
        </ul>
      </template>
      <template v-slot:cell(Actions)="row">
        <b-button size="sm" variant="outline-primary" @click="info(row.item, row.index, $event.target)">Details
        </b-button>
      </template>
    </b-table>
  </div>
</template>

<script>
export default {
  created() {
    this.getFrontConfig();
  },
  mounted() {
    this.getIngress();

    if (this.$route.hash) {
      this.filter = this.$route.hash.substring(1);
    }
  },
  computed: {
    gitBranchState() {
      if (!this.gitBranch) return false;
      return this.gitBranch.trim().length > 0 ? true : false;
    },
    gitOriginState() {
      if (!this.gitOrigin) return false;
      return this.gitOrigin.match("^[^@]+@[^:]+:[^/]+/[^.]+.git$") != null;
    },
  },
  data() {
    return {
      config: null,
      fields: [
        { key: "Actions", sortable: false, class: "text-center" },
        { key: "Status", sortable: false, class: "text-center" },
        { key: "GitBranch", sortable: false },
        { key: "Hosts", sortable: false },
      ],
      isBusy: true,
      error: null,
      items: [],
      filter: null,
      serviceFilter: null,
      externalServiceFilter: null,
      infoModal: {
        id: "info-modal",
        title: "",
        content: "",
        loading: false,
        error: null,
        info: null,
      },
      tabIndex: null,
      tab1Data: null,
      tab1DataFields: [
        { key: "ServiceHost", label: "Service Host" },
        { key: "Ports", label: "Service Ports" },
      ],
      tab2Data: null,
      tab4Data: null,
      tab5Data: null,
      tab7Data: null,
      tab7DataSelected: false,
      tab7DataFields: [
        { key: "Status", label: "Status", class: "text-center" },
        { key: "Deploy", label: "Deploy", class: "text-center" },
        { key: "Service", label: "Service", tdClass: "col-lg-9" },
      ],
      debug_enabled: "unknown",
      debug_text: "",
      gitOrigin: "",
      gitBranch: "",
      gitSyncEnabled: "",
      gitSyncShowPublicKey: false,
      gitSyncPublicKey: "",
      defaultPodInfo: "",

      podsNames: [],
      podsNamesSelectedTotal: 0,
      podsNamesSelected: null,

      hasDefaultPod: false,
    };
  },
  methods: {
    async getFrontConfig() {
      try {
        this.config = await this.$axios.$get("/api/getFrontConfig");
      } catch (e) {
        console.error(e);
        this.showAxiosError(e);
      }
    },
    async podsNamesChange() {
      this.showTab(this.tabIndex, true);
    },
    getNamespacedString(data) {
      return data.replace(
        /__Namespace__/g,
        this.infoModal.content.NamespaceName
      );
    },
    getNamespaceStatus(data) {
      const total = data.RunningPodsCount;
      const required =
        data.IngressAnotations["kubernetes-manager/requiredRunningPodsCount"];

      if (total == 0) {
        return;
      }
      if (!required || total >= required) {
        return "Ready";
      }
      return "Waiting";
    },
    async unpauseNamespace(namespace, version) {
      this.isBusy = true;
      this.error = null;
      try {
        var url = `/api/scaleNamespace?namespace=${namespace}&replicas=1`;
        if (version) {
          url += `&version=${version}`;
        }
        await this.$axios.$get(url);
      } catch (e) {
        console.error(e);
        this.showAxiosError(e);
      }
      this.isBusy = false;
    },
    async getIngress() {
      this.isBusy = true;
      this.error = null;
      try {
        const data1 = await this.$axios.$get("/api/getIngress");
        this.items = data1.result;

        if (this.items) {
          this.items.forEach(async (el) => {
            const data2 = await this.$axios.$get(
              "/api/getRunningPodsCount?namespace=" + el.Namespace
            );
            el.RunningPodsCount = data2.count;
          });
        }
      } catch (e) {
        console.error(e);
        this.showAxiosError(e);
      } finally {
        this.isBusy = false;
      }
    },
    async templateAction(data) {
      this.debug_text += `${data}\n`;
    },
    async showPublicKey() {
      this.gitSyncShowPublicKey = !this.gitSyncShowPublicKey;
    },
    makeAPICallUrl(api, cmd = "none", args = "" /* need & */) {
      for (var key in this.infoModal.content.IngressAnotations) {
        if (key.startsWith("kubernetes-manager")) {
          args += `&${key.substring(19)}=${this.infoModal.content.IngressAnotations[key]
            }`;
        }
      }
      return `/api/${api}?cmd=${cmd}&pod=${this.podsNamesSelected}&namespace=${this.infoModal.content.Namespace}${args}`;
    },
    async makeAPICall(api, cmd = "none", args = "" /* need & */) {
      try {
        switch (api) {
          case "scaleNamespace":
          case "deleteALL":
          case "disableHPA":
            break;
          default:
            if (!this.podsNamesSelected) {
              throw new Error("no pod selected");
            }
        }
        var realy = await this.$bvModal.msgBoxConfirm("Realy?");
        if (!realy) return;

        this.infoModal.loading = true;
        this.infoModal.info = false;

        const { result } = await this.$axios.$get(
          this.makeAPICallUrl(api, cmd, args)
        );

        if (result && result.ExecCode) {
          throw result;
        }

        if (api == "disableHPA") {
          await new Promise((resolve) => setTimeout(resolve, 10000));
          this.infoModal.loading = false;
          this.showTab(this.tabIndex, true, true);
        }

        this.infoModal.info = result;
      } catch (e) {
        console.error(e);
        if (e.response && e.response.data) {
          this.infoModal.info = e.response.data;
        } else {
          this.infoModal.info = e;
        }
      } finally {
        this.infoModal.loading = false;
      }
    },
    async enableGit() {
      if (!this.gitBranchState) return;
      if (!this.gitOrigin) return;

      this.makeAPICall(
        "exec",
        "enableGit",
        `&origin=${this.gitOrigin}&branch=${this.gitBranch}`
      );
    },
    async setPhpSettings() {
      this.makeAPICall(
        "exec",
        "setPhpSettings",
        `&text=${btoa(this.debug_text)}`
      );
    },
    async showTab(row, force, podsForce) {//NOSONAR
      if (this.infoModal.loading) return true;
      try {
        this.infoModal.loading = true;
        this.infoModal.info = false;

        //pods info
        const defaultPod =
          this.infoModal.content.IngressAnotations[
          "kubernetes-manager/default-pod"
          ];

        this.defaultPodInfo = ["app=default", "app"];
        var defaultPodLabelName = null;
        var defaultPodLabelValue = null;
        var defaultPodContainer = null;

        if (defaultPod && defaultPod.split(":").length == 2) {
          this.hasDefaultPod = true;
          this.defaultPodInfo = defaultPod.split(":");
          defaultPodLabelName = this.defaultPodInfo[0].split("=")[0];
          defaultPodLabelValue = this.defaultPodInfo[0].split("=")[1];
          defaultPodContainer = this.defaultPodInfo[1];
        }

        if (this.podsNames.length == 0 || podsForce) {
          this.podsNamesSelected = null;
          this.podsNamesSelectedTotal = 0;
          const { result } = await this.$axios.$get(
            this.makeAPICallUrl("getPods")
          );

          this.podsNames = [];

          this.podsNames.push({
            value: null,
            text: "Please select a POD",
          });

          if (result.ExecCode) {
            throw result;
          }

          result.forEach((pod) => {
            pod.PodContainers.forEach((container) => {
              if (defaultPod) {
                if (
                  pod.PodLabels[defaultPodLabelName] == defaultPodLabelValue
                ) {
                  if (container.ContainerName == defaultPodContainer) {
                    this.podsNamesSelectedTotal++;
                    if (!this.podsNamesSelected) {
                      this.podsNamesSelected = `${pod.PodName}:${container.ContainerName}`;
                    }
                  }
                }
              }
              this.podsNames.push({
                value: `${pod.PodName}:${container.ContainerName}`,
                text: `${pod.PodName}:${container.ContainerName}`,
              });
            });
          });
        }

        switch (row) {
          case 1:
            if (!force && this.tab1Data != null) return;

            const tab1Result = await this.$axios.$get(
              this.makeAPICallUrl("getServices")
            );

            if (tab1Result.result.ExecCode) {
              throw tab1Result.result;
            }

            this.tab1Data = tab1Result.result;
            break;
          case 2:
            if (force || this.tab2Data == null) {
              const tab2Result = await this.$axios.$get(
                this.makeAPICallUrl("exec", "mongoInfo")
              );
              if (tab2Result.result.ExecCode) {
                throw tab2Result.result;
              }

              this.tab2Data = JSON.parse(tab2Result.result.Stdout);
            }
            break;
          case 4:
            if (!this.podsNamesSelected) return;

            if (force || this.tab4Data == null) {
              this.debug_enabled = "unknown";

              const [obj1, obj2] = await Promise.all([
                this.$axios.$get(this.makeAPICallUrl("exec", "xdebugInfo")),
                this.$axios.$get(this.makeAPICallUrl("exec", "getPhpSettings")),
              ]);

              if (obj1.result.ExecCode) {
                throw obj1.result;
              }
              if (parseInt(obj1.result.Stdout) > 0) {
                this.debug_enabled = "enabled";
              } else {
                this.debug_enabled = "NOT enabled";
              }

              if (obj2.result.ExecCode) {
                throw obj2.result;
              }
              this.debug_text = obj2.result.Stdout;
              this.tab4Data = true;
            }
            break;
          case 5:
            if (!this.podsNamesSelected) return;

            if (force || this.tab5Data == null) {
              this.gitSyncEnabled = "unknown";
              this.gitSyncPublicKey = "";
              this.gitBranch = "";

              if (!this.gitOrigin) {
                this.gitOrigin =
                  this.infoModal.content.IngressAnotations[
                  "kubernetes-manager/git-project-origin"
                  ];
              }

              if (this.infoModal.content.IngressAnotations) {
                const b =
                  this.infoModal.content.IngressAnotations[
                  "kubernetes-manager/git-branch"
                  ];
                if (b) {
                  this.gitBranch = b;
                }
              }

              const [obj1, obj2] = await Promise.all([
                this.$axios.$get(this.makeAPICallUrl("exec", "getGitPubKey")),
                this.$axios.$get(this.makeAPICallUrl("exec", "getGitBranch")),
              ]);

              if (obj1.result.ExecCode) {
                throw obj1.result;
              }
              if (obj1.result.Stdout) {
                this.gitSyncEnabled = "enabled, fetch for changes";
                this.gitSyncPublicKey = obj1.result.Stdout;
              } else {
                this.gitSyncEnabled = "NOT enabled, need init";
              }
              if (obj2.result.ExecCode) {
                throw obj1.result;
              }
              if (obj2.result.Stdout) {
                this.gitBranch = obj2.result.Stdout;
              }
              this.tab5Data = true;
            }
            break;

          case 7:
            if (!force && this.tab7Data != null) return;

            const tab7Result = await this.$axios.$get(
              this.makeAPICallUrl("getProjects")
            );

            if (tab7Result.result.ExecCode) {
              throw tab7Result.result;
            }
            this.tab7DataSelected = false;
            this.tab7Data = tab7Result.result;

            this.tab7Data.forEach(async (el) => {
              const data = await this.$axios.$get(
                `/api/getProjectInfo?projectID=${el.ProjectID}&namespace=${this.infoModal.content.Namespace}`
              );
              if (data.result) {
                el.AdditionalInfo = data.result;
              }
            });

            break;
        }
      } catch (e) {
        console.error(e);
        if (e.response && e.response.data) {
          this.infoModal.info = e.response.data;
        } else {
          this.infoModal.info = e;
        }
      } finally {
        this.infoModal.loading = false;
      }
      return true;
    },
    info(item) {
      if (this.infoModal.content != item) {
        this.tab1Data = null;
        this.tab2Data = null;
        this.tab4Data = null;
        this.tab5Data = null;
        this.tab7Data = null;

        this.podsNamesSelectedTotal = 0;
        this.podsNamesSelected = null;
        this.podsNames = [];
        this.hasDefaultPod = false;
      }

      this.infoModal.title = item.NamespaceName;
      this.infoModal.content = item;

      for (let cluster of this.config.Clusters) {
        if (cluster.ClusterName == this.infoModal.content.Cluster) {
          this.infoModal.content.FrontConfig = cluster;
          break;
        }
      }

      this.$refs["info-modal"].show();
      this.showTab(this.tabIndex);
    },
    showAxiosError(e) {
      if (e.response && e.response.data) {
        this.error = e.response.data;
      } else {
        this.error = e;
      }
    },
    showProxyDialog(row) {
      const port = row.item.Ports.split(",")[0];
      var proxyType = "pod";
      if (/.+.svc.cluster.local$/.test(row.item.ServiceHost)) {
        proxyType = "svc";
      }
      const proxyString =
        `"Save As" this <a target="_blank" href="/getKubeConfig?cluster=${this.infoModal.content.Cluster}">file</a> to /tmp/kubeconfig-${this.infoModal.content.Cluster}` +
        `<br/><br/><textarea readonly style="background-color:#eeeeee;border:0px;padding:10px;outline:none;width:100%" onclick="this.focus();this.select()">kubectl --kubeconfig=/tmp/kubeconfig-${this.infoModal.content.Cluster} -n ${this.infoModal.content.NamespaceName} port-forward ${proxyType}/${row.item.Name} ${port}:${port}</textarea>` +
        `<br/><br/>this will listen localy 127.0.0.1:${port} and forward all requests to service in cluster, to listen other port for example 127.0.0.1:12345 - change end of command from ${port}:${port} to <strong>12345</strong>:${port}`;

      const h = this.$createElement;
      const messageVNode = h("div", { domProps: { innerHTML: proxyString } });

      this.$bvModal.msgBoxOk([messageVNode], {
        title: `Create proxy to remote service: ${proxyType}/${row.item.Name}`,
        size: "xl",
        centered: true,
      });
    },
    getProjectRefs(data) {
      const h = this.$createElement;
      const messageVNode = h("div", {
        domProps: {
          innerHTML:
            "Select branch: <select class='form-select' id='gitlabProjectBranchId' disabled><option value=false>Loading...</option></select>",
        },
      });

      this.$axios
        .$get(`/api/getProjectRefs?projectID=${data.item.ProjectID}`)
        .then((httpData) => {
          const select = document.getElementById("gitlabProjectBranchId");
          select.innerHTML = "";
          for (let row of httpData.result) {
            var opt = document.createElement("option");
            opt.value = row.Name;
            opt.innerHTML = row.Name;
            select.appendChild(opt);
          }
          select.disabled = false;
          select.value = data.item.DefaultBranch;
        });

      this.$bvModal
        .msgBoxConfirm([messageVNode], {
          title: data.item.Name,
          centered: true,
        })
        .then((value) => {
          if (!value) return false;

          const v = document.getElementById("gitlabProjectBranchId").value;
          if (!v) return false;

          for (let row of this.tab7Data) {
            if (data.item.ProjectID == row.ProjectID) {
              row.Deploy = v;
              this.tab7DataSelected = true;
              break;
            }
          }
        });
    },
    externalServiceTemplate(projectIDs) {
      const ids = projectIDs.split(";");

      for (let row of this.tab7Data) {
        if (ids.includes(row.ProjectID.toString())) {
          row.Deploy = row.DefaultBranch;
          this.tab7DataSelected = true;
        } else {
          row.Deploy = false;
        }
      }
    },
    filterTemplates(templates) {
      if (!templates) {
        return [];
      }
      return templates.filter((row) => {
        if (!row.NamespacePattern) {
          return true;
        }

        let filterResult = false;

        if (
          new RegExp(row.NamespacePattern).test(
            this.infoModal.content.NamespaceName
          )
        ) {
          filterResult = true;
        }

        return filterResult;
      });
    },
    deployButtonVariant(data) {
      if (data.item.Deploy) {
        return "success";
      }
      return "outline-secondary";
    },
    deploySelectedServices() {
      let selectedServices = [];
      for (let row of this.tab7Data) {
        if (row.Deploy) {
          selectedServices[
            selectedServices.length
          ] = `${row.ProjectID}:${row.Deploy}`;
        }
      }

      this.makeAPICall(
        "deploySelectedServices",
        "none",
        `&services=${encodeURIComponent(selectedServices.join(";"))}`
      );
    },
    getGitlabCommitURL(obj) {
      console.log(obj);
      if (obj && obj.AdditionalInfo && obj.AdditionalInfo.PodRunning.GitHash) {
        return `${obj.WebURL}/-/tree/${obj.AdditionalInfo.PodRunning.GitHash}`;
      }
    },
  },
};
</script>
