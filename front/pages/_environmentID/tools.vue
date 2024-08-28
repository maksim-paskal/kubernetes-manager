<template>
  <div class="detail-tab">
    <b-card no-body>
      <b-tabs pills card>
        <b-tab title="Kubectl" active><Kubectl/></b-tab>
        <b-tab title="Local proxy">Coming soon...</b-tab>
        <b-tab title="Debug"><Debug/></b-tab>
        <b-tab title="Git Sync"><GitSync/></b-tab>
      </b-tabs>
    </b-card>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Tools', true)
    }
  }
};
</script>