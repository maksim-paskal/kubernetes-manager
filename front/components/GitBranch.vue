<template>
  <ul style="list-style: none;padding-left:0px">
    <li v-bind:key="index" v-for="(gitBranch, index) in gitBranches">
      <JiraIssueLink :item="gitBranch" />
    </li>
  </ul>
</template>
<script>
export default {
  props: ["item"],
  computed: {
    gitBranches() {
      if (!this.item.NamespaceAnnotations) return;
      let branches = [];
      Object.keys(this.item.NamespaceAnnotations).forEach(item => {
        if (item.startsWith("kubernetes-manager/project-")) {
          const name = this.item.NamespaceAnnotations[item];
          if (!branches.includes(name)) {
            branches.push(name);
          }
        }
      })

      // sort branches alphabetically case insensitive
      branches = branches.sort(function (a, b) {
        return a.toLowerCase().localeCompare(b.toLowerCase());
      });

      return branches;
    }
  }
}
</script>
