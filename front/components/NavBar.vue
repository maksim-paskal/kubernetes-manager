<template>
  <b-navbar sticky fixed="top" toggleable="lg" type="dark" variant="dark">
    <b-navbar-nav>
      <b-input-group style="padding-left: 10px; padding-right: 10px">
        <b-button variant="outline-secondary" to="/">Dashboard</b-button>&nbsp;
        <b-button variant="outline-secondary" to="/create"><svg xmlns="http://www.w3.org/2000/svg" width="20"
            height="20" fill="currentColor" class="bi bi-plus-lg" viewBox="0 0 16 16">
            <path fill-rule="evenodd"
              d="M8 2a.5.5 0 0 1 .5.5v5h5a.5.5 0 0 1 0 1h-5v5a.5.5 0 0 1-1 0v-5h-5a.5.5 0 0 1 0-1h5v-5A.5.5 0 0 1 8 2Z">
            </path>
          </svg>&nbsp;Create</b-button>&nbsp;
        <b-button variant="outline-secondary" @click="refresh()"><svg xmlns="http://www.w3.org/2000/svg" width="20"
            height="20" fill="currentColor" class="bi bi-arrow-repeat" viewBox="0 0 16 16">
            <path
              d="M11.534 7h3.932a.25.25 0 0 1 .192.41l-1.966 2.36a.25.25 0 0 1-.384 0l-1.966-2.36a.25.25 0 0 1 .192-.41zm-11 2h3.932a.25.25 0 0 0 .192-.41L2.692 6.23a.25.25 0 0 0-.384 0L.342 8.59A.25.25 0 0 0 .534 9z">
            </path>
            <path fill-rule="evenodd"
              d="M8 3c-1.552 0-2.94.707-3.857 1.818a.5.5 0 1 1-.771-.636A6.002 6.002 0 0 1 13.917 7H12.9A5.002 5.002 0 0 0 8 3zM3.1 9a5.002 5.002 0 0 0 8.757 2.182.5.5 0 1 1 .771.636A6.002 6.002 0 0 1 2.083 9H3.1z">
            </path>
          </svg>&nbsp;Refresh</b-button>&nbsp;
        <b-dropdown variant="outline-secondary" text="Menu">
          <div v-if="this.config.Links">
            <b-dropdown-item v-for="(item, index) in this.config ? this.config.Links.Others : []" :key="index"
              target="_blank" :href="item.URL">
              <em style="font-size: 24px" class="bi bi-box-arrow-up-right" />&nbsp;{{ item.Name }}
            </b-dropdown-item>
            <b-dropdown-item target="_blank" v-if="this.config && this.config.Links.SlackURL"
              :href="this.config.Links.SlackURL">
              <em style="font-size: 24px" class="bi bi-slack" />&nbsp;Report Issue
            </b-dropdown-item>
            <b-dropdown-divider></b-dropdown-divider>
          </div>
          <b-dropdown-item v-if="this.config" target="_blank"
            href="https://github.com/maksim-paskal/kubernetes-manager">
            <em style="font-size: 24px" class="bi bi-github" />&nbsp;{{
                this.config.Version
            }}
          </b-dropdown-item>
        </b-dropdown>
      </b-input-group>
    </b-navbar-nav>

    <b-navbar-nav>
      <Search />&nbsp;
      <b-button v-if="this.user.user" variant="outline-secondary">{{ this.user.user }}</b-button>
    </b-navbar-nav>
  </b-navbar>
</template>
<script>
export default {
  methods: {
    refresh() {
      if (this.$route.params.environmentID) {
        this.loadEnvironment(this.$route.params.environmentID)
      }
      this.$router.app.refresh();
    }
  }
}
</script>