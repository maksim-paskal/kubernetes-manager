<template>
  <b-container>
    <b-row>
      <b-col v-for="(item, index) in this.config.WikiPages" :key="index" style="margin-top:20px" :cols="item.Size">
        <WikiPage :title="item.Title" :projectID="item.ProjectID" :slug="item.Slug" />
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
export default {
  head() {
    return {
      title: this.pageTitle('Wiki')
    }
  },
}
</script>
