<template>
  <b-button variant="link" v-if="!this.environment.ID || this.loading || !this.user.user"><em class="bi bi-hourglass" />
  </b-button>
  <b-alert variant="danger" v-else-if="this.errorText" show>{{ this.errorText }}</b-alert>
  <b-button variant="link" :title="likedUsers.join(', ')" style="text-decoration: none" v-else @click="save()"><em
      :class="this.fillClass">{{ likedUsers.length ? `&nbsp;${likedUsers.length}` : "" }}</em></b-button>
</template>
<script>
export default {
  mounted() {
    this.errorText = "";
  },
  data() {
    return {
      loading: false,
      data: {}
    }
  },
  methods: {
    async save() {
      try {
        this.loading = true;

        await this.call("make-save-user-like", { User: this.user.user })
        if (!this.errorText) {
          this.loadEnvironment(this.$route.params.environmentID)
        }
      } catch (e) {
        this.errorText = e.message
      } finally {
        this.loading = false;
      }
    }
  },
  computed: {
    userLabel() {
      return this.const({ user: this.user.user }).LabelLiked
    },
    hasLabel() {
      return this.namespaceLabel(this.userLabel) == "true" ? true : false
    },
    likedUsers() {
      let users = [];
      Object.keys(this.environment.NamespaceLabels).forEach(key => {
        if (this.environment.NamespaceLabels[key] == "true" && key.startsWith(this.const().LabelLikedPrefix)) {
          users.push(key.replace(this.const().LabelLikedPrefix, ""));
        }
      })
      return users
    },
    fillClass() {
      if (this.hasLabel) {
        return "bi bi-star-fill"
      } else {
        return "bi bi-star"
      }
    }
  }
}
</script>