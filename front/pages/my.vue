<template>
  <EnvironmentList v-if="this.user.user" :filter="this.filter" />
  <b-spinner v-else variant="primary" />
</template>
<script>
export default {
  head() {
    return {
      title: this.pageTitle('Created by me')
    }
  },
  computed: {
    userLabel() {
      return this.const({ user: this.user.user }).LabelCreator
    },
    filter() {
      return `${this.userLabel}=true`
    }
  }
}
</script>
