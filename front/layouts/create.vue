<template>
  <div>
    <NavBar />
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
        $fetchState.error.message
    }}</b-alert>
    <b-alert v-if="this.errorText" variant="danger" show>{{
        this.errorText
    }}</b-alert>
    <nuxt keep-alive />
  </div>
</template>
<script>
export default {
  async fetch() {
    this.loadConfig();
    this.loadUser("/oauth2/userinfo");
  },
}
</script>
