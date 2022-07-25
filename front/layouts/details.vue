<template>
  <div>
    <NavBar />
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-alert v-if="this.errorText" variant="danger" show>{{
        this.errorText
    }}</b-alert>
    <div v-if="!environment.ID" style="padding:20px">
      <b-spinner variant="primary" />
    </div>
    <div v-else>
      <DetailsTabs />
      <nuxt keep-alive :key="this.$route.params.environmentID" />
    </div>
  </div>
</template>
<script>
export default {
  mounted() {
    this.$store.commit("setEnvironment", {});
    this.$store.commit("clearDropDowns");
  },
  async fetch() {
    this.loadConfig();
    this.loadUser("/oauth2/userinfo");
    this.loadEnvironment(this.$route.params.environmentID)
  },
}
</script>
