<template>
  <EnvironmentList scope="liked" />
</template>
<script>
export default {
  head() {
    return {
      title: this.pageTitle('Favorites')
    }
  }
}
</script>
