<template>
  <span v-if="$fetchState.pending" class="badge-margin badge rounded-pill bg-primary">Loading...</span>
  <span v-else-if="data && data.Result.BranchNotFound" class="hand badge-margin badge rounded-pill bg-danger">branch not
    found</span>
  <span v-else-if="data && data.Result.CommitsBehind > 0"
    @click="openInNewTab(`${data.Result.WebURL}/-/compare/${branch}...${data.Result.DefaultBranch}`)"
    class="hand badge-margin badge rounded-pill bg-danger">{{
      data.Result.CommitsBehind
    }} commits behind</span>
</template>
<script>
export default {
  props: ["id", "projectID", "branch"],
  watch: {
    id() {
      this.$fetch();
    },
  },
  data() {
    return {
      data: false,
    }
  },
  async fetch() {
    const result = await fetch(`/api/commits-behind?projectID=${encodeURI(this.projectID)}&branch=${encodeURI(this.branch)}`);
    if (result.ok) {
      this.data = await result.json();
    }
  },
  methods: {
    openInNewTab(url) {
      window.open(url, '_blank', 'noreferrer');
    },
  },
}
</script>