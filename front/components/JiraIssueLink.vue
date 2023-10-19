<template>
  <div>
    <b-spinner v-if="$fetchState.pending" variant="primary" />
    <div v-else v-html="name" />
    <span v-if="status" class="badge rounded-pill bg-primary">{{ status.toUpperCase() }}</span>
  </div>
</template>
<script>
export default {
  props: ["item"],
  watch: {
    item() {
      this.$fetch();
    }
  },
  data() {
    return {
      name: "",
      status: "",
    }
  },
  async fetch() {
    const jira_matcher = /(\b[A-Z][A-Z0-9_]+-[1-9][0-9]*)/g;

    this.status = "";
    this.name = this.item;

    const jira_issue = this.name.match(jira_matcher, "$1")

    if (!jira_issue) {
      return
    }

    if (!this.config.Links.JiraURL) {
      return
    }

    this.name = this.item.replace(jira_matcher, `<a target='_blank' href='${this.config.Links.JiraURL}/browse/$1'>$1</a>`);

    // jira have CORS enabled, so we can't use fetch from browser
    const result = await fetch(`/api/jira-issue-info?issue=${encodeURI(jira_issue)}`);
    if (result.ok) {
      const data = await result.json();
      if (data.Result.fields.status.name) {
        this.status = data.Result.fields.status.name;
      }
    }
  }
}
</script>
