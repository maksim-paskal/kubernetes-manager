<template>
  <div>
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>
    <b-spinner v-if="callIsLoading" variant="primary" />
    <b-alert v-else variant="warning" show>
      <b-button @click="call('make-disable-mtls', null, false)">Disable mTLS verification</b-button>&nbsp;For proper
      usage you must disable mutual TLS verification
    </b-alert>
  </div>
</template>
<script>
export default {}
</script>