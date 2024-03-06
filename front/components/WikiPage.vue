<template>
  <b-alert v-if="$fetchState.error" variant="danger" show>{{
    $fetchState.error.message
  }}</b-alert>
  <b-card v-else bg-variant="light">
    <template v-if="!$fetchState.pending" #header>
      <b-button target="_blank" :href="data.EditURL" size="sm" variant="outline-secondary"
        class="float-right">Edit</b-button>&nbsp;
      {{ title ? title : data.Title }}
    </template>
    <b-card-text>
      <b-spinner v-if="$fetchState.pending" variant="primary" />
      <div v-else v-html="data.Content" />
    </b-card-text>
  </b-card>
</template>

<script>
export default {
  props: ["projectID", "slug", "title"],
  async fetch() {
    this.data = {}

    const result = await fetch(`/api/wiki-page?projectID=${this.projectID}&slug=${encodeURI(this.slug)}`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result
    } else {
      throw Error(await result.text());
    }
  },
}
</script>