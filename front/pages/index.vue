<template>
  <EnvironmentList v-if="this.user.user" :filter="this.filter" />
  <b-spinner v-else variant="primary" />
</template>
<script>
export default {
  head() {
    return {
      title: this.pageTitle('Favorites')
    }
  },
  computed: {
    userLabel() {
      return this.const({ user: this.user.user }).LabelLiked
    },
    filter() {
      return `${this.userLabel}=true`
    }
  }
}
</script>
