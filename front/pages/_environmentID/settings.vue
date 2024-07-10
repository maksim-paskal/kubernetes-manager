<template>
  <div class="detail-tab">
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>

    <b-spinner v-if="callIsLoading" variant="primary" />
    <div v-else>
      <b-card bg-variant="light" header="Pause branch" class="text-center">
        <ScaleDownDate />
        <br />
        <b-button @click="call('make-pause')"><em class="bi bi-pause-fill" />&nbsp;Pause
        </b-button>
        <b-button variant="success" @click="call('make-start')"><em class="bi bi-play-fill" />&nbsp;Start</b-button>
        <b-button @click="scaledownDelay('3h')">Delay autopause for next 3 hours</b-button>
      </b-card>
      <br />
      <b-card bg-variant="light" header="Actions" class="text-center">
        <b-button @click="call('make-disable-hpa')">Disable autoscaling</b-button>
        <b-button @click="call('make-restore-hpa')">Restore autoscaling</b-button>
        <b-button @click="call('make-disable-mtls')">Disable mTLS verification</b-button>
        <b-button @click="call('make-snapshot')">Make environment snapshot</b-button>
      </b-card>
      <br />
      <b-card bg-variant="light" header="Danger Zone" class="text-center">
        <b-button variant="danger" @click="call('make-delete')"><em class="bi bi-x-octagon-fill" />&nbsp;Delete All
        </b-button>
      </b-card>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Settings', true)
    }
  },
  methods: {
    async scaledownDelay(interval) {
      await this.call('make-scaledown-delay', { Delay: interval });
      this.loadEnvironment(this.environment.ID);
    }
  }
};
</script>