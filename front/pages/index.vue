<template>
  <div>
    <b-navbar sticky fixed="top" toggleable="lg" type="dark" variant="dark">
      <b-input-group>
        <b-button :disabled="isBusy" variant="outline-secondary" @click="getIngress()">Refresh</b-button>&nbsp;
        <b-form-input
          v-model="filter"
          :disabled="isBusy"
          autocomplete="off"
          placeholder="Type to Search"
        ></b-form-input>
      </b-input-group>
    </b-navbar>

    <div style="padding:50px" class="text-center" v-if="isBusy">
      <b-spinner style="width: 10rem; height: 10rem;" variant="primary"></b-spinner>
    </div>

    <b-modal :id="infoModal.id" :title="infoModal.title" size="xl" ok-only>
      <div style="padding:50px" class="text-center" v-if="infoModal.loading">
        <b-spinner style="width: 10rem; height: 10rem;" variant="primary"></b-spinner>
      </div>

      <b-alert
        v-if="infoModal.info && !infoModal.info.Stdout && !infoModal.info.Stderr"
        variant="danger"
        show
      >
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
            <b-card-text>
              <pre>{{ infoModal.content }}</pre>
            </b-card-text>
          </b-tab>
          <b-tab key="tab1" title="mysql" :disabled="!isMysqlTab">
            <pre>{{ tab1Data }}</pre>
            <div v-if="tab1Data != null">
              <b-button @click="showTab(1,true,true)">Refresh</b-button>
              <b-button
                v-if="tab1Data.result == 'found'"
                target="_blank"
                v-bind:href="tab1Data.phpmyadminURL"
              >Open Phpmyadmin</b-button>
              <b-button
                v-if="tab1Data.result == 'found'"
                variant="outline-primary"
                @click="makeAPICall('exec','mysqlMigrations')"
              >Migrations</b-button>
            </div>
          </b-tab>
          <b-tab key="tab2" title="mongo" :disabled="!isMongoTab">
            <pre>{{ tab2Data }}</pre>
            <div v-if="tab2Data != null">
              <b-button @click="showTab(2,true,true)">Refresh</b-button>
              <b-button
                v-if="tab2Data.result == 'found'"
                variant="danger"
                @click="makeAPICall('exec','mongoDelete')"
              >Delete</b-button>
              <b-button
                variant="outline-primary"
                @click="makeAPICall('exec','mongoMigrations')"
              >Migrations</b-button>
            </div>
          </b-tab>
          <b-tab key="tab3" title="settings">
            <b-card bg-variant="light" header="Pause branch" class="text-center">
              <div>branch autopause will be at 16:00 UTC</div>
              <br />
              <b-button @click="makeAPICall('scaleNamespace','none','&replicas=0')">Pause</b-button>
              <b-button
                variant="success"
                @click="makeAPICall('scaleNamespace','none','&replicas=1')"
              >Start</b-button>
            </b-card>
            <br />
            <b-card bg-variant="light" header="Autoscaling" class="text-center">
              <b-button @click="makeAPICall('disableHPA','none')">Disable autoscaling</b-button>
            </b-card>
            <br />
            <b-card bg-variant="light" header="Danger Zone" class="text-center">
              <b-button variant="danger" @click="makeAPICall('deleteALL','projectID=')">Delete All</b-button>
            </b-card>
          </b-tab>
          <b-tab key="tab4" title="debug">
            <b-form-select
              class="mb-3"
              v-model="podsNamesSelected"
              v-on:change="podsNamesChange"
              :options="podsNames"
            />
            <div v-if="podsNamesSelected">
              <b-card-text>
                XDEBUG is
                <strong>{{ this.debug_enabled }}</strong>
              </b-card-text>
              <b-card-text>extra setting in php-fpm.conf</b-card-text>

              <b-container class="bv-example-row">
                <b-row>
                  <b-col cols="9">
                    <b-form-textarea
                      spellcheck="false"
                      rows="10"
                      v-model="debug_text"
                      placeholder="Enter something..."
                    />
                  </b-col>
                  <b-col cols="2">
                    <b-dropdown id="dropdown-1" offset="25" text="Templates" class="m-2">
                      <b-dropdown-item @click="templateAction(0)">config xdebug</b-dropdown-item>
                      <b-dropdown-item @click="templateAction(1)">disable opcache</b-dropdown-item>
                      <b-dropdown-item @click="templateAction(2)">enable debug mode</b-dropdown-item>
                    </b-dropdown>
                  </b-col>
                </b-row>
              </b-container>

              <b-card-text>
                Use
                <strong>ngrok tcp 9000</strong>
              </b-card-text>
              <b-button @click="showTab(4,true,true)">Refresh</b-button>
              <b-button variant="success" @click="makeAPICall('exec','xdebugEnable')">Enable XDEBUG</b-button>
              <b-button @click="setPhpSettings()">Set settings in php-fpm.conf</b-button>
              <b-button variant="danger" @click="makeAPICall('deletePod','none')">Delete Pod</b-button>
            </div>
          </b-tab>
          <b-tab key="tab5" title="git-sync">
            <b-form-select
              class="mb-3"
              v-model="podsNamesSelected"
              v-on:change="podsNamesChange"
              :options="podsNames"
            />
            <div v-if="podsNamesSelected">
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
              <b-button @click="showTab(5,true,true)">Refresh</b-button>
              <b-button variant="success" @click="enableGit()">Init</b-button>
              <b-button @click="showPublicKey()">Show public key</b-button>
              <b-button variant="danger" @click="makeAPICall('deletePod','none')">Delete Pod</b-button>
              <b-button @click="makeAPICall('exec','gitFetch')">Fetch</b-button>
              <b-button @click="makeAPICall('exec','clearCache')">Clear Cache</b-button>
            </div>
          </b-tab>

          <b-tab key="tab6" title="kubectl">
            <b-card no-body>
              <b-tabs pills card vertical>
                <b-tab title="1. Install" active>
                  <b-card-text>
                    Install requred application
                    <a
                      target="_blank"
                      href="https://kubernetes.io/docs/tasks/tools/install-kubectl/"
                    >https://kubernetes.io/docs/tasks/tools/install-kubectl/</a>
                  </b-card-text>
                </b-tab>
                <b-tab title="2. Configure">
                  <b-card-text>
                    "Save As" this
                    <a target="_blank" href="/getKubeConfig">file</a> to /tmp/kubeconfig
                  </b-card-text>
                </b-tab>
                <b-tab title="3. Test">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} get pods</b-card-text>
                </b-tab>
                <b-tab title="4. Shell">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} exec -it `kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} get pods -l=app=paket -o jsonpath='{.items[0].metadata.name}'` -c backend sh</b-card-text>
                </b-tab>
                <b-tab title="5. Logs">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} logs `kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} get pods -l=app=paket -o jsonpath='{.items[0].metadata.name}'` -c backend</b-card-text>
                </b-tab>
                <b-tab title="6. Clear memcahed">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} delete `kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} get pods -l=app=memcached -o name`</b-card-text>
                </b-tab>
                <b-tab title="7. Restart react">
                  <b-card-text>kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} delete `kubectl --kubeconfig=/tmp/kubeconfig -n {{ infoModal.content.Namespace }} get pods -l=app=paket-site -o name`</b-card-text>
                </b-tab>
              </b-tabs>
            </b-card>
          </b-tab>
        </b-tabs>
      </b-card>
    </b-modal>

    <div style="padding:5px" v-if="!isBusy && items == null">
      <b-alert variant="warning" show>No available namespaces founded</b-alert>
    </div>

    <div style="padding:5px" v-if="error != null">
      <b-alert variant="danger" show>{{ error }}</b-alert>
    </div>

    <b-table
      striped
      hover
      :fields="fields"
      :items="items"
      v-if="!isBusy && items != null"
      :filter="filter"
    >
      <template v-slot:cell(Status)="data">
        <b-spinner v-if="data.item.RunningPodsCount<0" variant="primary"></b-spinner>
        <div v-else>
          <div v-if="data.item.RunningPodsCount==0">
            Paused
            <br />
            <b-button
              variant="success"
              @click="unpauseNamespace(data.item.Namespace,data.item.IngressAnotations['kubernetes-manager/version'])"
            >Start</b-button>
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
        <b-button
          size="sm"
          variant="outline-primary"
          @click="info(row.item, row.index, $event.target)"
        >Details</b-button>
      </template>
    </b-table>
  </div>
</template>

<script>
export default {
  mounted() {
    this.getIngress();
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
      fields: [
        { key: "Actions", sortable: false },
        { key: "Status", sortable: false },
        { key: "GitBranch", sortable: false },
        { key: "Hosts", sortable: false },
      ],
      isBusy: true,
      error: null,
      items: [],
      filter: null,
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
      tab2Data: null,
      tab4Data: null,
      tab5Data: null,
      debug_enabled: "unknown",
      debug_text: "",
      gitOrigin: "",
      gitBranch: "",
      gitSyncEnabled: "",
      gitSyncShowPublicKey: false,
      gitSyncPublicKey: "",

      podsNames: [],
      podsNamesSelected: null,

      isMysqlTab: false,
      isMongoTab: false,
    };
  },
  methods: {
    async podsNamesChange() {
      this.showTab(this.tabIndex, true);
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
        const { result } = await this.$axios.$get(url);
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
        const { result } = await this.$axios.$get("/api/getIngress");
        this.items = result;

        this.items.forEach(async (el) => {
          const result = await this.$axios.$get(
            "/api/getRunningPodsCount?namespace=" + el.Namespace
          );
          el.RunningPodsCount = result.count;
        });
      } catch (e) {
        console.error(e);
        this.showAxiosError(e);
      }
      this.isBusy = false;
    },
    async templateAction(id) {
      switch (id) {
        case 0:
          this.debug_text +=
            "env[XDEBUG_CONFIG]='remote_host=0.tcp.ngrok.io remote_port=17570'\nenv[PHP_IDE_CONFIG]='serverName=dev'";//TODO: dynamic value
          break;
        case 1:
          this.debug_text += "php_value[opcache.enable]=0";
          break;
        case 2:
          this.debug_text += "env[APP_ENV]='dev'";
          break;
      }
      this.debug_text += "\n";
    },
    async showPublicKey() {
      this.gitSyncShowPublicKey = !this.gitSyncShowPublicKey;
    },
    makeAPICallUrl(api, cmd = "none", args = "" /* need & */) {
      for (var key in this.infoModal.content.IngressAnotations) {
        if (key.startsWith("kubernetes-manager")) {
          args += `&${key.substring(19)}=${
            this.infoModal.content.IngressAnotations[key]
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
            break;
          default:
            if (!this.podsNamesSelected) {
              throw "no pod selected";
            }
        }
        var realy = await this.$bvModal.msgBoxConfirm("Realy?");
        if (!realy) return;

        //this.infoModal.error = false;
        this.infoModal.loading = true;
        this.infoModal.info = false;

        const { result } = await this.$axios.$get(
          this.makeAPICallUrl(api, cmd, args)
        );

        if (result && result.ExecCode) {
          throw result;
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
    async showTab(row, force, podsForce) {
      if (this.infoModal.loading) return true;
      try {
        //this.infoModal.error = false;
        this.infoModal.loading = true;
        this.infoModal.info = false;

        //pods info
        const defaultPod = this.infoModal.content.IngressAnotations[
          "kubernetes-manager/default-pod"
        ];

        var defaultPodInfo = null;
        var defaultPodLabelName = null;
        var defaultPodLabelValue = null;
        var defaultPodContainer = null;

        if (defaultPod) {
          defaultPodInfo = defaultPod.split(":");
          defaultPodLabelName = defaultPodInfo[0].split("=")[0];
          defaultPodLabelValue = defaultPodInfo[0].split("=")[1];
          defaultPodContainer = defaultPodInfo[1];
        }

        if (this.podsNames.length == 0 || podsForce) {
          const { result } = await this.$axios.$get(
            this.makeAPICallUrl("getPods")
          );

          this.podsNames = [];
          this.isMysqlTab = false;
          this.isMongoTab = false;

          this.podsNames.push({
            value: null,
            text: "Please select a POD",
          });

          if (result.ExecCode) {
            throw result;
          }

          result.forEach((pod) => {
            switch (pod.PodLabels["app"]) {
              case "mysql":
                this.isMysqlTab = true;
                break;
              case "mongo":
                this.isMongoTab = true;
                break;
            }

            pod.PodContainers.forEach((container) => {
              if (defaultPod && !this.podsNamesSelected) {
                if (
                  pod.PodLabels[defaultPodLabelName] == defaultPodLabelValue
                ) {
                  if (container.ContainerName == defaultPodContainer) {
                    this.podsNamesSelected = `${pod.PodName}:${container.ContainerName}`;
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
            if (force || this.tab1Data == null) {
              this.tab1Data = {
                result: "found",
                server: `mysql.${this.infoModal.content.Namespace}.svc.cluster.local`,
                phpmyadminURL:
                  /* TODO: Dynamic value */
                  '{{ env "FRONT_PHPMYADMIN_URL" }}',
              };
            }
            break;
          case 2:
            if (force || this.tab2Data == null) {
              const { result } = await this.$axios.$get(
                this.makeAPICallUrl("exec", "mongoInfo")
              );
              if (result.ExecCode) {
                throw result;
              }

              this.tab2Data = JSON.parse(result.Stdout);
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
                this.gitOrigin = this.infoModal.content.IngressAnotations[
                  "kubernetes-manager/git-project-origin"
                ];
              }

              if (this.infoModal.content.IngressAnotations) {
                const b = this.infoModal.content.IngressAnotations[
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
    info(item, index, button) {
      if (this.infoModal.content != item) {
        this.tab1Data = null;
        this.tab2Data = null;
        this.tab4Data = null;
        this.tab5Data = null;

        this.podsNamesSelected = null;
        this.podsNames = [];
      }

      this.infoModal.title = item.Namespace;
      this.infoModal.content = item; //JSON.stringify(item, null, 2);
      this.$root.$emit("bv::show::modal", this.infoModal.id, button);

      this.showTab(this.tabIndex);
    },
    showAxiosError(e) {
      if (e.response && e.response.data) {
        this.error = e.response.data;
      } else {
        this.error = e;
      }
    },
  },
};
</script>
