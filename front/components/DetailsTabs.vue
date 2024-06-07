<template>
  <div>
    <div v-if="alertMessage" style="padding:10px;padding-bottom: 0px;">
      <b-alert show dismissible fade variant="warning">{{ alertMessage }}</b-alert>
    </div>
    <div style="padding:10px;display: flex;align-items: center;">
      <EnvironmentName />
      <div style="min-height: 50px;">
        <EnvironmentResources />
      </div>
    </div>
    <ul class="nav nav-tabs">
      <li class="nav-item" style="margin-left:20px">
        <b-link class="nav-link" active-class="active" to="info">info</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="services">services</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="events">events</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="issues">
          <IssuesBadges :environmentID="this.$route.params.environmentID" />
        </b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="settings">settings</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="debug">debug</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="git-sync">git-sync</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="kubectl">kubectl</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="autotests">autotests</b-link>
      </li>
      <li class="nav-item">
        <b-link class="nav-link" active-class="active" to="external-services">external-services</b-link>
      </li>
    </ul>
  </div>
</template>
<script>
export default {
  computed: {
    alertMessage() {
      return this.namespaceAnnotation(this.const().AnnotationAlert);
    },
  },
}
</script>
