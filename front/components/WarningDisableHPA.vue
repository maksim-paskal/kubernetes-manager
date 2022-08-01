<template>
  <div>
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-alert v-if="infoText" variant="info" show>{{ infoText }}</b-alert>
    <b-spinner v-if="callIsLoading" variant="primary" />
    <b-alert v-else variant="warning" show>
      <b-button @click="call('make-disable-hpa', null, false)">Disable autoscaling</b-button>&nbsp;For proper usage you
      must disable autoscaler
    </b-alert>
  </div>
</template>
<script>
export default {}
</script>