<template>
  <EnvironmentList scope="created" />
</template>
<script>
export default {
  head() {
    return {
      title: this.pageTitle('Created by me')
    }
  }
}
</script>
